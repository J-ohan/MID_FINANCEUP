package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

// Referencia es una llave foranea tal como llega de los CRUD: {"Id": 1}
type Referencia struct {
	Id int `json:"Id"`
}

type ModuloEducativo struct {
	Id     int    `json:"Id"`
	Titulo string `json:"Titulo"`
	Nivel  string `json:"Nivel"`
}

type ProgresoEducativo struct {
	IdModulo             *Referencia `json:"IdModulo"`
	PorcentajeCompletado int         `json:"PorcentajeCompletado"`
	Calificacion         int         `json:"Calificacion"`
}

// EducacionController operations for Educacion
type EducacionController struct {
	beego.Controller
}

// ============================================================
// GET /v1/progreso-educativo/:id
// Avance del usuario en cada curso (crud_auth + crud_educacion)
// ============================================================

func (c *EducacionController) GetProgresoEducativo() {

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
	// 2. Revisar que el usuario exista (crud_auth)
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

	var usuario struct {
		Id int `json:"Id"`
	}

	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar todos los cursos (modulos)
	// --------------------------------------------------------

	responseModulos, err := http.Get("http://localhost:8083/v1/modulo_educativo?query=activo:true&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseModulos.Body.Close()

	bodyModulos, _ := io.ReadAll(responseModulos.Body)

	var modulos []ModuloEducativo
	json.Unmarshal(bodyModulos, &modulos)

	// --------------------------------------------------------
	// 4. Consultar el avance del usuario en cada curso
	// --------------------------------------------------------

	responseProgreso, err := http.Get(fmt.Sprintf("http://localhost:8083/v1/progreso_educativo?query=id_usuario:%d&limit=1000", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseProgreso.Body.Close()

	bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

	var avances []ProgresoEducativo
	json.Unmarshal(bodyProgreso, &avances)

	// Guardamos el avance de cada curso por su id
	avancePorModulo := map[int]ProgresoEducativo{}
	for _, avance := range avances {
		if avance.IdModulo != nil {
			avancePorModulo[avance.IdModulo.Id] = avance
		}
	}

	// --------------------------------------------------------
	// 5. Unir cada curso con su porcentaje y estado
	// --------------------------------------------------------

	listaModulos := []map[string]interface{}{}

	for _, modulo := range modulos {

		porcentaje := avancePorModulo[modulo.Id].PorcentajeCompletado

		estado := "sin iniciar"
		if porcentaje >= 100 {
			estado = "completado"
		} else if porcentaje > 0 {
			estado = "en curso"
		}

		listaModulos = append(listaModulos, map[string]interface{}{
			"id_modulo":  modulo.Id,
			"titulo":     modulo.Titulo,
			"nivel":      modulo.Nivel,
			"porcentaje": porcentaje,
			"estado":     estado,
		})
	}

	// --------------------------------------------------------
	// 6. Devolver el progreso
	//    Pendiente: porcentaje general y cursos completados
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_usuario": usuario.Id,
		"modulos":    listaModulos,
	}

	c.ServeJSON()
}
