package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
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

	// Obtener el ID enviado en la URL

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}


	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", id),)

	if err !=nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"]= map[string]interface{}{
			"error": "No fue posible comunicarse con la API del usuarios",
		}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario,_:io.ReadAll(responseUsuario.Body)

	var usuario Usuario
}