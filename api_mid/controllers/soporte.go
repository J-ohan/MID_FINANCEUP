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
	IdEstado      *Ref      `json:"IdEstado"`
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

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responseEstado.Body.Close()

	bodyEstado, _ := io.ReadAll(responseEstado.Body)

	var estados []EstadoPqr
	json.Unmarshal(bodyEstado, &estados)

	if len(estados) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"error": "No existe el estado Abierta en la base de datos"}
		c.ServeJSON()
		return
	}

	pqr := Pqr{
		IdUsuario:   datos.IdUsuario,
		Radicado:    "PQR-" + time.Now().Format("20060102-150405"),
		Titulo:      datos.Titulo,
		Tipo:        datos.Tipo,
		Categoria:   datos.Categoria,
		Prioridad:   datos.Prioridad,
		Descripcion: datos.Descripcion,
		IdEstado:    &Ref{Id: estados[0].Id},
		Activo:      true,
	}

	jsonPqr, _ := json.Marshal(pqr)

	responseCrear, err := http.Post("http://localhost:8085/v1/pqr", "application/json", bytes.NewBuffer(jsonPqr))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responseCrear.Body.Close()

	bodyCrear, _ := io.ReadAll(responseCrear.Body)

	if responseCrear.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo crear la PQR: " + string(bodyCrear)}
		c.ServeJSON()
		return
	}

	json.Unmarshal(bodyCrear, &pqr)

	actividad := RegistroActividad{
		IdUsuario:       &datos.IdUsuario,
		TipoActividad:   "CREAR_PQR",
		Descripcion:     "Se radico la PQR " + pqr.Radicado,
		EntidadAfectada: "soporte.pqr",
	}

	jsonActividad, _ := json.Marshal(actividad)

	if responseActividad, err := http.Post("http://localhost:8085/v1/registro_actividad", "application/json", bytes.NewBuffer(jsonActividad)); err == nil {
		responseActividad.Body.Close()
	}


	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"id_pqr":    pqr.Id,
		"radicado":  pqr.Radicado,
		"titulo":    pqr.Titulo,
		"tipo":      pqr.Tipo,
		"prioridad": pqr.Prioridad,
		"estado":    estados[0].Nombre,
	}

	c.ServeJSON()
}


func (c *SoporteController) GetPqrUsuario() {


	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}


	responsePqrs, err := http.Get(fmt.Sprintf("http://localhost:8085/v1/pqr?query=id_usuario:%d,activo:true&limit=1000", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responsePqrs.Body.Close()

	bodyPqrs, _ := io.ReadAll(responsePqrs.Body)

	var pqrs []Pqr
	json.Unmarshal(bodyPqrs, &pqrs)


	responseEstados, err := http.Get("http://localhost:8085/v1/estado_pqr?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los estados"}
		c.ServeJSON()
		return
	}

	defer responseEstados.Body.Close()

	bodyEstados, _ := io.ReadAll(responseEstados.Body)

	var estados []EstadoPqr
	json.Unmarshal(bodyEstados, &estados)

	nombreEstado := map[int]string{}
	for _, estado := range estados {
		nombreEstado[estado.Id] = estado.Nombre
	}

	lista :=[]map[string]interface{}{}

	for _, pqr := range pqrs {
		estado := ""
		if pqr.IdEstado !=nil{
			estado = nombreEstado[pqr.IdEstado.Id]
		}

		lista = append(lista, map[string]interface{}{
			"id_pqr": pqr.Id,
			"radicado": pqr.Radicado,
			"titulo": pqr.Titulo,
			"categoria" : pqr.Categoria,
			"prioridad": pqr.Prioridad,
			"estado": estado,
			"descripcion": pqr.Descripcion,
			"fecha_creacion": pqr.FechaCreacion,
		})
	}

	c.Data["json"] = lista
	c.ServeJSON()
}