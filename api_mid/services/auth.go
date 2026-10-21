   package services

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE USUARIOS (crud_auth)
// Registro, login con codigo, Google, contrasenas y sesion.
// ============================================================

// Tiempos y limites
const (
	minutosCodigoLogin      = 10 // vigencia del codigo del login
	minutosCodigoRecuperar  = 15 // vigencia del codigo para recuperar la clave
	segundosEntreCodigos    = 60 // maximo un codigo por minuto
	intentosAntesDeBloquear = 5
	minutosDeBloqueo        = 15
)

// ------------------------------------------------------------
// Apoyo: usuarios, roles y validaciones
// ------------------------------------------------------------

// BuscarUsuario trae un usuario por su id (lo usan otros modulos).
func BuscarUsuario(id int) (models.Usuario, error) {
	var usuario models.Usuario
	err := helpers.ConsultarUno(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, id), &usuario, "El usuario no existe")
	return usuario, err
}

// buscarUsuarioPorEmail devuelve el usuario y si existe.
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

// NombresUsuarios devuelve id -> "Nombre Apellido" de todos (lo usan otros modulos).
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

// normalizarEmail deja el correo en minuscula y revisa que sea valido.
func normalizarEmail(email string) (string, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	invalido := helpers.NuevoError(http.StatusBadRequest, "El correo no es valido")

	// Comas y dos puntos se rechazan porque rompen el ?query= del CRUD
	direccion, err := mail.ParseAddress(email)
	if err != nil || direccion.Address != email || strings.ContainsAny(email, ",:") {
		return "", invalido
	}

	dominio := email[strings.LastIndex(email, "@")+1:]
	if !strings.Contains(dominio, ".") || utf8.RuneCountInString(email) > 150 {
		return "", invalido
	}

	return email, nil
}

// validarContrasena: minimo 8, una mayuscula, un numero y maximo 72 (como la pagina).
func validarContrasena(clave string) error {

	if utf8.RuneCountInString(clave) < 8 {
		return helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener minimo 8 caracteres")
	}

	if len(clave) > helpers.LargoMaximoClave {
		return helpers.NuevoError(http.StatusBadRequest, fmt.Sprintf("La contrasena no puede tener mas de %d caracteres", helpers.LargoMaximoClave))
	}

	tieneMayuscula, tieneNumero := false, false
	for _, letra := range clave {
		tieneMayuscula = tieneMayuscula || unicode.IsUpper(letra)
		tieneNumero = tieneNumero || unicode.IsDigit(letra)
	}

	if !tieneMayuscula || !tieneNumero {
		return helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener al menos una mayuscula y un numero")
	}

	return nil
}

// revisarEstado dice si la cuenta puede entrar (activo, inactivo o suspendido).
func revisarEstado(usuario models.Usuario) error {

	if !usuario.Activo {
		return helpers.NuevoError(http.StatusForbidden, "Esta cuenta fue desactivada. Comunicate con soporte.")
	}

	switch usuario.Estado {
	case "activo":
		return nil
	case "suspendido":
		return helpers.NuevoError(http.StatusForbidden, "Tu cuenta esta suspendida temporalmente. Comunicate con soporte para mas informacion.")
	case "inactivo":
		return helpers.NuevoError(http.StatusForbidden, "Tu cuenta esta inactiva. Comunicate con soporte para activarla.")
	}

	return helpers.NuevoError(http.StatusForbidden, "Tu cuenta no puede iniciar sesion. Comunicate con soporte.")
}

// asignarRolUsuario le da el rol "Usuario" a una cuenta nueva.
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

// rolesDelUsuario devuelve los nombres de sus roles (Administrador, Usuario...).
func rolesDelUsuario(idUsuario int) ([]string, error) {

	var asignaciones []models.UsuarioRol

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudAuth, idUsuario), &asignaciones); err != nil {
		return nil, err
	}

	roles := []string{}

	for _, asignacion := range asignaciones {

		if asignacion.IdRol == nil {
			continue
		}

		var rol models.Rol
		if helpers.ConsultarUno(fmt.Sprintf("%s/rol/%d", helpers.UrlCrudAuth, asignacion.IdRol.Id), &rol, "") == nil {
			roles = append(roles, rol.NombreRol)
		}
	}

	return roles, nil
}

// rolPrincipal: "admin" si es Administrador, si no "user" (lo que usa la pagina).
func rolPrincipal(roles []string) string {
	for _, rol := range roles {
		if strings.EqualFold(rol, "Administrador") {
			return "admin"
		}
	}
	return "user"
}

