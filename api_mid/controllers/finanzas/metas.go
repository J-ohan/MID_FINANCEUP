package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
	servicio "api_mid_financeup/services/finanzas"
)

// GetMetas atiende GET /v1/metas/usuario/:id (pantalla Metas).
//
// Crear, editar o borrar una meta no pasa por el MID:
// la pagina lo hace directo en crud_finanzas (puerto 8082).
func (c *FinanzasController) GetMetas() {

	// 1. Leer el id del usuario
	id, err := helpers.LeerId(&c.Controller)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 2. Pedir las metas con su porcentaje al service
	resultado, err := servicio.MetasUsuario(id)
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// PutAbonoMeta atiende PUT /v1/metas/:id/abono (boton "Abonar").
// Body de ejemplo: {"valor": 50000}
func (c *FinanzasController) PutAbonoMeta() {

	// 1. Leer el id de la meta (de la URL) y el valor del abono (del body)
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

	// 2. El service suma el abono a la meta
	resultado, err := servicio.AbonarMeta(id, datos.Valor, helpers.DuenoQuePide(&c.Controller))
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	// 3. Responder como quedo la meta
	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
