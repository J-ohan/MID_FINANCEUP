package helpers

import (
	"errors"

	beego "github.com/beego/beego/v2/server/web"
)

// Respuesta es el formato unico con el que el MID responde al frontend:
//
//	{ "exito": true,  "datos": {...} }
//	{ "exito": false, "mensaje": "El correo es obligatorio" }
type Respuesta struct {
	Exito   bool        `json:"exito"`
	Mensaje string      `json:"mensaje,omitempty"`
	Datos   interface{} `json:"datos,omitempty"`
}

// ResponderDatos envia una respuesta exitosa con datos.
func ResponderDatos(c *beego.Controller, status int, datos interface{}) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = Respuesta{Exito: true, Datos: datos}
	c.ServeJSON()
}

// ResponderError envia un error. Si es un ErrorMid usa su codigo; si no, 500.
func ResponderError(c *beego.Controller, err error) {
	status := 500
	mensaje := "Ocurrio un error inesperado: " + err.Error()

	var errorMid *ErrorMid
	if errors.As(err, &errorMid) {
		status = errorMid.Status
		mensaje = errorMid.Mensaje
	}

	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = Respuesta{Exito: false, Mensaje: mensaje}
	c.ServeJSON()
}