// buscarCredencial trae la clave cifrada del usuario (si tiene).
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

// auditar guarda un evento en auditoria_login (si falla, no detiene nada).
func auditar(idUsuario int, tipoEvento string, exitoso bool, ip string, navegador string) {

	estado := "exitoso"
	if !exitoso {
		estado = "fallido"
	}

	helpers.Crear(helpers.UrlCrudAuth+"/auditoria_login", models.AuditoriaLogin{
		IdUsuario:    &models.Ref{Id: idUsuario},
		TipoEvento:   tipoEvento,
		IpAddress:    ip,
		Navegador:    navegador,
		EstadoEvento: estado,
	}, nil)
}

// ------------------------------------------------------------
// Codigos del login y de recuperar contrasena (tabla codigo_verificacion)
// ------------------------------------------------------------

// ultimoCodigo trae el codigo mas reciente (soloActivos ignora los ya usados).
func ultimoCodigo(idUsuario int, proposito string, soloActivos bool) (models.CodigoVerificacion, bool, error) {

	filtro := fmt.Sprintf("id_usuario:%d,proposito:%s", idUsuario, proposito)
	if soloActivos {
		filtro += ",activo:true"
	}

	var codigos []models.CodigoVerificacion

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/codigo_verificacion?query="+filtro+"&sortby=Id&order=desc&limit=1", &codigos); err != nil {
		return models.CodigoVerificacion{}, false, err
	}

	if len(codigos) == 0 {
		return models.CodigoVerificacion{}, false, nil
	}

	return codigos[0], true, nil
}

// crearCodigo genera un codigo de 6 numeros, lo guarda cifrado y lo envia por Gmail.
func crearCodigo(usuario models.Usuario, proposito string, minutos int) error {

	ahora := time.Now().UTC()

	// 1. Maximo un codigo por minuto
	if anterior, existe, err := ultimoCodigo(usuario.Id, proposito, false); err != nil {
		return err
	} else if existe && !anterior.FechaCreacion.IsZero() && ahora.Sub(anterior.FechaCreacion) < segundosEntreCodigos*time.Second {
		return helpers.NuevoError(http.StatusTooManyRequests, "Ya te enviamos un codigo hace menos de un minuto. Revisa tu correo o espera un momento.")
	}

	// 2. Crear y guardar cifrado
	codigo := helpers.NumeroAleatorio()

	hash, err := helpers.Cifrar(codigo)
	if err != nil {
		return err
	}

	nuevoCodigo := models.CodigoVerificacion{
		IdUsuario:       &models.Ref{Id: usuario.Id},
		Proposito:       proposito,
		Canal:           "email",
		CodigoHash:      hash,
		FechaExpiracion: ahora.Add(time.Duration(minutos) * time.Minute),
		Activo:          true,
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/codigo_verificacion", nuevoCodigo, &nuevoCodigo); err != nil {
		return err
	}

	// 3. Enviarlo; si el correo falla, el codigo se anula
	asunto, motivo := "Tu codigo para iniciar sesion en FinanceUp", "iniciar sesion"
	if proposito == "recuperar_contrasena" {
		asunto, motivo = "Codigo para recuperar tu contrasena de FinanceUp", "cambiar tu contrasena"
	}

	mensaje := fmt.Sprintf("Hola %s,\n\nTu codigo para %s es: %s\n\nVence en %d minutos y solo sirve una vez.\nSi no fuiste tu, ignora este correo y cambia tu contrasena.\n\nFinanceUp",
		usuario.Nombre, motivo, codigo, minutos)

	if err := helpers.EnviarCorreo(usuario.Email, asunto, mensaje); err != nil {
		if nuevoCodigo.Id > 0 {
			nuevoCodigo.Activo = false
			helpers.Actualizar(fmt.Sprintf("%s/codigo_verificacion/%d", helpers.UrlCrudAuth, nuevoCodigo.Id), nuevoCodigo)
		}
		return err
	}

	return nil
}

