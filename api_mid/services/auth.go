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

// ============================================================
// SERVICE DE USUARIOS (crud_auth)
// Aqui esta la logica: validar, cifrar, bloquear, crear codigos.
// ============================================================

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

// asignarRolUsuario le da el rol "Usuario" a un usuario nuevo.
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

// buscarCredencial trae la clave cifrada de un usuario (si tiene).
func buscarCredencial(idUsuario int) (models.Credencial, bool, error) {

	var credenciales []models.Credencial

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/credencial?query=id_usuario:%d", helpers.UrlCrudAuth, idUsuario), &credenciales); err != nil {
		return models.Credencial{}, false, err
	}

	if len(credenciales) == 0 {
		return models.Credencial{}, false, nil
	}

	return credenciales[0], true, nil
}

// crearCodigo crea un codigo de 6 numeros, lo guarda cifrado y lo "envia".
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

// validarCodigo revisa el ultimo codigo activo del usuario.
// Si es correcto lo desactiva (no sirve dos veces); si no, suma un intento.
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

// guardarContrasena guarda una clave nueva cifrada (crea la credencial si no existe).
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

// ------------------------------------------------------------
// Funciones que usan los controladores
// ------------------------------------------------------------

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
		"cedula":           usuario.Cedula,
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
		Cedula:                  datos.Cedula,
		Telefono:                datos.Telefono,
		Ciudad:                  datos.Ciudad,
		Direccion:               datos.Direccion,
		AceptaTerminos:          true,
		FechaAceptacionTerminos: time.Now().UTC(),
		Estado:                  "activo",
		Activo:                  true,
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

// IniciarSesion: paso 1 del login. Valida correo y clave y crea el codigo.
// Bloquea la cuenta 15 minutos tras 5 intentos fallidos.
func IniciarSesion(datos models.DatosLogin) (map[string]interface{}, error) {

	errorLogin := helpers.NuevoError(http.StatusUnauthorized, "Correo o contrasena incorrectos")

	// 1. Buscar el usuario (mismo mensaje si no existe o si la clave esta mal)
	usuario, existe, err := buscarUsuarioPorEmail(strings.ToLower(strings.TrimSpace(datos.Email)))

	if err != nil {
		return nil, err
	}

	if !existe {
		return nil, errorLogin
	}

	if !usuario.Activo || usuario.Estado != "activo" {
		return nil, helpers.NuevoError(http.StatusForbidden, "La cuenta esta inactiva")
	}

	// 2. Buscar su credencial y revisar si esta bloqueada
	credencial, tiene, err := buscarCredencial(usuario.Id)

	if err != nil {
		return nil, err
	}

	if !tiene {
		return nil, errorLogin
	}

	ahora := time.Now().UTC()

	if credencial.BloqueadoHasta.After(ahora) {
		return nil, helpers.NuevoError(http.StatusLocked, "Cuenta bloqueada por intentos fallidos. Intenta de nuevo en 15 minutos.")
	}

	// 3. Comparar la clave y actualizar los intentos fallidos
	claveCorrecta := helpers.CoincideCifrado(credencial.ContrasenaHash, datos.Contrasena)

	if claveCorrecta {
		credencial.IntentosFallidos = 0
	} else {
		credencial.IntentosFallidos++
		if credencial.IntentosFallidos >= 5 {
			credencial.IntentosFallidos = 0
			credencial.BloqueadoHasta = ahora.Add(15 * time.Minute)
		}
	}

	helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)

	// 4. Guardar el intento en la auditoria
	estadoEvento := "exitoso"
	if !claveCorrecta {
		estadoEvento = "fallido"
	}

	auditoria := models.AuditoriaLogin{
		IdUsuario:    &models.Ref{Id: usuario.Id},
		TipoEvento:   "login",
		IpAddress:    datos.Ip,
		Navegador:    datos.Navegador,
		EstadoEvento: estadoEvento,
	}

	helpers.Crear(helpers.UrlCrudAuth+"/auditoria_login", auditoria, nil)

	if !claveCorrecta {
		return nil, errorLogin
	}

	// 5. Crear el codigo de 6 numeros (vence en 10 minutos)
	if err := crearCodigo(usuario.Id, usuario.Email, "login_2fa", 10); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"mensaje":         "Te enviamos un codigo de verificacion",
		"requiere_codigo": true,
		"id_usuario":      usuario.Id,
	}, nil
}

// VerificarCodigo: paso 2 del login. Valida el codigo de 6 numeros.
func VerificarCodigo(datos models.DatosCodigo) (map[string]interface{}, error) {

	if err := validarCodigo(datos.IdUsuario, "login_2fa", datos.Codigo); err != nil {
		return nil, err
	}

	// Guardar la fecha de la sesion
	usuario, err := BuscarUsuario(datos.IdUsuario)

	if err != nil {
		return nil, err
	}

	usuario.FechaUltimaSesion = time.Now().UTC()
	helpers.Actualizar(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, usuario.Id), usuario)

	return map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"id_usuario": usuario.Id,
		"nombre":     usuario.Nombre,
		"apellido":   usuario.Apellido,
		"email":      usuario.Email,
	}, nil
}

