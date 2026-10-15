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


// leerFiltros lee los filtros opcionales del libro mayor de la URL:
// ?tipo=ingreso|gasto&categoria=3&desde=2026-01-01&hasta=2026-12-31&q=texto
func (c *FinanzasController) leerFiltros() models.FiltrosMovimientos {

	filtros := models.FiltrosMovimientos{
		Tipo:  c.GetString("tipo"),
		Desde: c.GetString("desde"),
		Hasta: c.GetString("hasta"),
		Texto: c.GetString("q"),
	}

	filtros.IdCategoria, _ = c.GetInt("categoria")

	return filtros
}

// ------------------------------------------------------------
// Panel principal
// ------------------------------------------------------------

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

// GET /v1/categorias
func (c *FinanzasController) GetCategorias() {

	resultado, err := services.Categorias()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// ------------------------------------------------------------
// Libro mayor
// ------------------------------------------------------------

// GET /v1/movimientos/usuario/:id?tipo=&categoria=&desde=&hasta=&q=
func (c *FinanzasController) GetMovimientos() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.LibroMayor(id, c.leerFiltros())
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// POST /v1/movimientos
func (c *FinanzasController) PostMovimiento() {

	var datos models.DatosMovimiento
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CrearMovimiento(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// PUT /v1/movimientos/:id
func (c *FinanzasController) PutMovimiento() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosMovimiento
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EditarMovimiento(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// DELETE /v1/movimientos/:id?id_usuario=3
func (c *FinanzasController) DeleteMovimiento() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	idUsuario, _ := c.GetInt("id_usuario")

	resultado, err := services.EliminarMovimiento(id, idUsuario)
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

// GET /v1/movimientos/usuario/:id/exportar?tipo=&categoria=&desde=&hasta=&q=
func (c *FinanzasController) GetExportarMovimientos() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.ExportarMovimientos(id, c.leerFiltros())
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// ------------------------------------------------------------
// Metas
// ------------------------------------------------------------

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

// POST /v1/metas
func (c *FinanzasController) PostMeta() {

	var datos models.DatosMeta
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CrearMeta(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// PUT /v1/metas/:id
func (c *FinanzasController) PutMeta() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosMeta
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EditarMeta(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// DELETE /v1/metas/:id?id_usuario=3
func (c *FinanzasController) DeleteMeta() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	idUsuario, _ := c.GetInt("id_usuario")

	resultado, err := services.EliminarMeta(id, idUsuario)
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

// ------------------------------------------------------------
// Inversiones
// ------------------------------------------------------------

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

// POST /v1/inversiones
func (c *FinanzasController) PostInversion() {

	var datos models.DatosInversion
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.CrearInversion(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusCreated, resultado)
}

// PUT /v1/inversiones/:id
func (c *FinanzasController) PutInversion() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	var datos models.DatosInversion
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	resultado, err := services.EditarInversion(id, datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// DELETE /v1/inversiones/:id?id_usuario=3
func (c *FinanzasController) DeleteInversion() {

	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	idUsuario, _ := c.GetInt("id_usuario")

	resultado, err := services.EliminarInversion(id, idUsuario)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// ------------------------------------------------------------
// Admin
// ------------------------------------------------------------

// GET /v1/admin/finanzas
func (c *FinanzasController) GetAdminFinanzas() {

	resultado, err := services.FinanzasAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GET /v1/admin/libro-mayor
func (c *FinanzasController) GetAdminLibroMayor() {

	resultado, err := services.LibroMayorAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
