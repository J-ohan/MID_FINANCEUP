package controllers

import (
	"net/http"

	"api_mid/models"
	"api_mid/services"

	beego "github.com/beego/beego/v2/server/web"
)


type AuthController struct{
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



//Post v1/registro

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


//Post v1/login (correo y clave)

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


//Post v1/login/verificar (codigo de 6 digitos)

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

func 
