// Package finanzas tiene los controladores del modulo de finanzas.
//
// Un controlador NO tiene logica: solo hace 3 pasos (seccion 6.6 de la guia):
//  1. Leer lo que llega (el :id de la URL, los filtros o el JSON del body).
//  2. Llamar a la funcion del service que hace el trabajo.
//  3. Responder a la pagina con el resultado o con el error.
//
// Toda la logica esta en services/finanzas.
package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	servicio "api_mid_financeup/services/finanzas"

	beego "github.com/beego/beego/v2/server/web"
)

// FinanzasController atiende todas las rutas de finanzas.
// Sus funciones estan repartidas en varios archivos de esta carpeta,
// uno por pantalla (libro mayor, excel, metas, inversiones, admin).
type FinanzasController struct {
	beego.Controller
}

// leerFiltros lee los filtros opcionales del libro mayor que llegan en la URL:
// ?tipo=ingreso|gasto&categoria=3&desde=2026-01-01&hasta=2026-12-31&q=texto
// Si no llega ninguno, la lista sale completa.
func (c *FinanzasController) leerFiltros() models.FiltrosMovimientos {

	filtros := models.FiltrosMovimientos{
		Tipo:  c.GetString("tipo"),
		Desde: c.GetString("desde"),
		Hasta: c.GetString("hasta"),
		Texto: c.GetString("q"),
	}

	// "categoria" puede llegar como Id (3) o como nombre (Vivienda)
	if id, err := c.GetInt("categoria"); err == nil {
		filtros.IdCategoria = id
	} else {
		filtros.Categoria = c.GetString("categoria")
	}

	return filtros
}

// GetResumenFinanciero atiende GET /v1/resumen-financiero/:id
// (el panel principal de Finanzas).
func (c *FinanzasController) GetResumenFinanciero() {

	// 1. Leer el id del usuario de la URL
	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. Pedirle el resumen al service
	resultado, err := servicio.ResumenFinanciero(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder 200 con el resumen
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
