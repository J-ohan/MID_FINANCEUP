package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// GET /v1/admin/aliados  (bancos para el selector del formulario)
func (c *NegocioController) GetAliadosAdmin() {

	resultado, err := services.AliadosAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/admin/ofertas  body: {"id_aliado", "categoria", "familia", "tasa_desde", "monto_minimo", "monto_maximo", ...}
func (c *NegocioController) PostOfertaAdmin() {

	var datos models.DatosOfertaAdmin
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CrearOferta(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// PUT /v1/admin/ofertas/:id  (mismo body que el POST)
func (c *NegocioController) PutOfertaAdmin() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosOfertaAdmin
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EditarOferta(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// DELETE /v1/admin/ofertas/:id  (no borra: la deja inactiva)
func (c *NegocioController) DeleteOfertaAdmin() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.RetirarOferta(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
