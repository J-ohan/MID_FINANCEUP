package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
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


// GET /v1/perfil/:id
// Usuario + tipo de documento + roles


func (c *AuthController) GetPerfil() {



	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{
			"error": "El ID debe ser numerico" ,
		}
		c.ServeJSON()
		return
	}


	responseUsuario, err := http.Get(fmt.Sprintf("%s/usuario/%d", urlCrudAuth, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{
			"error": "No fue posible comunicarse con la API de usuarios",
		}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario


	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{
			"error": "El usuario no existe",
		}
		c.ServeJSON()
		return
	}

	

	//Consultar su tipo de documento
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

	//Consultar roles

	responseRoles, err := http.Get(fmt.Sprintf("%s/usuario_rol?query=id_usuario:%d,activo:true&limit=1000", urlCrudAuth, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{
			"error": "No fue posible consultar los roles",
		}
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

	//construir el peil y devolverlo
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


// POST /v1/registro
// Crea usuario, clave cifrada y rol 


func (c *AuthController) PostRegistro(){

	var datos DatosRegistro

		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{
			"error": "El cuerpo de la peticion no es un JSON valido",
		}
		c.ServeJSON()
		return
	}
	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))
}