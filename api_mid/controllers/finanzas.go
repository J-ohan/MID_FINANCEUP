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

// @Title Editar movimiento.
// @Description Edita un movimiento. Si no se envían metodoPago u observaciones, se dejan como estaban.
// @Param id path int true "Id del movimiento"
// @Param body body models.SolicitudMovimiento true "Datos del movimiento"
// @Success 200 {object} models.RespuestaMovimiento
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "El movimiento no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/movimientos/:id [put]
func (c *FinanzasController) EditarMovimiento() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id del movimiento debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

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

	err = json.Unmarshal(body, &solicitud)

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

	movimiento, err := crud.ObtenerMovimiento(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El movimiento no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar el movimiento.",
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

	// Se conserva el mismo dueño y el mismo id
	cambios := services.NuevoMovimiento(movimiento.IdUsuario, solicitud, categoria)
	cambios.Id = movimiento.Id

	if solicitud.MetodoPago == nil {
		cambios.MetodoPago = movimiento.MetodoPago
	}

	if solicitud.Observaciones == nil {
		cambios.Observaciones = movimiento.Observaciones
	}

	err = crud.Actualizar("movimiento_ingreso_egreso", id, cambios)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible actualizar el movimiento.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaMovimiento{
		OK:         true,
		Mensaje:    "Movimiento actualizado correctamente.",
		Movimiento: services.ArmarMovimiento(cambios, categorias),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Eliminar movimiento.
// @Description Elimina un movimiento (queda con activo = false, no se borra de la base).
// @Param id path int true "Id del movimiento"
// @Success 200 {object} map[string]interface{} "Movimiento eliminado"
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "El movimiento no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/movimientos/:id [delete]
func (c *FinanzasController) EliminarMovimiento() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id del movimiento debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	movimiento, err := crud.ObtenerMovimiento(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El movimiento no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar el movimiento.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	movimiento.Activo = false

	err = crud.Actualizar("movimiento_ingreso_egreso", id, movimiento)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible eliminar el movimiento.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"OK":      true,
		"Mensaje": "Movimiento eliminado correctamente.",
		"Id":      id,
	}
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Importar movimientos.
// @Description Carga masiva desde la plantilla de Excel. En modo "validar" solo revisa las filas; en modo "confirmar" guarda las válidas y, si una falla, borra las que ya guardó.
// @Param body body models.SolicitudImportacion true "Usuario, modo y filas del Excel"
// @Success 200 {object} models.RespuestaImportacion "Modo validar"
// @Success 201 {object} models.RespuestaImportacion "Modo confirmar"
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/movimientos/importar [post]
func (c *FinanzasController) ImportarMovimientos() {
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

	var solicitud models.SolicitudImportacion

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

	solicitud.Modo = strings.ToLower(strings.TrimSpace(solicitud.Modo))

	if solicitud.Modo != "validar" && solicitud.Modo != "confirmar" {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El modo debe ser validar o confirmar.",
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

	if len(solicitud.Filas) == 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "Debe ingresar al menos una fila para importar.",
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

	// Revisar fila por fila: las buenas van a "validos" y las malas a "errores"
	validos := []models.MovimientoIngresoEgreso{}
	errores := []models.ResultadoErrorFila{}
	hoy := time.Now().Format("2006-01-02")

	for i, fila := range solicitud.Filas {
		if strings.TrimSpace(fila.Fecha) == "" {
			fila.Fecha = hoy
		}

		mensaje := validarMovimiento(&fila)

		if mensaje != "" {
			errores = append(errores, models.ResultadoErrorFila{
				Fila:     i + 1,
				Concepto: fila.Concepto,
				Motivo:   mensaje,
			})
			continue
		}

		categoria, err := services.BuscarCategoria(categorias, fila.Categoria)

		if err != nil {
			errores = append(errores, models.ResultadoErrorFila{
				Fila:     i + 1,
				Concepto: fila.Concepto,
				Motivo:   err.Error(),
			})
			continue
		}

		validos = append(validos, services.NuevoMovimiento(solicitud.IdUsuario, fila, categoria))
	}

	respuesta := models.RespuestaImportacion{
		OK:      true,
		Modo:    solicitud.Modo,
		Validos: len(validos),
		Errores: errores,
	}

	// En modo "validar" solo se informa, no se guarda nada
	if solicitud.Modo == "validar" {
		respuesta.Mensaje = "Validación terminada."

		c.Data["json"] = respuesta
		c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
		c.ServeJSON()
		return
	}

	if len(validos) == 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "Ninguna fila es válida, no se importó nada.",
			"Errores": errores,
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	// Modo "confirmar": guardar una por una; si una falla, deshacer las anteriores
	guardados := []int{}

	for i := range validos {
		err = crud.Crear("movimiento_ingreso_egreso", &validos[i])

		if err != nil {
			for _, idGuardado := range guardados {
				crud.Borrar("movimiento_ingreso_egreso", idGuardado)
			}

			c.Data["json"] = map[string]interface{}{
				"OK":      false,
				"Mensaje": "No fue posible guardar el movimiento \"" + validos[i].Nombre + "\". No se importó ninguna fila.",
				"Error":   err.Error(),
			}
			c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
			c.ServeJSON()
			return
		}

		guardados = append(guardados, validos[i].Id)
	}

	respuesta.Mensaje = "Movimientos importados correctamente."
	respuesta.Guardados = len(guardados)

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
	c.ServeJSON()
}

// ============================================================
// METAS (meta)
// ============================================================

// @Title Metas del usuario.
// @Description Devuelve las metas de un usuario con su porcentaje, lo que falta y si están cumplidas, más los totales.
// @Param id path int true "Id del usuario"
// @Success 200 {object} models.RespuestaMetas
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/metas/usuario/:id [get]
func (c *FinanzasController) ObtenerMetas() {
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

	calculos := services.CalcularMetas(metas)

	respuesta := models.RespuestaMetas{
		OK:              true,
		Mensaje:         "Metas consultadas correctamente.",
		Metas:           calculos.Metas,
		TotalMetas:      calculos.TotalMetas,
		MetasAlcanzadas: calculos.MetasAlcanzadas,
		TotalAhorrado:   calculos.TotalAhorrado,
		TotalObjetivo:   calculos.TotalObjetivo,
		ProgresoGeneral: calculos.ProgresoGeneral,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Crear meta.
// @Description Crea una meta de ahorro. Si no se envía ícono se usa 🎯.
// @Param body body models.SolicitudMeta true "Datos de la meta"
// @Success 201 {object} models.RespuestaMeta
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/metas [post]
func (c *FinanzasController) CrearMeta() {
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

	var solicitud models.SolicitudMeta

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

	mensaje := validarMeta(&solicitud)

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

	// El formulario "Nueva meta" no usa editar_meta: IdEditarMeta queda vacío
	meta := models.Meta{
		IdUsuario:     solicitud.IdUsuario,
		Nombre:        solicitud.Nombre,
		Icono:         solicitud.Icono,
		MontoObjetivo: solicitud.Objetivo,
		MontoActual:   solicitud.Actual,
		Activo:        true,
	}

	err = crud.Crear("meta", &meta)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible guardar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaMeta{
		OK:      true,
		Mensaje: "Meta creada correctamente.",
		Meta:    services.ArmarMeta(meta),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
	c.ServeJSON()
}

// @Title Editar meta.
// @Description Edita el nombre, el ícono, el monto actual y el objetivo de una meta.
// @Param id path int true "Id de la meta"
// @Param body body models.SolicitudMeta true "Datos de la meta"
// @Success 200 {object} models.RespuestaMeta
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "La meta no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/metas/:id [put]
func (c *FinanzasController) EditarMeta() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id de la meta debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

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

	var solicitud models.SolicitudMeta

	err = json.Unmarshal(body, &solicitud)

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

	mensaje := validarMeta(&solicitud)

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

	meta, err := crud.ObtenerMeta(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "La meta no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	meta.Nombre = solicitud.Nombre
	meta.MontoObjetivo = solicitud.Objetivo
	meta.MontoActual = solicitud.Actual
	meta.Icono = solicitud.Icono

	err = crud.Actualizar("meta", id, meta)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible actualizar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaMeta{
		OK:      true,
		Mensaje: "Meta actualizada correctamente.",
		Meta:    services.ArmarMeta(meta),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Abonar a una meta.
// @Description Suma un abono al monto que lleva ahorrado la meta.
// @Param id path int true "Id de la meta"
// @Param body body models.SolicitudAbono true "Monto del abono"
// @Success 200 {object} models.RespuestaMeta
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "La meta no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/metas/:id/abono [put]
func (c *FinanzasController) AbonarMeta() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id de la meta debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

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

	var solicitud models.SolicitudAbono

	err = json.Unmarshal(body, &solicitud)

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

	if solicitud.Monto <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El monto del abono debe ser mayor a cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	meta, err := crud.ObtenerMeta(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "La meta no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	meta.MontoActual += solicitud.Monto

	err = crud.Actualizar("meta", id, meta)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible registrar el abono.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaMeta{
		OK:      true,
		Mensaje: "Abono registrado correctamente.",
		Meta:    services.ArmarMeta(meta),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Eliminar meta.
// @Description Elimina una meta (queda con activo = false, no se borra de la base).
// @Param id path int true "Id de la meta"
// @Success 200 {object} map[string]interface{} "Meta eliminada"
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "La meta no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/metas/:id [delete]
func (c *FinanzasController) EliminarMeta() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id de la meta debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	meta, err := crud.ObtenerMeta(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "La meta no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	meta.Activo = false

	err = crud.Actualizar("meta", id, meta)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible eliminar la meta.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"OK":      true,
		"Mensaje": "Meta eliminada correctamente.",
		"Id":      id,
	}
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// ============================================================
// INVERSIONES (inversion, nivel_riesgo, historial_inversion)
// ============================================================

// @Title Inversiones del usuario.
// @Description Devuelve las inversiones de un usuario con el total invertido, el rendimiento, el reparto por riesgo, la alerta de concentración y la gráfica de crecimiento.
// @Param id path int true "Id del usuario"
// @Success 200 {object} models.RespuestaInversiones
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/inversiones/usuario/:id [get]
func (c *FinanzasController) ObtenerInversiones() {
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

	calculos := services.CalcularInversiones(inversiones, niveles, crud.ObtenerHistorial())

	respuesta := models.RespuestaInversiones{
		OK:                    true,
		Mensaje:               "Inversiones consultadas correctamente.",
		Inversiones:           calculos.Inversiones,
		BalanceInversiones:    calculos.BalanceInversiones,
		RendimientoTotal:      calculos.RendimientoTotal,
		RendimientoPorcentaje: calculos.RendimientoPorcentaje,
		TotalPorRiesgo:        calculos.TotalPorRiesgo,
		AlertaConcentracion:   calculos.AlertaConcentracion,
		Carteras:              calculos.Carteras,
		Crecimiento:           calculos.Crecimiento,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Crear inversión.
// @Description Crea una inversión. La rentabilidad (%) se calcula con el rendimiento en pesos.
// @Param body body models.SolicitudInversion true "Datos de la inversión"
// @Success 201 {object} models.RespuestaInversion
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "El usuario no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/inversiones [post]
func (c *FinanzasController) CrearInversion() {
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

	var solicitud models.SolicitudInversion

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

	mensaje := validarInversion(&solicitud)

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

	nivel, err := services.BuscarNivelRiesgo(niveles, solicitud.Riesgo)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	inversion := models.Inversion{
		IdUsuario:     solicitud.IdUsuario,
		IdNivelRiesgo: &models.Ref{Id: nivel.Id},
		Nombre:        solicitud.Nombre,
		Monto:         solicitud.Monto,
		Rendimiento:   solicitud.Rendimiento,
		Rentabilidad:  services.CalcularRentabilidad(solicitud.Rendimiento, solicitud.Monto),
		Duracion:      solicitud.Duracion,
		FechaInicio:   time.Now(),
		Activo:        true,
	}

	err = crud.Crear("inversion", &inversion)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible guardar la inversión.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaInversion{
		OK:        true,
		Mensaje:   "Inversión creada correctamente.",
		Inversion: services.ArmarInversion(inversion, niveles),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
	c.ServeJSON()
}

// @Title Editar inversión.
// @Description Edita una inversión.
// @Param id path int true "Id de la inversión"
// @Param body body models.SolicitudInversion true "Datos de la inversión"
// @Success 200 {object} models.RespuestaInversion
// @Failure 400 {object} map[string]interface{} "Solicitud inválida"
// @Failure 404 {object} map[string]interface{} "La inversión no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/inversiones/:id [put]
func (c *FinanzasController) EditarInversion() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id de la inversión debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

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

	var solicitud models.SolicitudInversion

	err = json.Unmarshal(body, &solicitud)

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

	mensaje := validarInversion(&solicitud)

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

	inversion, err := crud.ObtenerInversion(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "La inversión no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar la inversión.",
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

	nivel, err := services.BuscarNivelRiesgo(niveles, solicitud.Riesgo)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	inversion.Nombre = solicitud.Nombre
	inversion.Monto = solicitud.Monto
	inversion.Rendimiento = solicitud.Rendimiento
	inversion.Rentabilidad = services.CalcularRentabilidad(solicitud.Rendimiento, solicitud.Monto)
	inversion.Duracion = solicitud.Duracion
	inversion.IdNivelRiesgo = &models.Ref{Id: nivel.Id}

	err = crud.Actualizar("inversion", id, inversion)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible actualizar la inversión.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	respuesta := models.RespuestaInversion{
		OK:        true,
		Mensaje:   "Inversión actualizada correctamente.",
		Inversion: services.ArmarInversion(inversion, niveles),
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Eliminar inversión.
// @Description Elimina una inversión (queda con activo = false, no se borra de la base).
// @Param id path int true "Id de la inversión"
// @Success 200 {object} map[string]interface{} "Inversión eliminada"
// @Failure 400 {object} map[string]interface{} "Id inválido"
// @Failure 404 {object} map[string]interface{} "La inversión no existe"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/inversiones/:id [delete]
func (c *FinanzasController) EliminarInversion() {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "El id de la inversión debe ser un número mayor que cero.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.ServeJSON()
		return
	}

	crud := services.NewCrudFinanzasService()

	inversion, err := crud.ObtenerInversion(id)

	if errors.Is(err, services.ErrNoEncontrado) {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "La inversión no existe.",
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar la inversión.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	inversion.Activo = false

	err = crud.Actualizar("inversion", id, inversion)

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible eliminar la inversión.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"OK":      true,
		"Mensaje": "Inversión eliminada correctamente.",
		"Id":      id,
	}
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// ============================================================
// ADMIN (datos de todos los usuarios)
// ============================================================

// @Title Movimientos de todos los usuarios.
// @Description Devuelve los movimientos de todos los usuarios y sus totales. Acepta los mismos filtros del Libro mayor.
// @Success 200 {object} models.RespuestaMovimientos
// @Failure 400 {object} map[string]interface{} "Filtros inválidos"
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/admin/movimientos [get]
func (c *FinanzasController) ObtenerMovimientosAdmin() {
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

	movimientos, err := crud.ObtenerMovimientos(0)

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

// @Title Metas de todos los usuarios.
// @Description Devuelve las metas de todos los usuarios y sus totales.
// @Success 200 {object} models.RespuestaMetas
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/admin/metas [get]
func (c *FinanzasController) ObtenerMetasAdmin() {
	crud := services.NewCrudFinanzasService()

	metas, err := crud.ObtenerMetas(0)

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

	calculos := services.CalcularMetas(metas)

	respuesta := models.RespuestaMetas{
		OK:              true,
		Mensaje:         "Metas consultadas correctamente.",
		Metas:           calculos.Metas,
		TotalMetas:      calculos.TotalMetas,
		MetasAlcanzadas: calculos.MetasAlcanzadas,
		TotalAhorrado:   calculos.TotalAhorrado,
		TotalObjetivo:   calculos.TotalObjetivo,
		ProgresoGeneral: calculos.ProgresoGeneral,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Inversiones de todos los usuarios.
// @Description Devuelve las inversiones de todos los usuarios y sus totales.
// @Success 200 {object} models.RespuestaInversiones
// @Failure 502 {object} map[string]interface{} "Error al consultar el CRUD"
// @Router /v1/admin/inversiones [get]
func (c *FinanzasController) ObtenerInversionesAdmin() {
	crud := services.NewCrudFinanzasService()

	inversiones, err := crud.ObtenerInversiones(0)

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

	calculos := services.CalcularInversiones(inversiones, niveles, crud.ObtenerHistorial())

	respuesta := models.RespuestaInversiones{
		OK:                    true,
		Mensaje:               "Inversiones consultadas correctamente.",
		Inversiones:           calculos.Inversiones,
		BalanceInversiones:    calculos.BalanceInversiones,
		RendimientoTotal:      calculos.RendimientoTotal,
		RendimientoPorcentaje: calculos.RendimientoPorcentaje,
		TotalPorRiesgo:        calculos.TotalPorRiesgo,
		AlertaConcentracion:   calculos.AlertaConcentracion,
		Carteras:              calculos.Carteras,
		Crecimiento:           calculos.Crecimiento,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// @Title Libro mayor de todos los usuarios.
// @Description Devuelve el libro mayor agrupado por usuario, con su nombre, y los totales de la plataforma.
// @Success 200 {object} models.RespuestaLibrosAdmin
// @Failure 502 {object} map[string]interface{} "Error al consultar los CRUD"
// @Router /v1/admin/libro-mayor [get]
func (c *FinanzasController) ObtenerLibroMayorAdmin() {
	crud := services.NewCrudFinanzasService()

	nombres, err := crud.ObtenerNombresUsuarios()

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"OK":      false,
			"Mensaje": "No fue posible consultar los usuarios.",
			"Error":   err.Error(),
		}
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.ServeJSON()
		return
	}

	movimientos, err := crud.ObtenerMovimientos(0)

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

	calculos := services.CalcularLibrosAdmin(services.ArmarMovimientos(movimientos, categorias), nombres)

	respuesta := models.RespuestaLibrosAdmin{
		OK:                      true,
		Mensaje:                 "Libro mayor consultado correctamente.",
		Libros:                  calculos.Libros,
		TotalIngresosPlataforma: calculos.TotalIngresosPlataforma,
		TotalEgresosPlataforma:  calculos.TotalEgresosPlataforma,
		UsuariosActivos:         calculos.UsuariosActivos,
	}

	c.Data["json"] = respuesta
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	c.ServeJSON()
}

// ============================================================
// VALIDACIONES
// Devuelven el mensaje del primer error encontrado, o "" si todo está bien.
// ============================================================

// validarMovimiento revisa los datos de un movimiento y los limpia
// (quita espacios y pasa el tipo a minúsculas).
func validarMovimiento(solicitud *models.SolicitudMovimiento) string {
	solicitud.Concepto = strings.TrimSpace(solicitud.Concepto)
	solicitud.Categoria = strings.TrimSpace(solicitud.Categoria)
	solicitud.Tipo = strings.ToLower(strings.TrimSpace(solicitud.Tipo))
	solicitud.Fecha = strings.TrimSpace(solicitud.Fecha)

	if solicitud.Concepto == "" {
		return "El concepto es obligatorio."
	}

	if utf8.RuneCountInString(solicitud.Concepto) > 120 {
		return "El concepto no puede tener más de 120 caracteres."
	}

	if solicitud.Valor <= 0 {
		return "El valor debe ser mayor a cero."
	}

	if solicitud.Tipo != "ingreso" && solicitud.Tipo != "gasto" {
		return "El tipo debe ser ingreso o gasto."
	}

	if solicitud.Categoria == "" {
		return "La categoría es obligatoria."
	}

	_, err := time.Parse("2006-01-02", solicitud.Fecha)

	if err != nil {
		return "La fecha debe tener el formato AAAA-MM-DD."
	}

	if solicitud.MetodoPago != nil && utf8.RuneCountInString(*solicitud.MetodoPago) > 50 {
		return "El método de pago no puede tener más de 50 caracteres."
	}

	return ""
}

// validarMeta revisa los datos de una meta. Si no trae ícono, le pone 🎯.
func validarMeta(solicitud *models.SolicitudMeta) string {
	solicitud.Nombre = strings.TrimSpace(solicitud.Nombre)
	solicitud.Icono = strings.TrimSpace(solicitud.Icono)

	if solicitud.Icono == "" {
		solicitud.Icono = "🎯"
	}

	if solicitud.Nombre == "" {
		return "El nombre de la meta es obligatorio."
	}

	if utf8.RuneCountInString(solicitud.Nombre) > 120 {
		return "El nombre no puede tener más de 120 caracteres."
	}

	if solicitud.Objetivo <= 0 {
		return "El monto objetivo debe ser mayor a cero."
	}

	if solicitud.Actual < 0 {
		return "El monto actual no puede ser negativo."
	}

	if utf8.RuneCountInString(solicitud.Icono) > 50 {
		return "El ícono no puede tener más de 50 caracteres."
	}

	return ""
}

// validarInversion revisa los datos de una inversión y deja el riesgo
// bien escrito ("bajo" -> "Bajo").
func validarInversion(solicitud *models.SolicitudInversion) string {
	solicitud.Nombre = strings.TrimSpace(solicitud.Nombre)
	solicitud.Duracion = strings.TrimSpace(solicitud.Duracion)
	solicitud.Riesgo = services.Capitalizar(solicitud.Riesgo)

	if solicitud.Nombre == "" {
		return "El nombre de la inversión es obligatorio."
	}

	if utf8.RuneCountInString(solicitud.Nombre) > 120 {
		return "El nombre no puede tener más de 120 caracteres."
	}

	if solicitud.Duracion == "" {
		return "La duración es obligatoria."
	}

	if utf8.RuneCountInString(solicitud.Duracion) > 30 {
		return "La duración no puede tener más de 30 caracteres."
	}

	if solicitud.Monto <= 0 {
		return "El monto debe ser mayor a cero."
	}

	if solicitud.Riesgo != "Bajo" && solicitud.Riesgo != "Medio" && solicitud.Riesgo != "Alto" {
		return "El riesgo debe ser Bajo, Medio o Alto."
	}

	if solicitud.Rendimiento < -solicitud.Monto {
		return "La pérdida no puede ser mayor que el monto invertido."
	}

	// La base guarda la rentabilidad con máximo 4 cifras enteras
	if math.Abs(services.CalcularRentabilidad(solicitud.Rendimiento, solicitud.Monto)) >= 10000 {
		return "El rendimiento no es coherente con el monto invertido."
	}

	return ""
}

// leerFiltros lee los filtros del Libro mayor de la URL y revisa que estén bien.
func (c *FinanzasController) leerFiltros() (models.FiltrosMovimientos, string) {
	filtros := models.FiltrosMovimientos{
		Tipo:      strings.ToLower(strings.TrimSpace(c.GetString("tipo"))),
		Categoria: c.GetString("categoria"),
		Desde:     strings.TrimSpace(c.GetString("desde")),
		Hasta:     strings.TrimSpace(c.GetString("hasta")),
		Texto:     c.GetString("q"),
	}

	if filtros.Tipo != "" && filtros.Tipo != "todos" && filtros.Tipo != "ingreso" && filtros.Tipo != "gasto" {
		return filtros, "El filtro tipo debe ser todos, ingreso o gasto."
	}

	for _, fecha := range []string{filtros.Desde, filtros.Hasta} {
		if fecha == "" {
			continue
		}

		_, err := time.Parse("2006-01-02", fecha)

		if err != nil {
			return filtros, "Las fechas de los filtros deben tener el formato AAAA-MM-DD."
		}
	}

	return filtros, ""
}
