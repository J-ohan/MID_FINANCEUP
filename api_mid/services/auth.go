package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// BuscarUsuario trae un usuario por su id (lo usan otros modulos).
func BuscarUsuario(id int) (models.Usuario, error) {
	var usuario models.Usuario
	err := helpers.ConsultarUno(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, id), &usuario, "El usuario no existe")
	return usuario, err
}

// buscarUsuarioPorEmail devuelve el usuario y si existe o no.
func buscarUsuarioPorEmail(email string) (models.Usuario, bool, error) {

	var usuarios []models.Usuario

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?query=email:"+url.QueryEscape(email), &usuarios); err != nil {
		return models.Usuario{}, false, err
	}

	if len(usuarios) == 0 {
		return models.Usuario{}, false, nil
	}

	return usuarios[0], true, nil
}

// NombresUsuarios devuelve un mapa id -> "Nombre Apellido" de todos los usuarios.
func NombresUsuarios() (map[int]string, error) {

	var usuarios []models.Usuario

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?limit=10000", &usuarios); err != nil {
		return nil, err
	}

	nombres := map[int]string{}
	for _, usuario := range usuarios {
		nombres[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	return nombres, nil
}

// asignarRolUsuario le da el rol "Usuario" a un usuario nuevo

func asignarRolUsuario(idUsuario int) error {

	var roles []models.Rol

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/rol?query=nombre_rol:Usuario", &roles); err != nil {
		return err
	}

	if len(roles) == 0 {
		return helpers.NuevoError(http.StatusInternalServerError, "No existe el rol Usuario en la base de datos")
	}

	asignacion := models.UsuarioRol{IdUsuario: &models.Ref{Id: idUsuario}, IdRol: &models.Ref{Id: roles[0].Id}, Activo: true}

	return helpers.Crear(helpers.UrlCrudAuth+"/usuario_rol", asignacion, nil)
}

// Por ahora el envio es mostrarlo en la consola del MID.
func crearCodigo(idUsuario int, email string, proposito string, minutos int) error {

	codigo := helpers.NumeroAleatorio()

	nuevoCodigo := models.CodigoVerificacion{
		IdUsuario:       &models.Ref{Id: idUsuario},
		Proposito:       proposito,
		Canal:           "email",
		CodigoHash:      helpers.Cifrar(codigo),
		FechaExpiracion: time.Now().UTC().Add(time.Duration(minutos) * time.Minute),
		Activo:          true,
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/codigo_verificacion", nuevoCodigo, nil); err != nil {
		return err
	}
	
// Cuando haya servicio de correo, aqui se envia al email
	fmt.Println("[CODIGO "+proposito+"]", email, codigo)

	return nil
}


// Si es correcto lo desactiva (no sirve dos veces); si no, suma un intento

func validarCodigo(idUsuario int, proposito string, codigoEscrito string) error {

	var codigos []models.CodigoVerificacion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/codigo_verificacion?query=id_usuario:%d,proposito:%s,activo:true&sortby=Id&order=desc&limit=1", helpers.UrlCrudAuth, idUsuario, proposito), &codigos); err != nil {
		return err
	}

	if len(codigos) == 0 {
		return helpers.NuevoError(http.StatusUnauthorized, "No hay un codigo pendiente. Pide uno nuevo.")
	}

	codigo := codigos[0]
	ahora := time.Now().UTC()

	if codigo.FechaExpiracion.Before(ahora) || codigo.Intentos >= 5 {
		return helpers.NuevoError(http.StatusUnauthorized, "El codigo vencio. Pide uno nuevo.")
	}

	correcto := helpers.CoincideCifrado(codigo.CodigoHash, codigoEscrito)

	if correcto {
		codigo.Activo = false
		codigo.FechaUso = ahora
	} else {
		codigo.Intentos++
	}

	helpers.Actualizar(fmt.Sprintf("%s/codigo_verificacion/%d", helpers.UrlCrudAuth, codigo.Id), codigo)

	if !correcto {
		return helpers.NuevoError(http.StatusUnauthorized, "El codigo no es correcto")
	}

	return nil
}


// guardarContrasena guarda una clave nueva cifrada (crea la credencial si no existe)

func guardarContrasena(idUsuario int, contrasena string) error {

	credencial, existe, err := buscarCredencial(idUsuario)

	if err != nil {
		return err
	}

	hash := helpers.Cifrar(contrasena)

	credencial.ContrasenaHash = hash
	credencial.Salt = hash[7:29]
	credencial.Algoritmo = "bcrypt"
	credencial.Activo = true
	credencial.FechaUltimoCambio = time.Now().UTC()
	credencial.IntentosFallidos = 0

	if !existe {
		credencial.IdUsuario = &models.Ref{Id: idUsuario}
		return helpers.Crear(helpers.UrlCrudAuth+"/credencial", credencial, nil)
	}

	return helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)
}


