package services

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// ADMINISTRAR USUARIOS (solo admin: el filtro ya lo revisa)
// Lista con rol, cambiar rol y cambiar estado con motivo.
// Editar los datos de un usuario va directo al CRUD (una sola tabla).
// ============================================================

const minimoLetrasMotivo = 5

// Estados que acepta la tabla auth.usuario
var estadosUsuario = map[string]bool{"activo": true, "inactivo": true, "suspendido": true}

// La pagina usa "admin" y "user"; en la base se llaman asi
var nombresDeRol = map[string]string{"admin": "Administrador", "user": "Usuario"}

// ListarUsuariosAdmin devuelve todos los usuarios con su rol.
// Son solo 3 consultas (usuarios, asignaciones y roles), no una por usuario.
func ListarUsuariosAdmin() ([]map[string]interface{}, error) {

	var usuarios []models.Usuario
	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?limit=10000&sortby=Id&order=desc", &usuarios); err != nil {
		return nil, err
	}

	var asignaciones []models.UsuarioRol
	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario_rol?query=activo:true&limit=10000", &asignaciones); err != nil {
		return nil, err
	}

	roles, err := todosLosRoles()
	if err != nil {
		return nil, err
	}

	// Nombre del rol por su id
	nombreRol := map[int]string{}
	for _, rol := range roles {
		nombreRol[rol.Id] = rol.NombreRol
	}

	// Roles de cada usuario
	rolesPorUsuario := map[int][]string{}
	for _, asignacion := range asignaciones {
		if asignacion.IdUsuario != nil && asignacion.IdRol != nil {
			id := asignacion.IdUsuario.Id
			rolesPorUsuario[id] = append(rolesPorUsuario[id], nombreRol[asignacion.IdRol.Id])
		}
	}

	lista := []map[string]interface{}{}

	for _, usuario := range usuarios {

		rolesUsuario := rolesPorUsuario[usuario.Id]
		if rolesUsuario == nil {
			rolesUsuario = []string{}
		}

		cedula := ""
		if usuario.Cedula != nil {
			cedula = *usuario.Cedula
		}

		var ultimaSesion interface{} // null si nunca ha entrado
		if !usuario.FechaUltimaSesion.IsZero() {
			ultimaSesion = usuario.FechaUltimaSesion
		}

		lista = append(lista, map[string]interface{}{
			"id_usuario":     usuario.Id,
			"nombre":         usuario.Nombre,
			"apellido":       usuario.Apellido,
			"email":          usuario.Email,
			"telefono":       usuario.Telefono,
			"cedula":         cedula,
			"estado":         usuario.Estado,
			"activo":         usuario.Activo,
			"rol":            rolPrincipal(rolesUsuario),
			"roles":          rolesUsuario,
			"fecha_registro": usuario.FechaRegistro,
			"ultima_sesion":  ultimaSesion,
		})
	}

	return lista, nil
}