// IniciarSesionSocial: entrar con Google, Apple o Facebook.
// La pagina habla con esos servicios y aqui llegan sus datos.
func IniciarSesionSocial(datos models.DatosLoginSocial) (map[string]interface{}, error) {

	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))

	// 1. Validar
	if datos.Proveedor != "google" && datos.Proveedor != "apple" && datos.Proveedor != "facebook" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El proveedor debe ser: google, apple o facebook")
	}

	if datos.IdExterno == "" || datos.Email == "" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Faltan el id_externo o el email")
	}

	// 2. Buscar si la cuenta social ya existe
	var cuentas []models.CuentaSocial

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/cuenta_social?query=proveedor:"+datos.Proveedor+",id_externo:"+url.QueryEscape(datos.IdExterno), &cuentas); err != nil {
		return nil, err
	}

	if len(cuentas) > 0 {
		return respuestaSocial(cuentas[0].IdUsuario.Id, datos), nil
	}

	// 3. Si no existe, buscar el usuario por correo; si tampoco existe, crearlo
	usuario, existe, err := buscarUsuarioPorEmail(datos.Email)

	if err != nil {
		return nil, err
	}

	if !existe {

		usuario = models.Usuario{Nombre: datos.Nombre, Apellido: datos.Apellido, Email: datos.Email, Estado: "activo", Activo: true}

		if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario", usuario, &usuario); err != nil {
			return nil, err
		}

		asignarRolUsuario(usuario.Id)
	}

	// 4. Vincular la cuenta social al usuario
	cuenta := models.CuentaSocial{
		IdUsuario:      &models.Ref{Id: usuario.Id},
		Proveedor:      datos.Proveedor,
		IdExterno:      datos.IdExterno,
		EmailProveedor: datos.Email,
		Activo:         true,
	}

	helpers.Crear(helpers.UrlCrudAuth+"/cuenta_social", cuenta, nil)

	return respuestaSocial(usuario.Id, datos), nil
}

func respuestaSocial(idUsuario int, datos models.DatosLoginSocial) map[string]interface{} {
	return map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"id_usuario": idUsuario,
		"email":      datos.Email,
		"proveedor":  datos.Proveedor,
	}
}

// PedirCodigoRecuperacion: crea un codigo para cambiar la clave olvidada.
// Siempre responde lo mismo, para no revelar que correos existen.
func PedirCodigoRecuperacion(email string) (map[string]interface{}, error) {

	respuesta := map[string]interface{}{"mensaje": "Si el correo esta registrado, te enviamos un codigo"}

	email = strings.ToLower(strings.TrimSpace(email))

	usuario, existe, err := buscarUsuarioPorEmail(email)

	if err != nil {
		return nil, err
	}

	if existe {
		crearCodigo(usuario.Id, email, "recuperar_contrasena", 15)
	}

	return respuesta, nil
}

// RestablecerContrasena: valida el codigo y guarda la clave nueva.
func RestablecerContrasena(datos models.DatosContrasena) (map[string]interface{}, error) {

	if len(datos.ContrasenaNueva) < 8 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener minimo 8 caracteres")
	}

	usuario, existe, err := buscarUsuarioPorEmail(strings.ToLower(strings.TrimSpace(datos.Email)))

	if err != nil {
		return nil, err
	}

	if !existe {
		return nil, helpers.NuevoError(http.StatusUnauthorized, "El codigo no es correcto")
	}

	if err := validarCodigo(usuario.Id, "recuperar_contrasena", datos.Codigo); err != nil {
		return nil, err
	}

	if err := guardarContrasena(usuario.Id, datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Contrasena actualizada"}, nil
}

// CambiarContrasena: desde el perfil, pidiendo la clave actual.
func CambiarContrasena(idUsuario int, datos models.DatosContrasena) (map[string]interface{}, error) {

	if len(datos.ContrasenaNueva) < 8 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "La contrasena nueva debe tener minimo 8 caracteres")
	}

	credencial, tiene, err := buscarCredencial(idUsuario)

	if err != nil {
		return nil, err
	}

	if !tiene {
		return nil, helpers.NuevoError(http.StatusNotFound, "El usuario no tiene contrasena")
	}

	if !helpers.CoincideCifrado(credencial.ContrasenaHash, datos.ContrasenaActual) {
		return nil, helpers.NuevoError(http.StatusUnauthorized, "La contrasena actual no es correcta")
	}

	if err := guardarContrasena(idUsuario, datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Contrasena actualizada"}, nil
}