// Funciones que usan los controladores


// ObtenerPerfil: usuario + tipo de documento + roles

func ObtenerPerfil(id int) (map[string]interface{}, error) {

	// 1. El usuario
	usuario, err := BuscarUsuario(id)

	if err != nil {
		return nil, err
	}

	// 2. Su tipo de documento
	tipoDocumento := ""

	if usuario.TipoDocumento != nil {
		var tipo models.TipoDocumento
		if helpers.ConsultarUno(fmt.Sprintf("%s/tipo_documento/%d", helpers.UrlCrudAuth, usuario.TipoDocumento.Id), &tipo, "") == nil {
			tipoDocumento = tipo.Codigo
		}
	}

	// 3. Sus roles

	var asignaciones []models.UsuarioRol

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudAuth, id), &asignaciones); err != nil {
		return nil, err
	}

	roles := []string{}

	for _, asignacion := range asignaciones {
		var rol models.Rol
		if helpers.ConsultarUno(fmt.Sprintf("%s/rol/%d", helpers.UrlCrudAuth, asignacion.IdRol.Id), &rol, "") == nil {
			roles = append(roles, rol.NombreRol)
		}
	}

	// 4. Armar el perfil
	
	cedula := ""
	if usuario.Cedula != nil {
		cedula = *usuario.Cedula
	}

	fechaNacimiento := ""
	if !usuario.FechaNacimiento.IsZero() {
		fechaNacimiento = usuario.FechaNacimiento.Format("2006-01-02")
	}

	return map[string]interface{}{
		"id_usuario":       usuario.Id,
		"nombre":           usuario.Nombre,
		"apellido":         usuario.Apellido,
		"email":            usuario.Email,
		"tipo_documento":   tipoDocumento,
		"cedula":           cedula,
		"telefono":         usuario.Telefono,
		"ciudad":           usuario.Ciudad,
		"direccion":        usuario.Direccion,
		"fecha_nacimiento": fechaNacimiento,
		"estado":           usuario.Estado,
		"roles":            roles,
	}, nil
}

// Registrar: crea usuario + clave cifrada + rol "Usuario"

func Registrar(datos models.DatosRegistro) (map[string]interface{}, error) {

	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))

	// 1. Validar
	if datos.Nombre == "" || datos.Apellido == "" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El nombre y el apellido son obligatorios")
	}

	if !strings.Contains(datos.Email, "@") {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El correo no es valido")
	}

	if len(datos.Contrasena) < 8 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener minimo 8 caracteres")
	}

	if !datos.AceptaTerminos {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Debes aceptar los terminos y condiciones")
	}

	// 2. Revisar que el correo no este registrado

	_, existe, err := buscarUsuarioPorEmail(datos.Email)

	if err != nil {
		return nil, err
	}

	if existe {
		return nil, helpers.NuevoError(http.StatusConflict, "Ya existe un usuario con ese correo")
	}

	// 3. Crear el usuario

	nuevoUsuario := models.Usuario{
		Nombre:                  datos.Nombre,
		Apellido:                datos.Apellido,
		Email:                   datos.Email,
		Telefono:                datos.Telefono,
		Ciudad:                  datos.Ciudad,
		Direccion:               datos.Direccion,
		AceptaTerminos:          true,
		FechaAceptacionTerminos: time.Now().UTC(),
		Estado:                  "activo",
		Activo:                  true,
	}

	// La cedula es opcional: si no llega se deja en nil para que se guarde NULL

	if datos.Cedula != "" {
		nuevoUsuario.Cedula = &datos.Cedula
	}

	if datos.IdTipoDocumento > 0 {
		nuevoUsuario.TipoDocumento = &models.Ref{Id: datos.IdTipoDocumento}
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario", nuevoUsuario, &nuevoUsuario); err != nil {
		return nil, err
	}

	// 4. Guardar la clave cifrada y el rol

	if err := guardarContrasena(nuevoUsuario.Id, datos.Contrasena); err != nil {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El usuario se creo pero no se pudo guardar la contrasena")
	}

	if err := asignarRolUsuario(nuevoUsuario.Id); err != nil {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El usuario se creo pero no se pudo asignar el rol")
	}

	return map[string]interface{}{
		"mensaje":    "Usuario registrado",
		"id_usuario": nuevoUsuario.Id,
		"nombre":     nuevoUsuario.Nombre,
		"apellido":   nuevoUsuario.Apellido,
		"email":      nuevoUsuario.Email,
		"rol":        "Usuario",
	}, nil
}


