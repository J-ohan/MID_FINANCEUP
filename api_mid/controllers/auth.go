package controllers

import (
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// AuthController atiende perfil, registro e inicio de sesion (usa crud_auth).
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

// Registrar POST /v1/registro
func (c *AuthController) Registrar() {
	var datos models.DatosRegistro
	if err := c.leerCuerpo(&datos); err != nil {
		c.fallo(err)
		return
	}
	perfil, err := services.Registrar(datos)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(201, perfil)
}

// IniciarSesion POST /v1/login
func (c *AuthController) IniciarSesion() {
	var datos models.DatosLogin
	if err := c.leerCuerpo(&datos); err != nil {
		c.fallo(err)
		return
	}
	perfil, err := services.IniciarSesion(datos, c.Ctx.Input.IP(), c.Ctx.Input.UserAgent())
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, perfil)
}
