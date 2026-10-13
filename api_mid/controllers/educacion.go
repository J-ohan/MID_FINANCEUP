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
	// 4. Pendiente: consultar progreso_educativo del usuario
	//    y unir cada curso con su porcentaje y estado
	// --------------------------------------------------------

	// --------------------------------------------------------
	// 5. Por ahora se devuelven los cursos sin el avance
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_usuario": usuario.Id,
		"modulos":    modulos,
	}

	c.ServeJSON()
}
