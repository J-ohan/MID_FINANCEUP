package helpers

import (
	"encoding/json"
	"net/http"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

// Funciones que usan los controladores para leer lo que
// llega y para responder.

// LeerId lee el :id de la URL y revisa que sea un numero.
func LeerId(c *beego.Controller) (int, error) {

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		return 0, NuevoError(http.StatusBadRequest, "El ID debe ser numerico")
	}

	return id, nil
}

// LeerBody convierte el JSON que manda la pagina a una estructura.
func LeerBody(c *beego.Controller, destino interface{}) error {

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, destino); err != nil {
		return NuevoError(http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido")
	}

	return nil
}

// Responder devuelve el JSON con el codigo indicado (200, 201...).
func Responder(c *beego.Controller, codigo int, datos interface{}) {
	c.Ctx.ResponseWriter.WriteHeader(codigo)
	c.Data["json"] = datos
	c.ServeJSON()
}

// ResponderError devuelve {"error": "..."} con su codigo.
// Si el error no es un ErrorApi se responde 500.
func ResponderError(c *beego.Controller, err error) {

	codigo := http.StatusInternalServerError

	if errorApi, ok := err.(*ErrorApi); ok {
		codigo = errorApi.Codigo
	}

	Responder(c, codigo, map[string]interface{}{"error": err.Error()})
}
