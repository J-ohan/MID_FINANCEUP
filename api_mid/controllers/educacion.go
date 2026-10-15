package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// EducacionController atiende las rutas de educacion.
type EducacionController struct {
	beego.Controller
}

// GET /v1/cursos
func (c *EducacionController) GetCursos() {

	resultado, err := services.Catalogo()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/cursos/:id?usuario=5
func (c *EducacionController) GetCurso() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	idUsuario, _ := c.GetInt("usuario") // opcional

	resultado, err := services.DetalleCurso(id, idUsuario)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/progreso-educativo/:id
func (c *EducacionController) GetProgresoEducativo() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ProgresoEducativo(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/clase/:id
func (c *EducacionController) GetClase() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.Clase(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/clase/:id/completar
func (c *EducacionController) PostCompletarLeccion() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosUsuarioCurso
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CompletarLeccion(id, datos.IdUsuario)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/cursos/:id/reiniciar
func (c *EducacionController) PostReiniciarCurso() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosUsuarioCurso
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ReiniciarCurso(id, datos.IdUsuario)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/certificado
func (c *EducacionController) PostCertificado() {

	var datos models.DatosUsuarioCurso
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EmitirCertificado(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
