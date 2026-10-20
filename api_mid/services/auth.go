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


// SERVICE DE USUARIOS (crud_auth)
// Aqui esta la logica: validar, cifrar, bloquear, crear codigos.


// Tiempos y limites del login
const (
	minutosCodigoLogin      = 10 // el codigo del login vence en 10 minutos
	minutosCodigoRecuperar  = 15 // el de recuperar contrasena en 15
	segundosEntreCodigos    = 60 // no se manda mas de un codigo por minuto
	intentosAntesDeBloquear = 5  // a los 5 intentos fallidos se bloquea
	minutosDeBloqueo        = 15 // durante 15 minutos
)

// ------------------------------------------------------------
// Funciones de apoyo
// ------------------------------------------------------------

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

// normalizarEmail deja el correo en minuscula y sin espacios, y revisa
// que sea un correo de verdad (nombre@dominio.algo).
func normalizarEmail(email string) (string, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	invalido := helpers.NuevoError(http.StatusBadRequest, "El correo no es valido")

	direccion, err := mail.ParseAddress(email)

	// direccion.Address != email descarta cosas como "Ana <ana@x.com>".
	// Las comas y los dos puntos se rechazan porque rompen el ?query= del CRUD.
	if err != nil || direccion.Address != email || strings.ContainsAny(email, ",:") {
		return "", invalido
	}

	dominio := email[strings.LastIndex(email, "@")+1:]

	if !strings.Contains(dominio, ".") || utf8.RuneCountInString(email) > 150 {
		return "", invalido
	}

	return email, nil
}

// validarContrasena aplica las mismas reglas que la pagina:
// minimo 8 caracteres, una mayuscula y un numero (y maximo 72, el limite de bcrypt).
func validarContrasena(clave string) error {

	if utf8.RuneCountInString(clave) < 8 {
		return helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener minimo 8 caracteres")
	}

	if len(clave) > helpers.LargoMaximoClave {
		return helpers.NuevoError(http.StatusBadRequest, fmt.Sprintf("La contrasena no puede tener mas de %d caracteres", helpers.LargoMaximoClave))
	}

	tieneMayuscula := false
	tieneNumero := false

	for _, letra := range clave {
		if unicode.IsUpper(letra) {
			tieneMayuscula = true
		}
		if unicode.IsDigit(letra) {
			tieneNumero = true
		}
	}

	if !tieneMayuscula || !tieneNumero {
		return helpers.NuevoError(http.StatusBadRequest, "La contrasena debe tener al menos una mayuscula y un numero")
	}

	return nil
}

// revisarEstado dice si la cuenta puede entrar. Los estados son los de la
// tabla auth.usuario: activo, inactivo y suspendido.
// (El "bloqueado" de la pagina es el bloqueo de 15 minutos: error 423.)
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

// rolesDelUsuario devuelve los nombres de los roles activos del usuario
// (Administrador, Asesor, Usuario, Gerente).
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

