package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// AuthController atiende las rutas de usuarios.
// Cada funcion hace 3 cosas:
//  1. leer lo que llega (id o JSON)
//  2. llamar al service que hace el trabajo
//  3. responder (el resultado o el error)
type AuthController struct {
	beego.Controller
}

// GET /v1/perfil/:id
func (c *AuthController) GetPerfil() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	perfil, err := services.ObtenerPerfil(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, perfil)
}

// POST /v1/registro
func (c *AuthController) PostRegistro() {

	var datos models.DatosRegistro
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.Registrar(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// POST /v1/login  (paso 1: correo y clave)
func (c *AuthController) PostLogin() {

	var datos models.DatosLogin
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// Para la auditoria
	datos.Ip = c.Ctx.Input.IP()
	datos.Navegador = c.Ctx.Input.UserAgent()

	resultado, err := services.IniciarSesion(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/login/verificar  (paso 2: codigo de 6 numeros)
func (c *AuthController) PostVerificarCodigo() {

	var datos models.DatosCodigo
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.VerificarCodigo(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/login/social
func (c *AuthController) PostLoginSocial() {

	var datos models.DatosLoginSocial
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.IniciarSesionSocial(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/recuperar-contrasena  (pide el codigo)
func (c *AuthController) PostRecuperarContrasena() {

	var datos models.DatosContrasena
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.PedirCodigoRecuperacion(datos.Email)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// PUT /v1/recuperar-contrasena  (codigo + clave nueva)
func (c *AuthController) PutRestablecerContrasena() {

	var datos models.DatosContrasena
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.RestablecerContrasena(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// PUT /v1/perfil/:id/contrasena
func (c *AuthController) PutCambiarContrasena() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosContrasena
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CambiarContrasena(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
