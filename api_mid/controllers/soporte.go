package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// SoporteController atiende las rutas de PQR.
type SoporteController struct {
	beego.Controller
}

// POST /v1/pqr
func (c *SoporteController) PostPqr() {

	var datos models.DatosPqr
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.RadicarPqr(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// GET /v1/pqr/usuario/:id
func (c *SoporteController) GetPqrUsuario() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.PqrUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/pqr/radicado/:radicado
func (c *SoporteController) GetPqr() {

	resultado, err := services.VerPqr(c.Ctx.Input.Param(":radicado"))
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/pqr/:id/respuesta
func (c *SoporteController) PostRespuestaPqr() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosRespuestaPqr
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ResponderPqr(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// GET /v1/admin/pqr
func (c *SoporteController) GetPqrAdmin() {

	resultado, err := services.PqrAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