// rolPrincipal convierte los roles de la base en el rol que usa la pagina:
// "admin" si es Administrador, si no "user".
func rolPrincipal(roles []string) string {

	for _, rol := range roles {
		if strings.EqualFold(rol, "Administrador") {
			return "admin"
		}
	}

	return "user"
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

// auditar guarda un evento en auditoria_login. Si falla no se detiene
// nada: la auditoria no debe impedir que el usuario use la app.
// tipoEvento: login, logout, cambio_contrasena o acceso_denegado.
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

// ultimoCodigo trae el codigo mas reciente del usuario para ese proposito.
// Con soloActivos = true ignora los que ya se usaron.
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

// crearCodigo crea un codigo de 6 numeros, lo guarda cifrado y lo envia
// al correo del usuario. No deja pedir mas de un codigo por minuto.
func crearCodigo(usuario models.Usuario, proposito string, minutos int) error {

	ahora := time.Now().UTC()

	// 1. Limite: maximo un codigo por minuto
	if anterior, existe, err := ultimoCodigo(usuario.Id, proposito, false); err != nil {
		return err
	} else if existe && !anterior.FechaCreacion.IsZero() && ahora.Sub(anterior.FechaCreacion) < segundosEntreCodigos*time.Second {
		return helpers.NuevoError(http.StatusTooManyRequests, "Ya te enviamos un codigo hace menos de un minuto. Revisa tu correo o espera un momento.")
	}

	// 2. Crear el codigo y guardarlo cifrado
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

	// 3. Enviarlo por correo
	asunto := "Tu codigo para iniciar sesion en FinanceUp"
	motivo := "iniciar sesion"

	if proposito == "recuperar_contrasena" {
		asunto = "Codigo para recuperar tu contrasena de FinanceUp"
		motivo = "cambiar tu contrasena"
	}

	mensaje := fmt.Sprintf("Hola %s,\n\nTu codigo para %s es: %s\n\nVence en %d minutos y solo sirve una vez.\nSi no fuiste tu, ignora este correo y cambia tu contrasena.\n\nFinanceUp",
		usuario.Nombre, motivo, codigo, minutos)

	if err := helpers.EnviarCorreo(usuario.Email, asunto, mensaje); err != nil {

		// Si no se pudo enviar, el codigo no sirve: se desactiva
		if nuevoCodigo.Id > 0 {
			nuevoCodigo.Activo = false
			helpers.Actualizar(fmt.Sprintf("%s/codigo_verificacion/%d", helpers.UrlCrudAuth, nuevoCodigo.Id), nuevoCodigo)
		}

		return err
	}

	return nil
}

// validarCodigo revisa el ultimo codigo activo del usuario.
// Si es correcto lo desactiva (no sirve dos veces); si no, suma un intento.
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

// guardarContrasena guarda una clave nueva cifrada (crea la credencial si no
// existe). Tambien borra los intentos fallidos y el bloqueo de 15 minutos.
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
	credencial.BloqueadoHasta = time.Time{} // sin bloqueo

	if !existe {
		credencial.IdUsuario = &models.Ref{Id: idUsuario}
		return helpers.Crear(helpers.UrlCrudAuth+"/credencial", credencial, nil)
	}

	return helpers.Actualizar(fmt.Sprintf("%s/credencial/%d", helpers.UrlCrudAuth, credencial.Id), credencial)
}

// sumarIntentoFallido cuenta un intento fallido y bloquea la cuenta
// 15 minutos al llegar a 5. Devuelve true si la acaba de bloquear.
func sumarIntentoFallido(credencial *models.Credencial) bool {

	credencial.IntentosFallidos++

	if credencial.IntentosFallidos >= intentosAntesDeBloquear {
		credencial.IntentosFallidos = 0
		credencial.BloqueadoHasta = time.Now().UTC().Add(minutosDeBloqueo * time.Minute)
		return true
	}

	return false
}

// borrarUsuario deshace un registro que quedo a medias: borra la
// credencial (si alcanzo a crearse) y el usuario, para liberar el correo.
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

// sesionIniciada arma la respuesta de un login exitoso: crea el token,
// busca el rol y guarda la fecha de la ultima sesion.
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
		"expira_en":  int(helpers.DuracionToken.Seconds()), // segundos
		"id_usuario": usuario.Id,
		"nombre":     usuario.Nombre,
		"apellido":   usuario.Apellido,
		"email":      usuario.Email,
		"rol":        rol,   // "admin" o "user" (lo que usa la pagina)
		"roles":      roles, // los nombres de la base de datos
		"estado":     usuario.Estado,
	}, nil
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
	roles, err := rolesDelUsuario(id)

	if err != nil {
		return nil, err
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
		"rol":              rolPrincipal(roles),
		"roles":            roles,
	}, nil
}

