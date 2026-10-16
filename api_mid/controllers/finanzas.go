package controllers

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	"api_mid_financeup/services"

	beego "github.com/beego/beego/v2/server/web"
)

// FinanzasController atiende las rutas de finanzas.
type FinanzasController struct {
	beego.Controller
}


// GET /v1/resumen-financiero/:id
func (c *FinanzasController) GetResumenFinanciero() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}


	resultado, err := services.ResumenFinanciero(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/movimientos/usuario/:id?tipo=ingreso|gasto&categoria=3
func (c *FinanzasController) GetMovimientos() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// Filtros opcionales que llegan en la URL
	filtros := models.FiltrosMovimientos{Tipo: c.GetString("tipo")}
	filtros.IdCategoria, _ = c.GetInt("categoria")

	resultado, err := services.LibroMayor(id, filtros)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/movimientos/importar
func (c *FinanzasController) PostImportarMovimientos() {

	var datos models.DatosImportar
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ImportarMovimientos(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/metas/usuario/:id
func (c *FinanzasController) GetMetas() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.MetasUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)

}

// PUT /v1/metas/:id/abono
func (c *FinanzasController) PutAbonoMeta() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosAbono
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.AbonarMeta(id, datos.Valor)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}



// GET /v1/inversiones/usuario/:id
func (c *FinanzasController) GetInversiones() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.InversionesUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

