package controllers

import (
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// AuthController atiende perfil, registro e inicio de sesion.
type AuthController struct {
	BaseController
}

// ObtenerPerfil GET /v1/perfil/:id
func (c *AuthController) ObtenerPerfil() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	perfil, err := services.ObtenerPerfil(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, perfil)
}