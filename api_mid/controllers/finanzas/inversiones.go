package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	servicio "api_mid_financeup/services/finanzas"
)

// GetInversiones atiende GET /v1/inversiones/usuario/:id (pantalla Inversiones).
//
// Crear, editar o borrar una inversion no pasa por el MID:
// la pagina lo hace directo en crud_finanzas (puerto 8082).
func (c *FinanzasController) GetInversiones() {

	// 1. Leer el id del usuario
	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. Pedir las inversiones con su balance y grafica al service
	resultado, err := servicio.InversionesUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
