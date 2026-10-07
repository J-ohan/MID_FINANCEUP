package services

import (
	"fmt"
	"strings"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"

	"golang.org/x/crypto/bcrypt"
)

const (
	intentosMaximos  = 5                // intentos fallidos antes de bloquear la cuenta
	tiempoBloqueo    = 15 * time.Minute // tiempo que la cuenta queda bloqueada
	rolPorDefecto    = "Usuario"        // rol que recibe quien se registra
	largoMinimoClave = 8
)

// ObtenerPerfil junta el usuario, su tipo de documento y sus roles.
func ObtenerPerfil(idUsuario int) (models.Perfil, error) {
	var usuario models.Usuario
	if err := helpers.Consultar("auth", fmt.Sprintf("usuario/%d", idUsuario), &usuario); err != nil {
		return models.Perfil{}, mensajeNoEncontrado(err, "El usuario no existe")
	}

	perfil := models.Perfil{
		IdUsuario:     usuario.Id,
		Nombre:        usuario.Nombre,
		Apellido:      usuario.Apellido,
		Email:         usuario.Email,
		Cedula:        usuario.Cedula,
		Telefono:      usuario.Telefono,
		Ciudad:        usuario.Ciudad,
		Direccion:     usuario.Direccion,
		Estado:        usuario.Estado,
		FechaRegistro: usuario.FechaRegistro,
		Roles:         []string{},
	}
	if !usuario.FechaNacimiento.IsZero() {
		perfil.FechaNacimiento = usuario.FechaNacimiento.Format("2006-01-02")
	}
	if !usuario.FechaUltimaSesion.IsZero() {
		perfil.UltimaSesion = &usuario.FechaUltimaSesion
	}

	if usuario.TipoDocumento != nil {
		var tipo models.TipoDocumento
		if err := helpers.Consultar("auth", fmt.Sprintf("tipo_documento/%d", usuario.TipoDocumento.Id), &tipo); err == nil {
			perfil.TipoDocumento = tipo.Codigo
		}
	}

	var asignaciones []models.UsuarioRol
	if err := helpers.Consultar("auth", "usuario_rol?"+helpers.Filtro("id_usuario", idUsuario, "activo", true), &asignaciones); err != nil {
		return models.Perfil{}, err
	}
	for _, asignacion := range asignaciones {
		var rol models.Rol
		if err := helpers.Consultar("auth", fmt.Sprintf("rol/%d", asignacion.IdRol.Id), &rol); err == nil {
			perfil.Roles = append(perfil.Roles, rol.NombreRol)
		}
	}
	return perfil, nil
}

// ExisteUsuario confirma que el usuario existe y esta activo.
// Lo usan los demas modulos antes de guardar algo a nombre de un usuario.
func ExisteUsuario(idUsuario int) error {
	var usuario models.Usuario
	if err := helpers.Consultar("auth", fmt.Sprintf("usuario/%d", idUsuario), &usuario); err != nil {
		return mensajeNoEncontrado(err, "El usuario no existe")
	}
	if !usuario.Activo {
		return helpers.NuevoError(403, "El usuario esta inactivo")
	}
	return nil
}

