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
