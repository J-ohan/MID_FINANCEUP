package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	servicio "api_mid_financeup/services/finanzas"
)

// GetMovimientos atiende GET /v1/movimientos/usuario/:id
// (pantalla Libro mayor). Acepta los filtros de leerFiltros.
//
// Crear, editar o borrar un movimiento no pasa por el MID:
// la pagina lo hace directo en crud_finanzas (puerto 8082).
func (c *FinanzasController) GetMovimientos() {

	// 1. Leer el id del usuario
	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. Pedir los movimientos filtrados al service
	resultado, err := servicio.LibroMayor(id, c.leerFiltros())
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder la lista
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