// Registrar crea un usuario nuevo con su clave y el rol "Usuario".
// Son tres registros en el CRUD de auth; si uno falla se borran los anteriores
// para no dejar un usuario a medias.
func Registrar(datos models.DatosRegistro) (models.Perfil, error) {
	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))
	if err := validarRegistro(datos); err != nil {
		return models.Perfil{}, err
	}

	// 1. El correo y la cedula no pueden estar repetidos.
	var repetidos []models.Usuario
	if err := helpers.Consultar("auth", "usuario?"+helpers.Filtro("email", datos.Email), &repetidos); err != nil {
		return models.Perfil{}, err
	}
	if len(repetidos) > 0 {
		return models.Perfil{}, helpers.NuevoError(409, "Ya existe un usuario con ese correo")
	}
	if datos.Cedula != "" {
		if err := helpers.Consultar("auth", "usuario?"+helpers.Filtro("cedula", datos.Cedula), &repetidos); err != nil {
			return models.Perfil{}, err
		}
		if len(repetidos) > 0 {
			return models.Perfil{}, helpers.NuevoError(409, "Ya existe un usuario con ese numero de documento")
		}
	}

	// 2. Buscar el rol por defecto antes de crear nada.
	var roles []models.Rol
	if err := helpers.Consultar("auth", "rol?"+helpers.Filtro("nombre_rol", rolPorDefecto), &roles); err != nil {
		return models.Perfil{}, err
	}
	if len(roles) == 0 {
		return models.Perfil{}, helpers.NuevoError(500, "No existe el rol '"+rolPorDefecto+"' en la base de datos")
	}

	// 3. Cifrar la clave. Nunca se guarda la clave en texto plano.
	hash, err := bcrypt.GenerateFromPassword([]byte(datos.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		return models.Perfil{}, helpers.NuevoError(500, "No fue posible proteger la contrasena")
	}

	// 4. Crear el usuario.
	usuario := models.Usuario{
		Nombre:    strings.TrimSpace(datos.Nombre),
		Apellido:  strings.TrimSpace(datos.Apellido),
		Email:     datos.Email,
		Cedula:    datos.Cedula,
		Telefono:  datos.Telefono,
		Ciudad:    datos.Ciudad,
		Direccion: datos.Direccion,
		Estado:    "activo",
		Activo:    true,
	}
	if datos.IdTipoDocumento > 0 {
		usuario.TipoDocumento = &models.Ref{Id: datos.IdTipoDocumento}
	}
	if datos.FechaNacimiento != "" {
		usuario.FechaNacimiento, _ = time.Parse("2006-01-02", datos.FechaNacimiento)
	}
	if err := helpers.Enviar("POST", "auth", "usuario", usuario, &usuario); err != nil {
		return models.Perfil{}, err
	}

	// 5. Guardar la clave cifrada. El "salt" de bcrypt va dentro del hash.
	credencial := models.Credencial{
		IdUsuario:      &models.Ref{Id: usuario.Id},
		ContrasenaHash: string(hash),
		Salt:           string(hash[7:29]),
		Algoritmo:      "bcrypt",
		Activo:         true,
	}
	if err := helpers.Enviar("POST", "auth", "credencial", credencial, &credencial); err != nil {
		deshacer("usuario", usuario.Id)
		return models.Perfil{}, err
	}

	// 6. Asignar el rol.
	asignacion := models.UsuarioRol{
		IdUsuario: &models.Ref{Id: usuario.Id},
		IdRol:     &models.Ref{Id: roles[0].Id},
		Activo:    true,
	}
	if err := helpers.Enviar("POST", "auth", "usuario_rol", asignacion, nil); err != nil {
		deshacer("credencial", credencial.Id)
		deshacer("usuario", usuario.Id)
		return models.Perfil{}, err
	}

	return ObtenerPerfil(usuario.Id)
}