// validarCodigo compara con el ultimo codigo activo: si es correcto lo anula, si no suma un intento.
func validarCodigo(idUsuario int, proposito string, codigoEscrito string) error {

	codigo, existe, err := ultimoCodigo(idUsuario, proposito, true)
	if err != nil {
		return err
	}

	if !existe {
		return helpers.NuevoError(http.StatusUnauthorized, "No hay un codigo pendiente. Pide uno nuevo.")
	}

	ahora := time.Now().UTC()

	if codigo.FechaExpiracion.Before(ahora) || codigo.Intentos >= intentosAntesDeBloquear {
		return helpers.NuevoError(http.StatusUnauthorized, "El codigo vencio. Pide uno nuevo.")
	}

	correcto := helpers.CoincideCifrado(codigo.CodigoHash, strings.TrimSpace(codigoEscrito))

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

// ------------------------------------------------------------
// Contrasena y sesion
// ------------------------------------------------------------

// guardarContrasena guarda la clave cifrada y quita intentos fallidos y bloqueo.
func guardarContrasena(idUsuario int, contrasena string) error {

	credencial, existe, err := buscarCredencial(idUsuario)
	if err != nil {
		return err
	}

	hash, err := helpers.Cifrar(contrasena)
	if err != nil {
		return err
	}

	ahora := time.Now().UTC()

	credencial.ContrasenaHash = hash
	credencial.Salt = hash[7:29] // bcrypt guarda la sal dentro del hash
	credencial.Algoritmo = "bcrypt"
	credencial.Activo = true
	credencial.FechaUltimoCambio = ahora
	credencial.FechaActualizacion = ahora
	credencial.IntentosFallidos = 0
	credencial.BloqueadoHasta = time.Time{}

	if !existe {
		credencial.IdUsuario = &models.Ref{Id: idUsuario}
		return helpers.Crear(helpers.UrlCrudAuth+"/credencial", credencial, nil)
	}

	return helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)
}

// sumarIntentoFallido cuenta un intento y bloquea 15 minutos al llegar a 5.
func sumarIntentoFallido(credencial *models.Credencial) {

	credencial.IntentosFallidos++

	if credencial.IntentosFallidos >= intentosAntesDeBloquear {
		credencial.IntentosFallidos = 0
		credencial.BloqueadoHasta = time.Now().UTC().Add(minutosDeBloqueo * time.Minute)
	}
}

// borrarUsuario deshace un registro a medias para que el correo quede libre.
func borrarUsuario(idUsuario int) {

	if credencial, existe, err := buscarCredencial(idUsuario); err == nil && existe {
		helpers.Borrar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id))
	}

	var asignaciones []models.UsuarioRol
	helpers.ConsultarLista(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d", helpers.UrlCrudAuth, idUsuario), &asignaciones)

	for _, asignacion := range asignaciones {
		helpers.Borrar(fmt.Sprintf("%s/usuario_rol/%d", helpers.UrlCrudAuth, asignacion.Id))
	}

	helpers.Borrar(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, idUsuario))
}

// sesionIniciada arma la respuesta de un login exitoso: token, rol y datos del usuario.
func sesionIniciada(usuario models.Usuario) (map[string]interface{}, error) {

	roles, err := rolesDelUsuario(usuario.Id)
	if err != nil {
		return nil, err
	}

	rol := rolPrincipal(roles)

	token, err := helpers.CrearToken(usuario.Id, rol)
	if err != nil {
		return nil, err
	}

	usuario.FechaUltimaSesion = time.Now().UTC()
	helpers.Actualizar(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, usuario.Id), usuario)

	return map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"token":      token,
		"expira_en":  int(helpers.DuracionToken.Seconds()),
		"id_usuario": usuario.Id,
		"nombre":     usuario.Nombre,
		"apellido":   usuario.Apellido,
		"email":      usuario.Email,
		"rol":        rol,
		"roles":      roles,
		"estado":     usuario.Estado,
	}, nil
}

// ------------------------------------------------------------
// Funciones que usan los controladores
// ------------------------------------------------------------

// ObtenerPerfil: datos del usuario, tipo de documento y roles.
func ObtenerPerfil(id int) (map[string]interface{}, error) {

	usuario, err := BuscarUsuario(id)
	if err != nil {
		return nil, err
	}

	tipoDocumento := ""
	if usuario.TipoDocumento != nil {
		var tipo models.TipoDocumento
		if helpers.ConsultarUno(fmt.Sprintf("%s/tipo_documento/%d", helpers.UrlCrudAuth, usuario.TipoDocumento.Id), &tipo, "") == nil {
			tipoDocumento = tipo.Codigo
		}
	}

	roles, err := rolesDelUsuario(id)
	if err != nil {
		return nil, err
	}

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
		"rol":              rolPrincipal(roles),
		"roles":            roles,
	}, nil
}

