package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// GeneralController atiende las rutas generales.
type GeneralController struct {
	beego.Controller
}

// GET /v1/estado
func (c *GeneralController) Estado() {
	helpers.Responder(&c.Controller, http.StatusOK, services.EstadoCruds())
}

// GET /v1/dashboard/:id
func (c *GeneralController) GetDashboard() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.Dashboard(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
