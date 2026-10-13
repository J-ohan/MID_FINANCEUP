package controllers

import (
	"encoding/json"
	"strconv"

	"api_mid_financeup/helpers"

	beego "github.com/beego/beego/v2/server/web"
)

// BaseController tiene las funciones que comparten todos los controladores del MID.
type BaseController struct {
	beego.Controller
}

// idDeLaRuta lee un numero de la URL, por ejemplo el 5 de /v1/perfil/5.
func (c *BaseController) idDeLaRuta(nombre string) (int, error) {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":" + nombre))
	if err != nil || id <= 0 {
		return 0, helpers.NuevoError(400, "El "+nombre+" de la URL debe ser un numero mayor que cero")
	}
	return id, nil
}

// leerCuerpo convierte el JSON enviado por el frontend en una estructura de Go.
func (c *BaseController) leerCuerpo(destino interface{}) error {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, destino); err != nil {
		return helpers.NuevoError(400, "El cuerpo de la peticion no es un JSON valido")
	}
	return nil
}