// Registrar crea la cuenta (usuario + clave + rol) y deja la sesion iniciada.
// Si algo falla a la mitad, borra lo creado.
func Registrar(datos models.DatosRegistro) (map[string]interface{}, error) {

	// 1. Validar
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Apellido = strings.TrimSpace(datos.Apellido)

	if utf8.RuneCountInString(datos.Nombre) < 3 || utf8.RuneCountInString(datos.Apellido) < 3 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El nombre y el apellido deben tener minimo 3 letras")
	}

	email, err := normalizarEmail(datos.Email)
	if err != nil {
		return nil, err
	}

	if err := validarContrasena(datos.Contrasena); err != nil {
		return nil, err
	}

	if !datos.AceptaTerminos {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Debes aceptar los terminos y condiciones")
	}

	// 2. El correo no debe estar registrado
	if _, existe, err := buscarUsuarioPorEmail(email); err != nil {
		return nil, err
	} else if existe {
		return nil, helpers.NuevoError(http.StatusConflict, "Ya existe un usuario con ese correo")
	}

	// 3. Crear el usuario (la cedula es opcional: vacia se guarda NULL)
	nuevoUsuario := models.Usuario{
		Nombre:                  datos.Nombre,
		Apellido:                datos.Apellido,
		Email:                   email,
		Telefono:                strings.TrimSpace(datos.Telefono),
		Ciudad:                  strings.TrimSpace(datos.Ciudad),
		Direccion:               strings.TrimSpace(datos.Direccion),
		AceptaTerminos:          true,
		FechaAceptacionTerminos: time.Now().UTC(),
		Estado:                  "activo",
		Activo:                  true,
	}

	if cedula := strings.TrimSpace(datos.Cedula); cedula != "" {
		nuevoUsuario.Cedula = &cedula
	}

	if datos.IdTipoDocumento > 0 {
		nuevoUsuario.TipoDocumento = &models.Ref{Id: datos.IdTipoDocumento}
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario", nuevoUsuario, &nuevoUsuario); err != nil {
		return nil, err
	}

	// 4. Clave y rol (si falla, se deshace todo)
	if guardarContrasena(nuevoUsuario.Id, datos.Contrasena) != nil || asignarRolUsuario(nuevoUsuario.Id) != nil {
		borrarUsuario(nuevoUsuario.Id)
		return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo completar el registro, intenta de nuevo")
	}

	// 5. Sesion iniciada
	sesion, err := sesionIniciada(nuevoUsuario)
	if err != nil {
		return nil, err
	}

	sesion["mensaje"] = "Usuario registrado"

	return sesion, nil
}

// IniciarSesion (paso 1): valida correo y clave y envia el codigo al correo.
// Bloquea 15 minutos tras 5 intentos fallidos.
func IniciarSesion(datos models.DatosLogin) (map[string]interface{}, error) {

	// Mismo mensaje si el correo no existe o la clave esta mal
	errorLogin := helpers.NuevoError(http.StatusUnauthorized, "Correo o contrasena incorrectos")

	email, err := normalizarEmail(datos.Email)
	if err != nil {
		return nil, errorLogin
	}

	usuario, existe, err := buscarUsuarioPorEmail(email)
	if err != nil {
		return nil, err
	}
	if !existe {
		return nil, errorLogin
	}

	credencial, tiene, err := buscarCredencial(usuario.Id)
	if err != nil {
		return nil, err
	}
	if !tiene {
		return nil, errorLogin // por ejemplo, una cuenta creada con Google
	}

	// Bloqueada: no se revisa la clave
	if credencial.BloqueadoHasta.After(time.Now().UTC()) {
		auditar(usuario.Id, "acceso_denegado", false, datos.Ip, datos.Navegador)
		return nil, helpers.NuevoError(http.StatusLocked, "Cuenta bloqueada por intentos fallidos. Intenta de nuevo en 15 minutos.")
	}

	claveCorrecta := helpers.CoincideCifrado(credencial.ContrasenaHash, datos.Contrasena)

	if claveCorrecta {
		credencial.IntentosFallidos = 0
	} else {
		sumarIntentoFallido(&credencial)
	}

	helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)
	auditar(usuario.Id, "login", claveCorrecta, datos.Ip, datos.Navegador)

	if !claveCorrecta {
		return nil, errorLogin
	}

	// Solo con la clave correcta se dice si la cuenta tiene problemas
	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	respuesta := map[string]interface{}{
		"mensaje":         "Te enviamos un codigo de verificacion a tu correo",
		"requiere_codigo": true,
		"id_usuario":      usuario.Id,
	}

	if err := crearCodigo(usuario, "login_2fa", minutosCodigoLogin); err != nil {
		// Si se envio uno hace menos de un minuto, ese sigue sirviendo
		if errorApi, ok := err.(*helpers.ErrorApi); ok && errorApi.Codigo == http.StatusTooManyRequests {
			respuesta["mensaje"] = "Ya te enviamos un codigo hace un momento, revisa tu correo"
			return respuesta, nil
		}
		return nil, err
	}

	return respuesta, nil
}

