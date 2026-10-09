package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

// Ref es una llave foranea tal como llega de los CRUD: {"Id": 1}
type Ref struct {
	Id int `json:"Id"`
}

type Usuario struct {
	Id                int       `json:"Id"`
	TipoDocumento     *Ref      `json:"TipoDocumento"`
	Nombre            string    `json:"Nombre"`
	Apellido          string    `json:"Apellido"`
	Email             string    `json:"Email"`
	Telefono          string    `json:"Telefono"`
	Cedula            string    `json:"Cedula"`
	Ciudad            string    `json:"Ciudad"`
	Direccion         string    `json:"Direccion"`
	FechaNacimiento   time.Time `json:"FechaNacimiento"`
	Estado            string    `json:"Estado"`
	FechaRegistro     time.Time `json:"FechaRegistro"`
	FechaUltimaSesion time.Time `json:"FechaUltimaSesion"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
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
}

type DatosLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
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

	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", id))

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

		responseTipo, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/tipo_documento/%d", usuario.TipoDocumento.Id))

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

	responseRoles, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario_rol?query=id_usuario:%d,activo:true&limit=1000", id))

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

		responseRol, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/rol/%d", asignacion.IdRol.Id))

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

	// --------------------------------------------------------
	// 3. Revisar que el correo no este registrado
	// --------------------------------------------------------

	responseRepetido, err := http.Get("http://localhost:8081/v1/usuario?query=email:" + url.QueryEscape(datos.Email))

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

	responseRol, err := http.Get("http://localhost:8081/v1/rol?query=nombre_rol:Usuario")

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
		Nombre:    datos.Nombre,
		Apellido:  datos.Apellido,
		Email:     datos.Email,
		Cedula:    datos.Cedula,
		Telefono:  datos.Telefono,
		Ciudad:    datos.Ciudad,
		Direccion: datos.Direccion,
		Estado:    "activo",
		Activo:    true,
	}

	if datos.IdTipoDocumento > 0 {
		nuevoUsuario.TipoDocumento = &Ref{Id: datos.IdTipoDocumento}
	}

	jsonUsuario, _ := json.Marshal(nuevoUsuario)

	responseCrear, err := http.Post("http://localhost:8081/v1/usuario", "application/json", bytes.NewBuffer(jsonUsuario))

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

	responseCredencial, err := http.Post("http://localhost:8081/v1/credencial", "application/json", bytes.NewBuffer(jsonCredencial))

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

	responseAsignacion, err := http.Post("http://localhost:8081/v1/usuario_rol", "application/json", bytes.NewBuffer(jsonAsignacion))

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
// Valida correo y clave. Bloquea 15 minutos tras 5 intentos.
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

	responseUsuario, err := http.Get("http://localhost:8081/v1/usuario?query=email:" + url.QueryEscape(email))

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

	responseCredencial, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/credencial?query=id_usuario:%d", usuario.Id))

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

	peticionPut, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8081/v1/credencial/%d", credencial.Id), bytes.NewBuffer(jsonCredencial))
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

	if responseAuditoria, err := http.Post("http://localhost:8081/v1/auditoria_login", "application/json", bytes.NewBuffer(jsonAuditoria)); err == nil {
		responseAuditoria.Body.Close()
	}

	// --------------------------------------------------------
	// 8. Responder
	// --------------------------------------------------------

	if !claveCorrecta {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"error": "Correo o contrasena incorrectos"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"mensaje":    "Inicio de sesion exitoso",
		"id_usuario": usuario.Id,
		"nombre":     usuario.Nombre,
		"apellido":   usuario.Apellido,
		"email":      usuario.Email,
	}

	c.ServeJSON()
}