// Registrar: crea usuario + clave cifrada + rol "Usuario" y deja la
// sesion iniciada (devuelve el token). Si algo falla a la mitad, borra
// lo que alcanzo a crear para que el correo quede libre.
func Registrar(datos models.DatosRegistro) (map[string]interface{}, error) {

	// 1. Validar (las mismas reglas que la pagina)
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

	// 2. Revisar que el correo no este registrado
	_, existe, err := buscarUsuarioPorEmail(email)

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
		Email:                   email,
		Telefono:                strings.TrimSpace(datos.Telefono),
		Ciudad:                  strings.TrimSpace(datos.Ciudad),
		Direccion:               strings.TrimSpace(datos.Direccion),
		AceptaTerminos:          true,
		FechaAceptacionTerminos: time.Now().UTC(),
		Estado:                  "activo",
		Activo:                  true,
	}

	// La cedula es opcional: si no llega se deja en nil para que se guarde NULL
	if cedula := strings.TrimSpace(datos.Cedula); cedula != "" {
		nuevoUsuario.Cedula = &cedula
	}

	if datos.IdTipoDocumento > 0 {
		nuevoUsuario.TipoDocumento = &models.Ref{Id: datos.IdTipoDocumento}
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/usuario", nuevoUsuario, &nuevoUsuario); err != nil {
		return nil, err
	}

	// 4. Guardar la clave cifrada y el rol. Si falla, se deshace todo.
	if err := guardarContrasena(nuevoUsuario.Id, datos.Contrasena); err != nil {
		borrarUsuario(nuevoUsuario.Id)
		return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo completar el registro, intenta de nuevo")
	}

	if err := asignarRolUsuario(nuevoUsuario.Id); err != nil {
		borrarUsuario(nuevoUsuario.Id)
		return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo completar el registro, intenta de nuevo")
	}

	// 5. Dejar la sesion iniciada (la pagina entra directo despues del registro)
	sesion, err := sesionIniciada(nuevoUsuario)

	if err != nil {
		return nil, err
	}

	sesion["mensaje"] = "Usuario registrado"

	return sesion, nil
}

// IniciarSesion: paso 1 del login. Valida correo y clave y envia el codigo.
// Bloquea la cuenta 15 minutos tras 5 intentos fallidos.
func IniciarSesion(datos models.DatosLogin) (map[string]interface{}, error) {

	// Mismo mensaje si el correo no existe o si la clave esta mal,
	// para no contarle a un extrano que correos estan registrados.
	errorLogin := helpers.NuevoError(http.StatusUnauthorized, "Correo o contrasena incorrectos")

	email, err := normalizarEmail(datos.Email)

	if err != nil {
		return nil, errorLogin
	}

	// 1. Buscar el usuario y su credencial
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
		// Por ejemplo, una cuenta creada con Google que nunca puso clave
		return nil, errorLogin
	}

	// 2. Si esta bloqueada no se revisa la clave (asi no se puede adivinar)
	ahora := time.Now().UTC()

	if credencial.BloqueadoHasta.After(ahora) {
		auditar(usuario.Id, "acceso_denegado", false, datos.Ip, datos.Navegador)
		return nil, helpers.NuevoError(http.StatusLocked, "Cuenta bloqueada por intentos fallidos. Intenta de nuevo en 15 minutos.")
	}

	// 3. Comparar la clave y actualizar los intentos fallidos
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

	// 4. Solo con la clave correcta se dice si la cuenta tiene problemas
	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	// 5. Enviar el codigo de 6 numeros (vence en 10 minutos)
	respuesta := map[string]interface{}{
		"mensaje":         "Te enviamos un codigo de verificacion a tu correo",
		"requiere_codigo": true,
		"id_usuario":      usuario.Id,
	}

	if err := crearCodigo(usuario, "login_2fa", minutosCodigoLogin); err != nil {

		// Si ya se envio uno hace menos de un minuto, ese sigue sirviendo
		if errorApi, ok := err.(*helpers.ErrorApi); ok && errorApi.Codigo == http.StatusTooManyRequests {
			respuesta["mensaje"] = "Ya te enviamos un codigo hace un momento, revisa tu correo"
			return respuesta, nil
		}

		return nil, err
	}

	return respuesta, nil
}

// VerificarCodigo: paso 2 del login. Valida el codigo y entrega el token.
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

	// Por si la cuenta se suspendio mientras escribia el codigo
	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	return sesionIniciada(usuario)
}