// VerificarCodigo (paso 2): valida el codigo y entrega el token.
func VerificarCodigo(datos models.DatosCodigo) (map[string]interface{}, error) {

	if datos.IdUsuario <= 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El id del usuario es obligatorio")
	}

	if err := validarCodigo(datos.IdUsuario, "login_2fa", datos.Codigo); err != nil {
		return nil, err
	}

	usuario, err := BuscarUsuario(datos.IdUsuario)
	if err != nil {
		return nil, err
	}

	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	return sesionIniciada(usuario)
}

// ReenviarCodigo: boton "Reenviar codigo". Solo si ya paso el paso 1 hace poco.
func ReenviarCodigo(datos models.DatosReenviar) (map[string]interface{}, error) {

	if datos.IdUsuario <= 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El id del usuario es obligatorio")
	}

	iniciaSesion := helpers.NuevoError(http.StatusUnauthorized, "Inicia sesion de nuevo para recibir un codigo")

	pendiente, existe, err := ultimoCodigo(datos.IdUsuario, "login_2fa", true)
	if err != nil {
		return nil, err
	}
	if !existe || pendiente.FechaExpiracion.Before(time.Now().UTC()) {
		return nil, iniciaSesion
	}

	usuario, err := BuscarUsuario(datos.IdUsuario)
	if err != nil {
		return nil, iniciaSesion
	}

	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	// El codigo anterior deja de servir: siempre se valida el mas nuevo
	if err := crearCodigo(usuario, "login_2fa", minutosCodigoLogin); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Te enviamos un codigo nuevo a tu correo", "id_usuario": usuario.Id}, nil
}

// IniciarSesionSocial: entrar con Google (verificando el token con Google).
func IniciarSesionSocial(datos models.DatosLoginSocial) (map[string]interface{}, error) {

	proveedor := strings.ToLower(strings.TrimSpace(datos.Proveedor))

	if proveedor == "apple" || proveedor == "facebook" {
		return nil, helpers.NuevoError(http.StatusNotImplemented, "El inicio de sesion con "+proveedor+" todavia no esta disponible")
	}
	if proveedor != "google" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El proveedor debe ser google")
	}

	google, err := helpers.VerificarTokenGoogle(datos.IdToken)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(google.Email))

	// 1. Cuenta de Google ya vinculada
	var cuentas []models.CuentaSocial
	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/cuenta_social?query=proveedor:google,id_externo:"+url.QueryEscape(google.IdExterno)+",activo:true", &cuentas); err != nil {
		return nil, err
	}

	var usuario models.Usuario

	if len(cuentas) > 0 && cuentas[0].IdUsuario != nil {

		if usuario, err = BuscarUsuario(cuentas[0].IdUsuario.Id); err != nil {
			return nil, err
		}

	} else {

		// 2. Si no, buscar por el correo que confirmo Google; si no existe, crearlo
		existente, existe, err := buscarUsuarioPorEmail(email)
		if err != nil {
			return nil, err
		}

		creado := false

		if existe {
			usuario = existente

			var vinculadas []models.CuentaSocial
			helpers.ConsultarLista(fmt.Sprintf("%s/cuenta_social?query=id_usuario:%d,proveedor:google", helpers.UrlCrudAuth, usuario.Id), &vinculadas)

			if len(vinculadas) > 0 {
				return nil, helpers.NuevoError(http.StatusConflict, "Esta cuenta ya esta vinculada a otra cuenta de Google")
			}
		} else {
			usuario = models.Usuario{Nombre: google.Nombre, Apellido: google.Apellido, Email: email, Estado: "activo", Activo: true}

			if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario", usuario, &usuario); err != nil {
				return nil, err
			}
			creado = true

			if err := asignarRolUsuario(usuario.Id); err != nil {
				borrarUsuario(usuario.Id)
				return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo completar el registro con Google, intenta de nuevo")
			}
		}

		// 3. Vincular la cuenta de Google
		cuenta := models.CuentaSocial{IdUsuario: &models.Ref{Id: usuario.Id}, Proveedor: "google", IdExterno: google.IdExterno, EmailProveedor: email, Activo: true}

		if err := helpers.Crear(helpers.UrlCrudAuth+"/cuenta_social", cuenta, nil); err != nil {
			if creado {
				borrarUsuario(usuario.Id)
			}
			return nil, err
		}
	}

	if err := revisarEstado(usuario); err != nil {
		auditar(usuario.Id, "acceso_denegado", false, datos.Ip, datos.Navegador)
		return nil, err
	}

	auditar(usuario.Id, "login", true, datos.Ip, datos.Navegador)

	sesion, err := sesionIniciada(usuario)
	if err != nil {
		return nil, err
	}

	sesion["proveedor"] = "google"

	return sesion, nil
}

