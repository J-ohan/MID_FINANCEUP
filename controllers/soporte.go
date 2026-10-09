package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type EstadoPqr struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Pqr struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	Radicado      string    `json:"Radicado"`
	Titulo        string    `json:"Titulo"`
	Tipo          string    `json:"Tipo"`
	Categoria     string    `json:"Categoria"`
	Prioridad     string    `json:"Prioridad"`
	Descripcion   string    `json:"Descripcion"`
	IdEstado      int     `json:"IdEstado"`
	Activo        bool      `json:"Activo"`
	FechaCreacion time.Time `json:"FechaCreacion"`
}

type RegistroActividad struct {
	IdUsuario       *int   `json:"IdUsuario"`
	TipoActividad   string `json:"TipoActividad"`
	Descripcion     string `json:"Descripcion"`
	EntidadAfectada string `json:"EntidadAfectada"`
}

type DatosPqr struct {
	IdUsuario   int    `json:"id_usuario"`
	Titulo      string `json:"titulo"`
	Tipo        string `json:"tipo"`
	Categoria   string `json:"categoria"`
	Prioridad   string `json:"prioridad"`
	Descripcion string `json:"descripcion"`
}

// SoporteController operations for Soporte
type SoporteController struct {
	beego.Controller
}

func (c *SoporteController) PostPqr() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosPqr

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	if datos.Titulo == "" || datos.Descripcion == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El titulo y la descripcion son obligatorios"}
		c.ServeJSON()
		return
	}

	if datos.Tipo != "peticion" && datos.Tipo != "queja" && datos.Tipo != "reclamo" && datos.Tipo != "sugerencia" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El tipo debe ser: peticion, queja, reclamo o sugerencia"}
		c.ServeJSON()
		return
	}

	if datos.Prioridad == "" {
		datos.Prioridad = "media"
	}

	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario

	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe"}
		c.ServeJSON()
		return
	}

	responseEstado, err := http.Get("http://localhost:8085/v1/estado_pqr?query=nombre:Abierta")


	defer responseEstado.Body.Close()
}