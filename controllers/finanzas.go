package controllers

import (
	"api_mid_financeup/models"
	"api_mid_financeup/services"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/beego/beego/v2/server/web"
)

// FinanzasController atiende el módulo de Finanzas:
// panel principal, Libro mayor, Metas, Inversiones y vistas de Admin.
// Las consultas a los CRUD están en services/crud_finanzas_service.go
// y los cálculos en services/calculos_finanzas_service.go.
type FinanzasController struct {
	web.Controller
}

// ============================================================
// FINANZAS (panel principal)
// ============================================================

// @Title Resumen financiero.
// @Description Devuelve las tarjetas, la salud financiera, la dona de gastos, los últimos movimientos, las metas y las inversiones de un usuario.
// @Param id path int true "Id del usuario"
// @Success 200 {object} models.RespuestaResumenFinanciero
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/resumen-financiero/:id [get]
func (c *FinanzasController) ObtenerResumenFinanciero() {
	idUsuario, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || idUsuario <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id del usuario debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	_, err = crud.ObtenerUsuario(idUsuario)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El usuario no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar el usuario.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	categorias, err := crud.ObtenerCategorias()

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar las categorías.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	movimientos, err := crud.ObtenerMovimientos(idUsuario)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar los movimientos.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	metas, err := crud.ObtenerMetas(idUsuario)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar las metas.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	inversiones, err := crud.ObtenerInversiones(idUsuario)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar las inversiones.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	niveles, err := crud.ObtenerNivelesRiesgo()

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar los niveles de riesgo.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	historial := crud.ObtenerHistorial()

	// Cálculos
	lista := services.ArmarMovimientos(movimientos, categorias)
	totales := services.CalcularTotales(lista)
	calculosMetas := services.CalcularMetas(metas)
	calculosInversiones := services.CalcularInversiones(inversiones, niveles, historial)

	respuesta := models.RespuestaResumenFinanciero{
		OK:                 true,
		Mensaje:            "Resumen financiero generado correctamente.",
		IdUsuario:          idUsuario,
		TotalIngresos:      totales.TotalIngresos,
		TotalGastos:        totales.TotalGastos,
		Disponible:         totales.Ahorro,
		PorcentajeAhorro:   totales.PorcentajeAhorro,
		Salud:              services.CalcularSalud(totales),
		Distribucion:       services.CalcularDistribucion(lista),
		UltimosMovimientos: services.PrimerosMovimientos(lista, 5),
		Metas:              calculosMetas.Metas,
		TotalInvertido:     calculosInversiones.BalanceInversiones,
		GananciaEstimada:   calculosInversiones.RendimientoTotal,
		Inversiones:        calculosInversiones.Inversiones,
		Crecimiento:        calculosInversiones.Crecimiento,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Categorías.
// @Description Devuelve las categorías activas con su ícono (selects y filtros del Libro mayor).
// @Success 200 {object} models.RespuestaCategorias
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/categorias [get]
func (c *FinanzasController) ObtenerCategorias() {
	crud := services.NewCrudFinanzasService()

	categorias, err := crud.ObtenerCategorias()

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar las categorías.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaCategorias{
		OK:         true,
		Mensaje:    "Categorías consultadas correctamente.",
		Categorias: services.ArmarCategorias(categorias),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