// PedirCodigoRecuperacion envia un codigo para cambiar la clave olvidada.
// Siempre responde lo mismo, para no revelar que correos existen.
func PedirCodigoRecuperacion(email string) (map[string]interface{}, error) {

	respuesta := map[string]interface{}{"mensaje": "Si el correo esta registrado, te enviamos un codigo"}

	email, err := normalizarEmail(email)
	if err != nil {
		return nil, err
	}

	usuario, existe, err := buscarUsuarioPorEmail(email)
	if err != nil {
		return nil, err
	}

	if existe && revisarEstado(usuario) == nil {
		if err := crearCodigo(usuario, "recuperar_contrasena", minutosCodigoRecuperar); err != nil {
			fmt.Println("[RECUPERAR CONTRASENA] No se envio el codigo:", err.Error())
		}
	}

	return respuesta, nil
}

// RestablecerContrasena valida el codigo y guarda la clave nueva (quita el bloqueo).
func RestablecerContrasena(datos models.DatosContrasena) (map[string]interface{}, error) {

	if err := validarContrasena(datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	codigoIncorrecto := helpers.NuevoError(http.StatusUnauthorized, "El codigo no es correcto")

	email, err := normalizarEmail(datos.Email)
	if err != nil {
		return nil, codigoIncorrecto
	}

	usuario, existe, err := buscarUsuarioPorEmail(email)
	if err != nil {
		return nil, err
	}
	if !existe {
		return nil, codigoIncorrecto
	}

	if err := validarCodigo(usuario.Id, "recuperar_contrasena", datos.Codigo); err != nil {
		return nil, err
	}

	if err := guardarContrasena(usuario.Id, datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	auditar(usuario.Id, "cambio_contrasena", true, datos.Ip, datos.Navegador)

	return map[string]interface{}{"mensaje": "Contrasena actualizada"}, nil
}

// CambiarContrasena desde el perfil (pide la clave actual; bloquea tras 5 intentos).
func CambiarContrasena(idUsuario int, datos models.DatosContrasena) (map[string]interface{}, error) {

	if err := validarContrasena(datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	credencial, tiene, err := buscarCredencial(idUsuario)
	if err != nil {
		return nil, err
	}
	if !tiene {
		return nil, helpers.NuevoError(http.StatusNotFound, "Tu cuenta no tiene contrasena (entraste con Google). Usa \"Olvide mi contrasena\" para crear una.")
	}

	if credencial.BloqueadoHasta.After(time.Now().UTC()) {
		return nil, helpers.NuevoError(http.StatusLocked, "Demasiados intentos. Intenta de nuevo en 15 minutos.")
	}

	if !helpers.CoincideCifrado(credencial.ContrasenaHash, datos.ContrasenaActual) {
		sumarIntentoFallido(&credencial)
		helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)
		auditar(idUsuario, "cambio_contrasena", false, datos.Ip, datos.Navegador)
		return nil, helpers.NuevoError(http.StatusUnauthorized, "La contrasena actual no es correcta")
	}

	if err := guardarContrasena(idUsuario, datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	auditar(idUsuario, "cambio_contrasena", true, datos.Ip, datos.Navegador)

	return map[string]interface{}{"mensaje": "Contrasena actualizada"}, nil
}

// CerrarSesion anula el token y lo anota en la auditoria.
func CerrarSesion(token helpers.DatosToken, ip string, navegador string) map[string]interface{} {

	helpers.RevocarToken(token)
	auditar(token.IdUsuario, "logout", true, ip, navegador)

	return map[string]interface{}{"mensaje": "Sesion cerrada"}
}
