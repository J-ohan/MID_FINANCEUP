package controllers

import (
	"api_mid_financeup/helpers"
)

// PasarelaController reenvia las operaciones simples (listar, crear, editar,
// borrar) a un CRUD, para que el frontend solo tenga que conocer al MID.
//
//	GET    /v1/crud/finanzas/categoria          -> GET    http://localhost:8082/v1/categoria
//	POST   /v1/crud/finanzas/meta               -> POST   http://localhost:8082/v1/meta
//	PUT    /v1/crud/soporte/pqr/3               -> PUT    http://localhost:8085/v1/pqr/3
type PasarelaController struct {
	BaseController
}

// Reenviar atiende /v1/crud/:conjunto/* con cualquier metodo.
func (c *PasarelaController) Reenviar() {
	conjunto := c.Ctx.Input.Param(":conjunto")
	if !helpers.ExisteConjunto(conjunto) {
		c.fallo(helpers.NuevoError(404, "El conjunto '"+conjunto+"' no existe. Usa: auth, finanzas, educacion, negocio o soporte"))
		return
	}

	status, cuerpo, err := helpers.Reenviar(
		c.Ctx.Input.Method(),
		conjunto,
		c.Ctx.Input.Param(":splat"),
		c.Ctx.Request.URL.RawQuery,
		c.Ctx.Input.RequestBody,
	)
	if err != nil {
		c.fallo(err)
		return
	}

	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Ctx.Output.SetStatus(status)
	c.Ctx.Output.Body(cuerpo)
}
