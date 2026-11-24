package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// NegocioController atiende las rutas de alianzas.
type NegocioController struct {
	beego.Controller
}

// GET /v1/ofertas
func (c *NegocioController) GetOfertas() {

	resultado, err := services.Ofertas()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/ofertas/:id
func (c *NegocioController) GetOferta() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.FichaOferta(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/solicitud-credito
func (c *NegocioController) PostSolicitudCredito() {

	var datos models.DatosSolicitudCredito
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.RadicarSolicitud(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// GET /v1/solicitud-credito/usuario/:id
func (c *NegocioController) GetSolicitudesUsuario() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.SolicitudesUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/admin/solicitudes-credito
func (c *NegocioController) GetSolicitudesAdmin() {

	resultado, err := services.SolicitudesAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
