package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// POST /v1/verificar-correo  (envia el codigo)  body: {"email": "ana@correo.com"}
func (c *AuthController) PostVerificarCorreo() {

	var datos models.DatosVerificarCorreo
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EnviarVerificacionCorreo(datos.Email)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// PUT /v1/verificar-correo  (confirma el codigo)  body: {"email": "ana@correo.com", "codigo": "123456"}
func (c *AuthController) PutVerificarCorreo() {

	var datos models.DatosVerificarCorreo
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ConfirmarVerificacionCorreo(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
