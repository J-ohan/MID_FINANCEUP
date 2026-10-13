package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type Usuario struct {
	Id                      int       `json:"Id"`
	TipoDocumento           *Ref      `json:"TipoDocumento"`
	Nombre                  string    `json:"Nombre"`
	Apellido                string    `json:"Apellido"`
	Email                   string    `json:"Email"`
	Telefono                string    `json:"Telefono"`
	Cedula                  string    `json:"Cedula"`
	Ciudad                  string    `json:"Ciudad"`
	Direccion               string    `json:"Direccion"`
	FechaNacimiento         time.Time `json:"FechaNacimiento"`
	AceptaTerminos          bool      `json:"AceptaTerminos"`
	FechaAceptacionTerminos time.Time `json:"FechaAceptacionTerminos"`
	Estado                  string    `json:"Estado"`
	FechaRegistro           time.Time `json:"FechaRegistro"`
	FechaUltimaSesion       time.Time `json:"FechaUltimaSesion"`
	Activo                  bool      `json:"Activo"`
	FechaCreacion           time.Time `json:"FechaCreacion"`
	FechaModificacion       time.Time `json:"FechaModificacion"`
}

type Credencial struct {
	Id                 int       `json:"Id"`
	IdUsuario          *Ref      `json:"IdUsuario"`
	ContrasenaHash     string    `json:"ContrasenaHash"`
	Salt               string    `json:"Salt"`
	Algoritmo          string    `json:"Algoritmo"`
	FechaActualizacion time.Time `json:"FechaActualizacion"`
	FechaUltimoCambio  time.Time `json:"FechaUltimoCambio"`
	IntentosFallidos   int       `json:"IntentosFallidos"`
	BloqueadoHasta     time.Time `json:"BloqueadoHasta"`
	RequiereCambio     bool      `json:"RequiereCambio"`
	Activo             bool      `json:"Activo"`
	FechaCreacion      time.Time `json:"FechaCreacion"`
	FechaModificacion  time.Time `json:"FechaModificacion"`
}

