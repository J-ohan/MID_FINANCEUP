package controllers

import (
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// NegocioController atiende las rutas de alianzas.
type NegocioController struct {
	beego.Controller
}

// GET /v1/ofertas
func (c *NegocioController) GetOfertas() {

	// 1. Llamar al service
	resultado, codigo, err := services.Ofertas()

	// 2. Responder
	c.Ctx.ResponseWriter.WriteHeader(codigo)

	if err != nil {
		c.Data["json"] = map[string]interface{}{"error": err.Error()}
	} else {
		c.Data["json"] = resultado
	}

	c.ServeJSON()
}