// IniciarSesion valida correo y clave. Bloquea la cuenta 15 minutos despues
// de 5 intentos fallidos y deja todo registrado en auditoria_login.
func IniciarSesion(datos models.DatosLogin, ip, navegador string) (models.Perfil, error) {
	errorCredenciales := helpers.NuevoError(401, "Correo o contrasena incorrectos")
	email := strings.ToLower(strings.TrimSpace(datos.Email))
	if email == "" || datos.Contrasena == "" {
		return models.Perfil{}, helpers.NuevoError(400, "El correo y la contrasena son obligatorios")
	}

	var usuarios []models.Usuario
	if err := helpers.Consultar("auth", "usuario?"+helpers.Filtro("email", email), &usuarios); err != nil {
		return models.Perfil{}, err
	}
	if len(usuarios) == 0 {
		return models.Perfil{}, errorCredenciales
	}
	usuario := usuarios[0]
	if !usuario.Activo || usuario.Estado != "activo" {
		return models.Perfil{}, helpers.NuevoError(403, "La cuenta esta "+usuario.Estado+". Comunicate con soporte.")
	}

	var credenciales []models.Credencial
	if err := helpers.Consultar("auth", "credencial?"+helpers.Filtro("id_usuario", usuario.Id), &credenciales); err != nil {
		return models.Perfil{}, err
	}
	if len(credenciales) == 0 {
		return models.Perfil{}, errorCredenciales
	}
	credencial := credenciales[0]

	ahora := time.Now().UTC()
	if credencial.BloqueadoHasta.After(ahora) {
		minutos := int(credencial.BloqueadoHasta.Sub(ahora).Minutes()) + 1
		return models.Perfil{}, helpers.NuevoError(423, fmt.Sprintf("Cuenta bloqueada por intentos fallidos. Intenta de nuevo en %d minutos.", minutos))
	}

	auditoria := models.AuditoriaLogin{
		IdUsuario:  &models.Ref{Id: usuario.Id},
		TipoEvento: "login",
		IpAddress:  ip,
		Navegador:  recortar(navegador, 200),
	}

	if bcrypt.CompareHashAndPassword([]byte(credencial.ContrasenaHash), []byte(datos.Contrasena)) != nil {
		credencial.IntentosFallidos++
		if credencial.IntentosFallidos >= intentosMaximos {
			credencial.IntentosFallidos = 0
			credencial.BloqueadoHasta = ahora.Add(tiempoBloqueo)
		}
		helpers.Enviar("PUT", "auth", fmt.Sprintf("credencial/%d", credencial.Id), credencial, nil)

		auditoria.EstadoEvento = "fallido"
		helpers.Enviar("POST", "auth", "auditoria_login", auditoria, nil)
		return models.Perfil{}, errorCredenciales
	}

	// Clave correcta: se reinician los intentos y se guarda la fecha de la sesion.
	if credencial.IntentosFallidos > 0 {
		credencial.IntentosFallidos = 0
		helpers.Enviar("PUT", "auth", fmt.Sprintf("credencial/%d", credencial.Id), credencial, nil)
	}
	usuario.FechaUltimaSesion = ahora
	helpers.Enviar("PUT", "auth", fmt.Sprintf("usuario/%d", usuario.Id), usuario, nil)

	auditoria.EstadoEvento = "exitoso"
	helpers.Enviar("POST", "auth", "auditoria_login", auditoria, nil)

	return ObtenerPerfil(usuario.Id)
}

func validarRegistro(datos models.DatosRegistro) error {
	switch {
	case strings.TrimSpace(datos.Nombre) == "":
		return helpers.NuevoError(400, "El nombre es obligatorio")
	case strings.TrimSpace(datos.Apellido) == "":
		return helpers.NuevoError(400, "El apellido es obligatorio")
	case !strings.Contains(datos.Email, "@") || !strings.Contains(datos.Email, "."):
		return helpers.NuevoError(400, "El correo no es valido")
	case len(datos.Contrasena) < largoMinimoClave:
		return helpers.NuevoError(400, fmt.Sprintf("La contrasena debe tener minimo %d caracteres", largoMinimoClave))
	}
	if datos.FechaNacimiento != "" {
		if _, err := time.Parse("2006-01-02", datos.FechaNacimiento); err != nil {
			return helpers.NuevoError(400, "La fecha de nacimiento debe tener el formato AAAA-MM-DD")
		}
	}
	return nil
}

// deshacer borra un registro de auth cuando un paso posterior del registro fallo.
func deshacer(recurso string, id int) {
	helpers.Enviar("DELETE", "auth", fmt.Sprintf("%s/%d", recurso, id), nil, nil)
}

func recortar(texto string, largo int) string {
	if len(texto) > largo {
		return texto[:largo]
	}
	return texto
}