// ReenviarCodigo: el boton "Reenviar codigo" del login.
// Solo funciona si la persona ya paso el paso 1 (correo y clave) hace
// poco, es decir, si tiene un codigo de login pendiente y sin vencer.
func ReenviarCodigo(datos models.DatosReenviar) (map[string]interface{}, error) {

	if datos.IdUsuario <= 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El id del usuario es obligatorio")
	}

	inicia := helpers.NuevoError(http.StatusUnauthorized, "Inicia sesion de nuevo para recibir un codigo")

	pendiente, existe, err := ultimoCodigo(datos.IdUsuario, "login_2fa", true)

	if err != nil {
		return nil, err
	}

	if !existe || pendiente.FechaExpiracion.Before(time.Now().UTC()) {
		return nil, inicia
	}

	usuario, err := BuscarUsuario(datos.IdUsuario)

	if err != nil {
		return nil, inicia
	}

	if err := revisarEstado(usuario); err != nil {
		return nil, err
	}

	// El codigo anterior queda sin efecto: validarCodigo siempre usa el mas nuevo
	if err := crearCodigo(usuario, "login_2fa", minutosCodigoLogin); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"mensaje":    "Te enviamos un codigo nuevo a tu correo",
		"id_usuario": usuario.Id,
	}, nil
}

// IniciarSesionSocial: entrar con Google.
// La pagina manda el id_token que le dio Google; aqui se verifica con
// Google y solo se usan el correo y el id que Google confirma.
func IniciarSesionSocial(datos models.DatosLoginSocial) (map[string]interface{}, error) {

	proveedor := strings.ToLower(strings.TrimSpace(datos.Proveedor))

	// 1. Por ahora solo Google se puede verificar
	if proveedor == "apple" || proveedor == "facebook" {
		return nil, helpers.NuevoError(http.StatusNotImplemented, "El inicio de sesion con "+proveedor+" todavia no esta disponible")
	}

	if proveedor != "google" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El proveedor debe ser google")
	}

	// 2. Preguntarle a Google si el token es real
	google, err := helpers.VerificarTokenGoogle(datos.IdToken)

	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(google.Email))

	// 3. Si la cuenta de Google ya estaba vinculada, entra con ese usuario
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

		// 4. Si no, buscar el usuario por el correo que Google confirmo; si no existe, crearlo
		usuarioExistente, existe, err := buscarUsuarioPorEmail(email)

		if err != nil {
			return nil, err
		}

		creado := false

		if existe {

			usuario = usuarioExistente

			// Ese usuario ya tiene otra cuenta de Google vinculada (la tabla no deja dos)
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

		// 5. Vincular la cuenta de Google al usuario
		cuenta := models.CuentaSocial{
			IdUsuario:      &models.Ref{Id: usuario.Id},
			Proveedor:      "google",
			IdExterno:      google.IdExterno,
			EmailProveedor: email,
			Activo:         true,
		}

		if err := helpers.Crear(helpers.UrlCrudAuth+"/cuenta_social", cuenta, nil); err != nil {
			if creado {
				borrarUsuario(usuario.Id)
			}
			return nil, err
		}
	}

	// 6. La cuenta debe estar activa
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

// PedirCodigoRecuperacion: crea un codigo para cambiar la clave olvidada.
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

		// Si falla (por ejemplo, el limite de un codigo por minuto) se
		// anota en la consola del MID, pero la respuesta es la misma.
		if err := crearCodigo(usuario, "recuperar_contrasena", minutosCodigoRecuperar); err != nil {
			fmt.Println("[RECUPERAR CONTRASENA] No se envio el codigo:", err.Error())
		}
	}

	return respuesta, nil
}

// RestablecerContrasena: valida el codigo y guarda la clave nueva.
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

	// guardarContrasena tambien quita el bloqueo de 15 minutos
	if err := guardarContrasena(usuario.Id, datos.ContrasenaNueva); err != nil {
		return nil, err
	}

	auditar(usuario.Id, "cambio_contrasena", true, datos.Ip, datos.Navegador)

	return map[string]interface{}{"mensaje": "Contrasena actualizada"}, nil
}

// CambiarContrasena: desde el perfil, pidiendo la clave actual.
// Tambien bloquea 15 minutos tras 5 intentos con la clave actual mal.
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

// CerrarSesion: invalida el token y lo anota en la auditoria.
func CerrarSesion(token helpers.DatosToken, ip string, navegador string) map[string]interface{} {

	helpers.RevocarToken(token)
	auditar(token.IdUsuario, "logout", true, ip, navegador)

	return map[string]interface{}{"mensaje": "Sesion cerrada"}
}
