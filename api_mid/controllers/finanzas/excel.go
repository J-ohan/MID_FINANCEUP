package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	servicio "api_mid_financeup/services/finanzas"
)

// PostImportarMovimientos atiende POST /v1/movimientos/importar
// (boton "Importar Excel" del Libro mayor).
//
// Body de ejemplo:
//
//	{"id_usuario": 1, "modo": "validar", "movimientos": [
//	   {"fecha": "2026-10-01", "concepto": "Sueldo", "tipo": "ingreso", "valor": 3000000}]}
func (c *FinanzasController) PostImportarMovimientos() {

	// 1. Leer el JSON que manda la pagina
	var datos models.DatosImportar
	if err := helpers.LeerBody(&c.Controller, &datos); err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. El service revisa (y si el modo es "confirmar", guarda) las filas
	resultado, err := servicio.ImportarMovimientos(datos)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder cuantas filas estaban bien, cuantas se guardaron y cuales fallaron
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GetExportarMovimientos atiende GET /v1/movimientos/usuario/:id/exportar
// (boton "Exportar" del Libro mayor). Devuelve los datos de la plantilla
// de Excel; la pagina es la que arma el archivo .xlsx.
func (c *FinanzasController) GetExportarMovimientos() {

	// 1. Leer el id del usuario
	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. Pedir los datos de la plantilla, con los mismos filtros del libro mayor
	resultado, err := servicio.ExportarMovimientos(id, c.leerFiltros())
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder los datos
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
