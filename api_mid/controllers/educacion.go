package controllers

import (
	"net/http"
	"strconv"

	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// EducacionController atiende las rutas de educacion.
type EducacionController struct {
	beego.Controller
}

// GET /v1/progreso-educativo/:id
func (c *EducacionController) GetProgresoEducativo() {

	// 1. Leer el id de la URL
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// 2. Llamar al service
	resultado, codigo, err := services.ProgresoEducativo(id)

	// 3. Responder
	c.Ctx.ResponseWriter.WriteHeader(codigo)

	if err != nil {
		c.Data["json"] = map[string]interface{}{"error": err.Error()}
	} else {
		c.Data["json"] = resultado
	}

	c.ServeJSON()
}
