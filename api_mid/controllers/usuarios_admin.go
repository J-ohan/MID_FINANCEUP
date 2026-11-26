package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// GET /v1/admin/usuarios  (todos los usuarios con su rol)
func (c *AuthController) GetAdminUsuarios() {

	usuarios, err := services.ListarUsuariosAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, usuarios)
}

// POST /v1/admin/usuarios  body: {"nombre", "apellido", "email", "contrasena", "cedula", "telefono", "rol"}
func (c *AuthController) PostAdminUsuario() {

	var datos models.DatosUsuarioAdmin
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CrearUsuarioAdmin(helpers.IdUsuarioDeLaSesion(&c.Controller), datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// PUT /v1/admin/usuarios/:id  body: {"nombre", "apellido", "cedula", "telefono"}
func (c *AuthController) PutAdminUsuario() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosUsuarioAdmin
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EditarUsuarioAdmin(helpers.IdUsuarioDeLaSesion(&c.Controller), id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// PUT /v1/admin/usuarios/:id/rol  body: {"rol": "admin"}
func (c *AuthController) PutRolUsuario() {
	c.cambiarUsuario(services.CambiarRolUsuario)
}

// PUT /v1/admin/usuarios/:id/estado  body: {"estado": "suspendido", "motivo": "..."}
func (c *AuthController) PutEstadoUsuario() {
	c.cambiarUsuario(services.CambiarEstadoUsuario)
}

// cambiarUsuario lee el :id y el body, y llama al service con el id del admin que lo pide.
func (c *AuthController) cambiarUsuario(cambio func(int, int, models.DatosAdminUsuario) (map[string]interface{}, error)) {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosAdminUsuario
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := cambio(helpers.IdUsuarioDeLaSesion(&c.Controller), id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