type CodigoVerificacion struct {
	Id                int       `json:"Id"`
	IdUsuario         *Ref      `json:"IdUsuario"`
	Proposito         string    `json:"Proposito"`
	Canal             string    `json:"Canal"`
	CodigoHash        string    `json:"CodigoHash"`
	FechaExpiracion   time.Time `json:"FechaExpiracion"`
	FechaUso          time.Time `json:"FechaUso"`
	Intentos          int       `json:"Intentos"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type CuentaSocial struct {
	Id             int    `json:"Id"`
	IdUsuario      *Ref   `json:"IdUsuario"`
	Proveedor      string `json:"Proveedor"`
	IdExterno      string `json:"IdExterno"`
	EmailProveedor string `json:"EmailProveedor"`
	Activo         bool   `json:"Activo"`
}

type Rol struct {
	Id        int    `json:"Id"`
	NombreRol string `json:"NombreRol"`
}

type UsuarioRol struct {
	Id        int  `json:"Id"`
	IdUsuario *Ref `json:"IdUsuario"`
	IdRol     *Ref `json:"IdRol"`
	Activo    bool `json:"Activo"`
}

type TipoDocumento struct {
	Id     int    `json:"Id"`
	Codigo string `json:"Codigo"`
}

type AuditoriaLogin struct {
	IdUsuario    *Ref   `json:"IdUsuario"`
	TipoEvento   string `json:"TipoEvento"`
	IpAddress    string `json:"IpAddress"`
	Navegador    string `json:"Navegador"`
	EstadoEvento string `json:"EstadoEvento"`
}

type DatosRegistro struct {
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	Email           string `json:"email"`
	Contrasena      string `json:"contrasena"`
	IdTipoDocumento int    `json:"id_tipo_documento"`
	Cedula          string `json:"cedula"`
	Telefono        string `json:"telefono"`
	Ciudad          string `json:"ciudad"`
	Direccion       string `json:"direccion"`
	AceptaTerminos  bool   `json:"acepta_terminos"`
}

type DatosLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

type DatosCodigo struct {
	IdUsuario int    `json:"id_usuario"`
	Email     string `json:"email"`
	Codigo    string `json:"codigo"`
}

type DatosLoginSocial struct {
	Proveedor string `json:"proveedor"`
	IdExterno string `json:"id_externo"`
	Email     string `json:"email"`
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
}

type DatosContrasena struct {
	Email            string `json:"email"`
	Codigo           string `json:"codigo"`
	ContrasenaActual string `json:"contrasena_actual"`
	ContrasenaNueva  string `json:"contrasena_nueva"`
}

// AuthController operations for Auth
type AuthController struct {
	beego.Controller
}

// ============================================================
// GET /v1/perfil/:id
// Usuario + tipo de documento + roles (crud_auth)
// ============================================================

func (c *AuthController) GetPerfil() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar el usuario en crud_auth
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("%s/usuario/%d", urlCrudAuth, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario

	// Si el usuario no existe, el CRUD responde un texto y no un usuario
	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar su tipo de documento
	// --------------------------------------------------------

	tipoDocumento := ""

	if usuario.TipoDocumento != nil {

		responseTipo, err := http.Get(fmt.Sprintf("%s/tipo_documento/%d", urlCrudAuth, usuario.TipoDocumento.Id))

		if err == nil {
			defer responseTipo.Body.Close()
			bodyTipo, _ := io.ReadAll(responseTipo.Body)

			var tipo TipoDocumento
			if json.Unmarshal(bodyTipo, &tipo) == nil {
				tipoDocumento = tipo.Codigo
			}
		}
	}

	// --------------------------------------------------------
	// 4. Consultar sus roles
	// --------------------------------------------------------

	responseRoles, err := http.Get(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d,activo:true&limit=1000", urlCrudAuth, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los roles"}
		c.ServeJSON()
		return
	}

	defer responseRoles.Body.Close()

	bodyRoles, _ := io.ReadAll(responseRoles.Body)

	var asignaciones []UsuarioRol
	json.Unmarshal(bodyRoles, &asignaciones)

	roles := []string{}

	for _, asignacion := range asignaciones {

		responseRol, err := http.Get(fmt.Sprintf("%s/rol/%d", urlCrudAuth, asignacion.IdRol.Id))

		if err != nil {
			continue
		}

		bodyRol, _ := io.ReadAll(responseRol.Body)
		responseRol.Body.Close()

		var rol Rol
		if json.Unmarshal(bodyRol, &rol) == nil {
			roles = append(roles, rol.NombreRol)
		}
	}

	// --------------------------------------------------------
	// 5. Construir el perfil y devolverlo
	// --------------------------------------------------------

	fechaNacimiento := ""
	if !usuario.FechaNacimiento.IsZero() {
		fechaNacimiento = usuario.FechaNacimiento.Format("2006-01-02")
	}

	c.Data["json"] = map[string]interface{}{
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
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/registro
// Crea usuario, clave cifrada y rol "Usuario" (crud_auth)
// ============================================================

func (c *AuthController) PostRegistro() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosRegistro

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))

	// --------------------------------------------------------
	// 2. Validar los datos
	// --------------------------------------------------------

	if datos.Nombre == "" || datos.Apellido == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El nombre y el apellido son obligatorios"}
		c.ServeJSON()
		return
	}

	if !strings.Contains(datos.Email, "@") {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El correo no es valido"}
		c.ServeJSON()
		return
	}

	if len(datos.Contrasena) < 8 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "La contrasena debe tener minimo 8 caracteres"}
		c.ServeJSON()
		return
	}

	if !datos.AceptaTerminos {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Debes aceptar los terminos y condiciones"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Revisar que el correo no este registrado
	// --------------------------------------------------------

	responseRepetido, err := http.Get(urlCrudAuth + "/usuario?query=email:" + url.QueryEscape(datos.Email))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseRepetido.Body.Close()

	bodyRepetido, _ := io.ReadAll(responseRepetido.Body)

	var repetidos []Usuario
	json.Unmarshal(bodyRepetido, &repetidos)

	if len(repetidos) > 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusConflict)
		c.Data["json"] = map[string]interface{}{"error": "Ya existe un usuario con ese correo"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Buscar el rol "Usuario"
	// --------------------------------------------------------

	responseRol, err := http.Get(urlCrudAuth + "/rol?query=nombre_rol:Usuario")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los roles"}
		c.ServeJSON()
		return
	}

	defer responseRol.Body.Close()

	bodyRol, _ := io.ReadAll(responseRol.Body)

	var roles []Rol
	json.Unmarshal(bodyRol, &roles)

	if len(roles) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"error": "No existe el rol Usuario en la base de datos"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Cifrar la contrasena (nunca se guarda en texto plano)
	// --------------------------------------------------------

	hash, _ := bcrypt.GenerateFromPassword([]byte(datos.Contrasena), bcrypt.DefaultCost)

	// --------------------------------------------------------
	// 6. Crear el usuario en crud_auth
	// --------------------------------------------------------

	nuevoUsuario := Usuario{
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
		nuevoUsuario.TipoDocumento = &Ref{Id: datos.IdTipoDocumento}
	}

	jsonUsuario, _ := json.Marshal(nuevoUsuario)

	responseCrear, err := http.Post(urlCrudAuth+"/usuario", "application/json", bytes.NewBuffer(jsonUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseCrear.Body.Close()

	bodyCrear, _ := io.ReadAll(responseCrear.Body)

	// El CRUD responde 201 cuando crea el registro
	if responseCrear.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo crear el usuario: " + string(bodyCrear)}
		c.ServeJSON()
		return
	}

	json.Unmarshal(bodyCrear, &nuevoUsuario)

	// --------------------------------------------------------
	// 7. Guardar la clave cifrada
	// --------------------------------------------------------

	credencial := Credencial{
		IdUsuario:      &Ref{Id: nuevoUsuario.Id},
		ContrasenaHash: string(hash),
		Salt:           string(hash[7:29]),
		Algoritmo:      "bcrypt",
		Activo:         true,
	}

	jsonCredencial, _ := json.Marshal(credencial)

	responseCredencial, err := http.Post(urlCrudAuth+"/credencial", "application/json", bytes.NewBuffer(jsonCredencial))

	if err != nil || responseCredencial.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El usuario se creo pero no se pudo guardar la contrasena"}
		c.ServeJSON()
		return
	}

	responseCredencial.Body.Close()

	// --------------------------------------------------------
	// 8. Asignar el rol "Usuario"
	// --------------------------------------------------------

	asignacion := UsuarioRol{
		IdUsuario: &Ref{Id: nuevoUsuario.Id},
		IdRol:     &Ref{Id: roles[0].Id},
		Activo:    true,
	}

	jsonAsignacion, _ := json.Marshal(asignacion)

	responseAsignacion, err := http.Post(urlCrudAuth+"/usuario_rol", "application/json", bytes.NewBuffer(jsonAsignacion))

	if err != nil || responseAsignacion.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El usuario se creo pero no se pudo asignar el rol"}
		c.ServeJSON()
		return
	}

	responseAsignacion.Body.Close()

	// --------------------------------------------------------
	// 9. Devolver el usuario creado
	// --------------------------------------------------------

	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"mensaje":    "Usuario registrado",
		"id_usuario": nuevoUsuario.Id,
		"nombre":     nuevoUsuario.Nombre,
		"apellido":   nuevoUsuario.Apellido,
		"email":      nuevoUsuario.Email,
		"rol":        "Usuario",
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/login
// Paso 1 del inicio de sesion: valida correo y clave.
// Si estan bien, crea un codigo de 6 numeros (2FA).
// Bloquea la cuenta 15 minutos tras 5 intentos fallidos.
// ============================================================

func (c *AuthController) PostLogin() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosLogin

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	email := strings.ToLower(strings.TrimSpace(datos.Email))

	// --------------------------------------------------------
	// 2. Buscar el usuario por correo
	// --------------------------------------------------------

	responseUsuario, err := http.Get(urlCrudAuth + "/usuario?query=email:" + url.QueryEscape(email))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuarios []Usuario
	json.Unmarshal(bodyUsuario, &usuarios)

	// Mismo mensaje si el correo no existe o si la clave esta mal
	if len(usuarios) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "Correo o contrasena incorrectos"}
		c.ServeJSON()
		return
	}

	usuario := usuarios[0]

	if !usuario.Activo || usuario.Estado != "activo" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		c.Data["json"] = map[string]interface{}{"error": "La cuenta esta inactiva"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Buscar su credencial (la clave cifrada)
	// --------------------------------------------------------

	responseCredencial, err := http.Get(fmt.Sprintf("%s/credencial?query=id_usuario:%d", urlCrudAuth, usuario.Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar la credencial"}
		c.ServeJSON()
		return
	}

	defer responseCredencial.Body.Close()

	bodyCredencial, _ := io.ReadAll(responseCredencial.Body)

	var credenciales []Credencial
	json.Unmarshal(bodyCredencial, &credenciales)

	if len(credenciales) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "Correo o contrasena incorrectos"}
		c.ServeJSON()
		return
	}

	credencial := credenciales[0]
	ahora := time.Now().UTC()

	// --------------------------------------------------------
	// 4. Revisar si la cuenta esta bloqueada
	// --------------------------------------------------------

	if credencial.BloqueadoHasta.After(ahora) {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusLocked)
		c.Data["json"] = map[string]interface{}{"error": "Cuenta bloqueada por intentos fallidos. Intenta de nuevo en 15 minutos."}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Comparar la clave
	// --------------------------------------------------------

	claveCorrecta := bcrypt.CompareHashAndPassword([]byte(credencial.ContrasenaHash), []byte(datos.Contrasena)) == nil

	// --------------------------------------------------------
	// 6. Actualizar los intentos fallidos en crud_auth
	//    (al 5to intento se bloquea 15 minutos)
	// --------------------------------------------------------

	if claveCorrecta {
		credencial.IntentosFallidos = 0
	} else {
		credencial.IntentosFallidos++
		if credencial.IntentosFallidos >= 5 {
			credencial.IntentosFallidos = 0
			credencial.BloqueadoHasta = ahora.Add(15 * time.Minute)
		}
	}

	jsonCredencial, _ := json.Marshal(credencial)

	peticionPut, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/credencial/%d", urlCrudAuth, credencial.Id), bytes.NewBuffer(jsonCredencial))
	peticionPut.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionPut); err == nil {
		responsePut.Body.Close()
	}

	// --------------------------------------------------------
	// 7. Guardar el intento en la auditoria
	// --------------------------------------------------------

	estadoEvento := "exitoso"
	if !claveCorrecta {
		estadoEvento = "fallido"
	}

	auditoria := AuditoriaLogin{
		IdUsuario:    &Ref{Id: usuario.Id},
		TipoEvento:   "login",
		IpAddress:    c.Ctx.Input.IP(),
		Navegador:    c.Ctx.Input.UserAgent(),
		EstadoEvento: estadoEvento,
	}

	jsonAuditoria, _ := json.Marshal(auditoria)

	if responseAuditoria, err := http.Post(urlCrudAuth+"/auditoria_login", "application/json", bytes.NewBuffer(jsonAuditoria)); err == nil {
		responseAuditoria.Body.Close()
	}

	if !claveCorrecta {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "Correo o contrasena incorrectos"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 8. Crear el codigo de verificacion (6 numeros, 10 minutos)
	//    Se guarda cifrado, igual que la contrasena.
	// --------------------------------------------------------

	numero, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	codigo := fmt.Sprintf("%06d", numero.Int64())

	hashCodigo, _ := bcrypt.GenerateFromPassword([]byte(codigo), bcrypt.DefaultCost)

	nuevoCodigo := CodigoVerificacion{
		IdUsuario:       &Ref{Id: usuario.Id},
		Proposito:       "login_2fa",
		Canal:           "email",
		CodigoHash:      string(hashCodigo),
		FechaExpiracion: ahora.Add(10 * time.Minute),
		Activo:          true,
	}

	jsonCodigo, _ := json.Marshal(nuevoCodigo)

	responseCodigo, err := http.Post(urlCrudAuth+"/codigo_verificacion", "application/json", bytes.NewBuffer(jsonCodigo))

	if err != nil || responseCodigo.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo crear el codigo de verificacion"}
		c.ServeJSON()
		return
	}

	responseCodigo.Body.Close()

	// --------------------------------------------------------
	// 9. "Enviar" el codigo
	//    Por ahora se muestra en la consola del MID. Cuando haya
	//    servicio de correo, aqui se envia al email del usuario.
	// --------------------------------------------------------

	fmt.Println("[CODIGO LOGIN]", usuario.Email, codigo)

	// --------------------------------------------------------
	// 10. Responder: falta el paso 2 (POST /v1/login/verificar)
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"mensaje":         "Te enviamos un codigo de verificacion",
		"requiere_codigo": true,
		"id_usuario":      usuario.Id,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/login/verificar
// Paso 2 del inicio de sesion: valida el codigo de 6 numeros
// ============================================================

func (c *AuthController) PostVerificarCodigo() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosCodigo

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Buscar el ultimo codigo activo del usuario
	// --------------------------------------------------------

	responseCodigos, err := http.Get(fmt.Sprintf("%s/codigo_verificacion?query=id_usuario:%d,proposito:login_2fa,activo:true&sortby=Id&order=desc&limit=1", urlCrudAuth, datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseCodigos.Body.Close()

	bodyCodigos, _ := io.ReadAll(responseCodigos.Body)

	var codigos []CodigoVerificacion
	json.Unmarshal(bodyCodigos, &codigos)

	if len(codigos) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "No hay un codigo pendiente. Inicia sesion de nuevo."}
		c.ServeJSON()
		return
	}

	codigo := codigos[0]
	ahora := time.Now().UTC()

	// --------------------------------------------------------
	// 3. Revisar que no este vencido ni tenga 5 intentos
	// --------------------------------------------------------

	if codigo.FechaExpiracion.Before(ahora) || codigo.Intentos >= 5 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "El codigo vencio. Inicia sesion de nuevo."}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Comparar el codigo y guardar el resultado
	//    Bien: se desactiva (no se puede usar dos veces)
	//    Mal: se suma un intento
	// --------------------------------------------------------

	codigoCorrecto := bcrypt.CompareHashAndPassword([]byte(codigo.CodigoHash), []byte(datos.Codigo)) == nil

	if codigoCorrecto {
		codigo.Activo = false
		codigo.FechaUso = ahora
	} else {
		codigo.Intentos++
	}

	jsonCodigo, _ := json.Marshal(codigo)

	peticionCodigo, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/codigo_verificacion/%d", urlCrudAuth, codigo.Id), bytes.NewBuffer(jsonCodigo))
	peticionCodigo.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionCodigo); err == nil {
		responsePut.Body.Close()
	}

	if !codigoCorrecto {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "El codigo no es correcto"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Consultar el usuario y guardar la fecha de su sesion
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("%s/usuario/%d", urlCrudAuth, datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario
	json.Unmarshal(bodyUsuario, &usuario)

	usuario.FechaUltimaSesion = ahora

	jsonUsuario, _ := json.Marshal(usuario)

	peticionUsuario, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/usuario/%d", urlCrudAuth, usuario.Id), bytes.NewBuffer(jsonUsuario))
	peticionUsuario.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionUsuario); err == nil {
		responsePut.Body.Close()
	}

	// --------------------------------------------------------
	// 6. Responder con los datos de la sesion
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"id_usuario": usuario.Id,
		"nombre":     usuario.Nombre,
		"apellido":   usuario.Apellido,
		"email":      usuario.Email,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/login/social
// Entrar con Google, Apple o Facebook.
// Si la cuenta social no existe, se crea (y el usuario tambien).
// Nota: la pagina es la que habla con Google/Apple/Facebook;
// aqui llegan los datos que esos servicios devolvieron.
// ============================================================

func (c *AuthController) PostLoginSocial() {

	// --------------------------------------------------------
	// 1. Leer y validar los datos enviados
	// --------------------------------------------------------

	var datos DatosLoginSocial

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))

	if datos.Proveedor != "google" && datos.Proveedor != "apple" && datos.Proveedor != "facebook" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El proveedor debe ser: google, apple o facebook"}
		c.ServeJSON()
		return
	}

	if datos.IdExterno == "" || datos.Email == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Faltan el id_externo o el email"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Buscar si la cuenta social ya existe
	// --------------------------------------------------------

	responseCuenta, err := http.Get(urlCrudAuth + "/cuenta_social?query=proveedor:" + datos.Proveedor + ",id_externo:" + url.QueryEscape(datos.IdExterno))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseCuenta.Body.Close()

	bodyCuenta, _ := io.ReadAll(responseCuenta.Body)

	var cuentas []CuentaSocial
	json.Unmarshal(bodyCuenta, &cuentas)

	idUsuario := 0

	if len(cuentas) > 0 {
		idUsuario = cuentas[0].IdUsuario.Id
	}

	// --------------------------------------------------------
	// 3. Si no existe, buscar el usuario por correo
	// --------------------------------------------------------

	if idUsuario == 0 {

		responseUsuarios, err := http.Get(urlCrudAuth + "/usuario?query=email:" + url.QueryEscape(datos.Email))

		if err != nil {
			c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
			c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
			c.ServeJSON()
			return
		}

		defer responseUsuarios.Body.Close()

		bodyUsuarios, _ := io.ReadAll(responseUsuarios.Body)

		var usuarios []Usuario
		json.Unmarshal(bodyUsuarios, &usuarios)

		if len(usuarios) > 0 {
			idUsuario = usuarios[0].Id
		}
	}

	// --------------------------------------------------------
	// 4. Si tampoco existe el usuario, crearlo con rol "Usuario"
	//    (no tiene contrasena: siempre entra con su red social)
	// --------------------------------------------------------

	if idUsuario == 0 {

		nuevoUsuario := Usuario{
			Nombre:   datos.Nombre,
			Apellido: datos.Apellido,
			Email:    datos.Email,
			Estado:   "activo",
			Activo:   true,
		}

		jsonUsuario, _ := json.Marshal(nuevoUsuario)

		responseCrear, err := http.Post(urlCrudAuth+"/usuario", "application/json", bytes.NewBuffer(jsonUsuario))

		if err != nil || responseCrear.StatusCode != http.StatusCreated {
			c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
			c.Data["json"] = map[string]interface{}{"error": "No se pudo crear el usuario"}
			c.ServeJSON()
			return
		}

		defer responseCrear.Body.Close()

		bodyCrear, _ := io.ReadAll(responseCrear.Body)
		json.Unmarshal(bodyCrear, &nuevoUsuario)
		idUsuario = nuevoUsuario.Id

		responseRol, err := http.Get(urlCrudAuth + "/rol?query=nombre_rol:Usuario")

		if err == nil {
			defer responseRol.Body.Close()
			bodyRol, _ := io.ReadAll(responseRol.Body)

			var roles []Rol
			json.Unmarshal(bodyRol, &roles)

			if len(roles) > 0 {
				asignacion := UsuarioRol{IdUsuario: &Ref{Id: idUsuario}, IdRol: &Ref{Id: roles[0].Id}, Activo: true}
				jsonAsignacion, _ := json.Marshal(asignacion)

				if responseAsignacion, err := http.Post(urlCrudAuth+"/usuario_rol", "application/json", bytes.NewBuffer(jsonAsignacion)); err == nil {
					responseAsignacion.Body.Close()
				}
			}
		}
	}

	// --------------------------------------------------------
	// 5. Vincular la cuenta social (si era nueva)
	// --------------------------------------------------------

	if len(cuentas) == 0 {

		cuenta := CuentaSocial{
			IdUsuario:      &Ref{Id: idUsuario},
			Proveedor:      datos.Proveedor,
			IdExterno:      datos.IdExterno,
			EmailProveedor: datos.Email,
			Activo:         true,
		}

		jsonCuenta, _ := json.Marshal(cuenta)

		if responseVincular, err := http.Post(urlCrudAuth+"/cuenta_social", "application/json", bytes.NewBuffer(jsonCuenta)); err == nil {
			responseVincular.Body.Close()
		}
	}

	// --------------------------------------------------------
	// 6. Responder con los datos de la sesion
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"id_usuario": idUsuario,
		"email":      datos.Email,
		"proveedor":  datos.Proveedor,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/recuperar-contrasena
// Crea un codigo para cambiar la contrasena olvidada
// ============================================================

func (c *AuthController) PostRecuperarContrasena() {

	// --------------------------------------------------------
	// 1. Leer el correo enviado
	// --------------------------------------------------------

	var datos DatosContrasena

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	email := strings.ToLower(strings.TrimSpace(datos.Email))

	// Siempre se responde lo mismo, exista o no el correo,
	// para que nadie pueda averiguar que correos estan registrados.
	respuesta := map[string]interface{}{"mensaje": "Si el correo esta registrado, te enviamos un codigo"}

	// --------------------------------------------------------
	// 2. Buscar el usuario por correo
	// --------------------------------------------------------

	responseUsuario, err := http.Get(urlCrudAuth + "/usuario?query=email:" + url.QueryEscape(email))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuarios []Usuario
	json.Unmarshal(bodyUsuario, &usuarios)

	if len(usuarios) == 0 {
		c.Data["json"] = respuesta
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Crear el codigo (6 numeros, 15 minutos) y "enviarlo"
	// --------------------------------------------------------

	numero, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	codigo := fmt.Sprintf("%06d", numero.Int64())

	hashCodigo, _ := bcrypt.GenerateFromPassword([]byte(codigo), bcrypt.DefaultCost)

	nuevoCodigo := CodigoVerificacion{
		IdUsuario:       &Ref{Id: usuarios[0].Id},
		Proposito:       "recuperar_contrasena",
		Canal:           "email",
		CodigoHash:      string(hashCodigo),
		FechaExpiracion: time.Now().UTC().Add(15 * time.Minute),
		Activo:          true,
	}

	jsonCodigo, _ := json.Marshal(nuevoCodigo)

	if responseCodigo, err := http.Post(urlCrudAuth+"/codigo_verificacion", "application/json", bytes.NewBuffer(jsonCodigo)); err == nil {
		responseCodigo.Body.Close()
	}

	fmt.Println("[CODIGO RECUPERAR]", email, codigo)

	// --------------------------------------------------------
	// 4. Responder
	// --------------------------------------------------------

	c.Data["json"] = respuesta
	c.ServeJSON()
}

// ============================================================
// PUT /v1/recuperar-contrasena
// Valida el codigo y guarda la contrasena nueva
// ============================================================

func (c *AuthController) PutRestablecerContrasena() {

	// --------------------------------------------------------
	// 1. Leer y validar los datos enviados
	// --------------------------------------------------------

	var datos DatosContrasena

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	if len(datos.ContrasenaNueva) < 8 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "La contrasena debe tener minimo 8 caracteres"}
		c.ServeJSON()
		return
	}

	email := strings.ToLower(strings.TrimSpace(datos.Email))

	// --------------------------------------------------------
	// 2. Buscar el usuario por correo
	// --------------------------------------------------------

	responseUsuario, err := http.Get(urlCrudAuth + "/usuario?query=email:" + url.QueryEscape(email))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuarios []Usuario
	json.Unmarshal(bodyUsuario, &usuarios)

	if len(usuarios) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "El codigo no es correcto"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Buscar el ultimo codigo activo y revisarlo
	// --------------------------------------------------------

	responseCodigos, err := http.Get(fmt.Sprintf("%s/codigo_verificacion?query=id_usuario:%d,proposito:recuperar_contrasena,activo:true&sortby=Id&order=desc&limit=1", urlCrudAuth, usuarios[0].Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseCodigos.Body.Close()

	bodyCodigos, _ := io.ReadAll(responseCodigos.Body)

	var codigos []CodigoVerificacion
	json.Unmarshal(bodyCodigos, &codigos)

	ahora := time.Now().UTC()

	if len(codigos) == 0 || codigos[0].FechaExpiracion.Before(ahora) || codigos[0].Intentos >= 5 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "El codigo vencio. Pide uno nuevo."}
		c.ServeJSON()
		return
	}

	codigo := codigos[0]
	codigoCorrecto := bcrypt.CompareHashAndPassword([]byte(codigo.CodigoHash), []byte(datos.Codigo)) == nil

	if codigoCorrecto {
		codigo.Activo = false
		codigo.FechaUso = ahora
	} else {
		codigo.Intentos++
	}

	jsonCodigo, _ := json.Marshal(codigo)

	peticionCodigo, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/codigo_verificacion/%d", urlCrudAuth, codigo.Id), bytes.NewBuffer(jsonCodigo))
	peticionCodigo.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionCodigo); err == nil {
		responsePut.Body.Close()
	}

	if !codigoCorrecto {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "El codigo no es correcto"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Buscar la credencial y guardar la clave nueva cifrada
	// --------------------------------------------------------

	responseCredencial, err := http.Get(fmt.Sprintf("%s/credencial?query=id_usuario:%d", urlCrudAuth, usuarios[0].Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar la credencial"}
		c.ServeJSON()
		return
	}

	defer responseCredencial.Body.Close()

	bodyCredencial, _ := io.ReadAll(responseCredencial.Body)

	var credenciales []Credencial
	json.Unmarshal(bodyCredencial, &credenciales)

	hash, _ := bcrypt.GenerateFromPassword([]byte(datos.ContrasenaNueva), bcrypt.DefaultCost)

	credencial := Credencial{IdUsuario: &Ref{Id: usuarios[0].Id}, Algoritmo: "bcrypt", Activo: true}

	if len(credenciales) > 0 {
		credencial = credenciales[0]
	}

	credencial.ContrasenaHash = string(hash)
	credencial.Salt = string(hash[7:29])
	credencial.FechaUltimoCambio = ahora
	credencial.IntentosFallidos = 0

	jsonCredencial, _ := json.Marshal(credencial)

	// Si el usuario entraba solo con red social no tiene credencial: se crea
	metodo := http.MethodPost
	direccion := urlCrudAuth + "/credencial"

	if credencial.Id > 0 {
		metodo = http.MethodPut
		direccion = fmt.Sprintf("%s/credencial/%d", urlCrudAuth, credencial.Id)
	}

	peticion, _ := http.NewRequest(metodo, direccion, bytes.NewBuffer(jsonCredencial))
	peticion.Header.Set("Content-Type", "application/json")

	responseGuardar, err := http.DefaultClient.Do(peticion)

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo guardar la contrasena"}
		c.ServeJSON()
		return
	}

	responseGuardar.Body.Close()

	// --------------------------------------------------------
	// 5. Responder
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{"mensaje": "Contrasena actualizada"}
	c.ServeJSON()
}

// ============================================================
// PUT /v1/perfil/:id/contrasena
// Cambia la contrasena desde el perfil (pide la actual)
// ============================================================

func (c *AuthController) PutCambiarContrasena() {

	// --------------------------------------------------------
	// 1. Leer el ID y los datos enviados
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	var datos DatosContrasena

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	if len(datos.ContrasenaNueva) < 8 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "La contrasena nueva debe tener minimo 8 caracteres"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Buscar la credencial del usuario
	// --------------------------------------------------------

	responseCredencial, err := http.Get(fmt.Sprintf("%s/credencial?query=id_usuario:%d", urlCrudAuth, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar la credencial"}
		c.ServeJSON()
		return
	}

	defer responseCredencial.Body.Close()

	bodyCredencial, _ := io.ReadAll(responseCredencial.Body)

	var credenciales []Credencial
	json.Unmarshal(bodyCredencial, &credenciales)

	if len(credenciales) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no tiene contrasena"}
		c.ServeJSON()
		return
	}

	credencial := credenciales[0]

	// --------------------------------------------------------
	// 3. Revisar que la contrasena actual sea correcta
	// --------------------------------------------------------

	if bcrypt.CompareHashAndPassword([]byte(credencial.ContrasenaHash), []byte(datos.ContrasenaActual)) != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "La contrasena actual no es correcta"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Guardar la contrasena nueva cifrada
	// --------------------------------------------------------

	hash, _ := bcrypt.GenerateFromPassword([]byte(datos.ContrasenaNueva), bcrypt.DefaultCost)

	credencial.ContrasenaHash = string(hash)
	credencial.Salt = string(hash[7:29])
	credencial.FechaUltimoCambio = time.Now().UTC()

	jsonCredencial, _ := json.Marshal(credencial)

	peticionPut, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/credencial/%d", urlCrudAuth, credencial.Id), bytes.NewBuffer(jsonCredencial))
	peticionPut.Header.Set("Content-Type", "application/json")

	responsePut, err := http.DefaultClient.Do(peticionPut)

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo guardar la contrasena"}
		c.ServeJSON()
		return
	}

	responsePut.Body.Close()

	// --------------------------------------------------------
	// 5. Responder
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{"mensaje": "Contrasena actualizada"}
	c.ServeJSON()
}
