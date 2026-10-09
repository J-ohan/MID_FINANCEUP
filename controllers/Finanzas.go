package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type Categoria struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Movimiento struct {
	Id          int       `json:"Id"`
	IdCategoria *Ref      `json:"IdCategoria"`
	Nombre      string    `json:"Nombre"`
	Monto       float64   `json:"Monto"`
	EsIngreso   bool      `json:"EsIngreso"`
	Fecha       time.Time `json:"Fecha"`
}

type Meta struct {
	Id            int     `json:"Id"`
	Nombre        string  `json:"Nombre"`
	MontoObjetivo float64 `json:"MontoObjetivo"`
	MontoActual   float64 `json:"MontoActual"`
	Icono         string  `json:"Icono"`
}

type Inversion struct {
	Id           int     `json:"Id"`
	Nombre       string  `json:"Nombre"`
	Monto        float64 `json:"Monto"`
	Rentabilidad float64 `json:"Rentabilidad"`
}

type SolicitudConsolidacion struct {
	Id             int     `json:"Id"`
	SaldoTotal     float64 `json:"SaldoTotal"`
	CuotaActual    float64 `json:"CuotaActual"`
	CuotaPropuesta float64 `json:"CuotaPropuesta"`
	Estado         string  `json:"Estado"`
}

// FinanzasController operations for Finanzas
type FinanzasController struct {
	beego.Controller
}


// GET /v1/resumen-financiero/:id
// Ingresos, gastos, balance, metas, inversiones y deudas
// (crud_auth + crud_finanzas)


func (c *FinanzasController) GetResumenFinanciero() {


	// 1. Obtener el ID enviado en la URL
	

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

}
	// 2. Revisar que el usuario existe (crud_auth)

	respuestaUsuario, err := http.Get(fmt.sprintf("http://localhost:8081/v1/usuario%id",id))

	if err != nil{
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}
	