// CambiarRolUsuario le deja al usuario un solo rol activo: el que eligio el admin.
func CambiarRolUsuario(idAdmin int, idUsuario int, datos models.DatosAdminUsuario) (map[string]interface{}, error) {

	if idUsuario == idAdmin {
		return nil, helpers.NuevoError(http.StatusBadRequest, "No puedes cambiar tu propio rol")
	}

	nombre := strings.TrimSpace(datos.Rol)
	if equivalente, existe := nombresDeRol[strings.ToLower(nombre)]; existe {
		nombre = equivalente
	}

	if nombre == "" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El rol es obligatorio")
	}

	if _, err := BuscarUsuario(idUsuario); err != nil {
		return nil, err
	}

	// 1. El rol debe existir en la base
	roles, err := todosLosRoles()
	if err != nil {
		return nil, err
	}

	var rolNuevo *models.Rol
	for i := range roles {
		if strings.EqualFold(roles[i].NombreRol, nombre) {
			rolNuevo = &roles[i]
		}
	}

	if rolNuevo == nil {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Ese rol no existe")
	}

	// 2. Asignaciones que ya tiene (activas e inactivas)
	var asignaciones []models.UsuarioRol
	if err := helpers.ConsultarLista(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d&limit=100", helpers.UrlCrudAuth, idUsuario), &asignaciones); err != nil {
		return nil, err
	}

	// 3. Primero se activa el rol nuevo (si falla algo, no queda sin rol)
	yaLoTenia := false
	for _, asignacion := range asignaciones {
		if asignacion.IdRol != nil && asignacion.IdRol.Id == rolNuevo.Id {
			yaLoTenia = true
			if !asignacion.Activo {
				asignacion.Activo = true
				if err := helpers.Actualizar(fmt.Sprintf("%s/usuario_rol/%d", helpers.UrlCrudAuth, asignacion.Id), asignacion); err != nil {
					return nil, err
				}
			}
		}
	}

	if !yaLoTenia {
		nueva := models.UsuarioRol{IdUsuario: &models.Ref{Id: idUsuario}, IdRol: &models.Ref{Id: rolNuevo.Id}, Activo: true}
		if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario_rol", nueva, nil); err != nil {
			return nil, err
		}
	}

	// 4. Despues se apagan los demas
	for _, asignacion := range asignaciones {
		if asignacion.Activo && (asignacion.IdRol == nil || asignacion.IdRol.Id != rolNuevo.Id) {
			asignacion.Activo = false
			if err := helpers.Actualizar(fmt.Sprintf("%s/usuario_rol/%d", helpers.UrlCrudAuth, asignacion.Id), asignacion); err != nil {
				return nil, err
			}
		}
	}

	// 5. Su token tiene el rol viejo: debe entrar de nuevo
	helpers.CerrarSesionesDeUsuario(idUsuario)

	RegistrarActividad(idUsuario, "CAMBIO_ROL",
		fmt.Sprintf("Rol cambiado a %s por el administrador con id %d", rolNuevo.NombreRol, idAdmin), "auth.usuario_rol")

	return map[string]interface{}{
		"mensaje":    "Rol actualizado. El usuario debe iniciar sesion de nuevo.",
		"id_usuario": idUsuario,
		"rol":        rolPrincipal([]string{rolNuevo.NombreRol}),
		"roles":      []string{rolNuevo.NombreRol},
	}, nil
}

// CambiarEstadoUsuario activa, suspende o inactiva una cuenta.
// Para suspender o inactivar se pide el motivo (queda en el historial).
func CambiarEstadoUsuario(idAdmin int, idUsuario int, datos models.DatosAdminUsuario) (map[string]interface{}, error) {

	if idUsuario == idAdmin {
		return nil, helpers.NuevoError(http.StatusBadRequest, "No puedes cambiar el estado de tu propia cuenta")
	}

	estado := strings.ToLower(strings.TrimSpace(datos.Estado))
	if !estadosUsuario[estado] {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El estado debe ser activo, inactivo o suspendido")
	}

	motivo := strings.TrimSpace(datos.Motivo)
	if estado != "activo" && utf8.RuneCountInString(motivo) < minimoLetrasMotivo {
		return nil, helpers.NuevoError(http.StatusBadRequest, fmt.Sprintf("Escribe el motivo (minimo %d caracteres)", minimoLetrasMotivo))
	}

	usuario, err := BuscarUsuario(idUsuario)
	if err != nil {
		return nil, err
	}

	if usuario.Estado == estado {
		return nil, helpers.NuevoError(http.StatusConflict, "El usuario ya esta "+estado)
	}

	anterior := usuario.Estado
	usuario.Estado = estado

	// El CRUD reemplaza el registro: se manda el usuario completo
	if err := helpers.Actualizar(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, idUsuario), usuario); err != nil {
		return nil, err
	}

	// Si ya no puede entrar, se le cierran las sesiones abiertas
	if estado != "activo" {
		helpers.CerrarSesionesDeUsuario(idUsuario)
	}

	descripcion := fmt.Sprintf("Estado cambiado de %s a %s por el administrador con id %d", anterior, estado, idAdmin)
	if motivo != "" {
		descripcion += ". Motivo: " + motivo
	}
	RegistrarActividad(idUsuario, "CAMBIO_ESTADO", descripcion, "auth.usuario")

	return map[string]interface{}{"mensaje": "Estado actualizado", "id_usuario": idUsuario, "estado": estado}, nil
}

// todosLosRoles trae la tabla de roles (son pocos).
func todosLosRoles() ([]models.Rol, error) {
	var roles []models.Rol
	err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/rol?limit=100", &roles)
	return roles, err
}
