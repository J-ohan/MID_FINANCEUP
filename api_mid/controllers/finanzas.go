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

// ============================================================
// LIBRO MAYOR (movimiento_ingreso_egreso)
// ============================================================

// @Title Movimientos del usuario.
// @Description Devuelve los movimientos de un usuario y sus totales. Acepta filtros opcionales en la URL.
// @Param id path int true "Id del usuario"
// @Param tipo query string false "todos, ingreso o gasto"
// @Param categoria query string false "Nombre de la categoría"
// @Param desde query string false "Fecha inicial AAAA-MM-DD"
// @Param hasta query string false "Fecha final AAAA-MM-DD"
// @Param q query string false "Texto a buscar en el concepto"
// @Success 200 {object} models.RespuestaMovimientos
// @Failure 400 {object} map[string]interface{} "Id o filtros inválidos"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/movimientos/usuario/:id [get]
func (c *FinanzasController) ObtenerMovimientos() {
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

	filtros, mensaje := c.leerFiltros()

	if mensaje != "" {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": mensaje,
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

	lista := services.FiltrarMovimientos(services.ArmarMovimientos(movimientos, categorias), filtros)

	respuesta := models.RespuestaMovimientos{
		OK:          true,
		Mensaje:     "Movimientos consultados correctamente.",
		Movimientos: lista,
		Totales:     services.CalcularTotales(lista),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Crear movimiento.
// @Description Registra un ingreso o un gasto.
// @Param body body models.SolicitudMovimiento true "Datos del movimiento"
// @Success 201 {object} models.RespuestaMovimiento
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/movimientos [post]
func (c *FinanzasController) CrearMovimiento() {
	body := c.Ctx.Input.RequestBody

	if len(body) == 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El cuerpo de la solicitud se encuentra vacío.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	var solicitud models.SolicitudMovimiento

	err := json.Unmarshal(body, &solicitud)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El cuerpo de la solicitud no tiene formato válido.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	if solicitud.IdUsuario <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id del usuario es obligatorio.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	mensaje := validarMovimiento(&solicitud)

	if mensaje != "" {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": mensaje,
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	_, err = crud.ObtenerUsuario(solicitud.IdUsuario)

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

	categoria, err := services.BuscarCategoria(categorias, solicitud.Categoria)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	movimiento := services.NuevoMovimiento(solicitud.IdUsuario, solicitud, categoria)

	err = crud.Crear("movimiento_ingreso_egreso", &movimiento)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible guardar el movimiento.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaMovimiento{
		OK:         true,
		Mensaje:    "Movimiento creado correctamente.",
		Movimiento: services.ArmarMovimiento(movimiento, categorias),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
	c.ServeJSON()
}
