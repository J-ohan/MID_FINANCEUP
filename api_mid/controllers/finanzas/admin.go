package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
	servicio "api_mid_financeup/services/finanzas"
)

// GetAdminFinanzas atiende GET /v1/admin/finanzas
// (pantallas del administrador: movimientos, metas e inversiones de todos).
func (c *FinanzasController) GetAdminFinanzas() {

	// Esta ruta no recibe nada: se llama directo al service
	resultado, err := servicio.FinanzasAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}

// GetAdminLibroMayor atiende GET /v1/admin/libro-mayor
// (pantalla admin/libro-mayor: un libro por cada usuario).
func (c *FinanzasController) GetAdminLibroMayor() {

	resultado, err := servicio.LibroMayorAdmin()
	if err != nil {
		helpers.ResponderError(&c.Controller, err)
		return
	}

	helpers.Responder(&c.Controller, http.StatusOK, resultado)
}
