package finanzas

import (
	"fmt"
	"math"
	"sort"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// INVERSIONES (tablas inversion, nivel_riesgo e historial_inversion)
// GET /v1/inversiones/usuario/:id
//
// Crear, editar o borrar una inversion no esta aqui: la pagina lo
// hace directo en crud_finanzas (/v1/inversion), como dice la guia.
// ============================================================

// riesgosPorId trae los niveles de riesgo (Bajo, Medio, Alto) y los
// guarda por su Id, para ponerle el nombre del riesgo a cada inversion.
func riesgosPorId() (map[int]string, error) {

	var niveles []models.NivelRiesgo

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/nivel_riesgo?limit=100", &niveles); err != nil {
		return nil, err
	}

	mapa := map[int]string{}
	for _, nivel := range niveles {
		mapa[nivel.Id] = nivel.Nombre
	}

	return mapa, nil
}

// armarInversion pasa una inversion al formato de la pagina,
// con el nombre de su nivel de riesgo.
func armarInversion(inversion models.Inversion, riesgoPorId map[int]string) map[string]interface{} {

	return map[string]interface{}{
		"id":          inversion.Id,
		"nombre":      inversion.Nombre,
		"monto":       inversion.Monto,
		"rendimiento": inversion.Rendimiento,
		"riesgo":      nombreRiesgo(inversion, riesgoPorId),
		"duracion":    inversion.Duracion,
	}
}

// nombreRiesgo devuelve "Bajo", "Medio", "Alto" o "" si no tiene.
func nombreRiesgo(inversion models.Inversion, riesgoPorId map[int]string) string {
	if inversion.IdNivelRiesgo == nil {
		return ""
	}
	return riesgoPorId[inversion.IdNivelRiesgo.Id]
}

// datosInversiones: todo lo que se carga de los CRUD para las inversiones
// de un usuario. Lo usan la pantalla (InversionesUsuario) y el Excel
// (ExportarInversiones), asi las dos muestran exactamente lo mismo.
type datosInversiones struct {
	inversiones   []models.Inversion
	riesgoPorId   map[int]string
	valorPorFecha map[string]float64 // historial de todas, sumado por fecha
	avisos        []string
}

// cargarInversiones revisa el usuario y trae sus inversiones, los niveles
// de riesgo y el historial. Si el historial de alguna inversion no se puede
// traer, no se detiene: se agrega un aviso y la grafica sale con lo que hay.
func cargarInversiones(idUsuario int) (datosInversiones, error) {

	datos := datosInversiones{valorPorFecha: map[string]float64{}, avisos: []string{}}

	// 0. El usuario debe existir
	if err := usuarioExiste(idUsuario); err != nil {
		return datos, err
	}

	// 1. Inversiones activas del usuario y niveles de riesgo
	if err := helpers.ConsultarLista(fmt.Sprintf("%s/inversion?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &datos.inversiones); err != nil {
		return datos, err
	}

	riesgoPorId, err := riesgosPorId()

	if err != nil {
		return datos, err
	}

	datos.riesgoPorId = riesgoPorId

	// 2. El historial de cada inversion (para la grafica)
	for _, inversion := range datos.inversiones {

		var historial []models.HistorialInversion

		if err := helpers.ConsultarLista(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, inversion.Id), &historial); err != nil {
			if len(datos.avisos) == 0 {
				datos.avisos = append(datos.avisos, avisoGraficaInversiones)
			}
			continue
		}

		// Si dos inversiones tienen un valor el mismo dia, se suman
		for _, punto := range historial {
			datos.valorPorFecha[punto.Fecha.Format("2006-01-02")] += punto.Valor
		}
	}

	return datos, nil
}

// grafica devuelve los puntos del historial ordenados de la fecha mas
// vieja a la mas nueva.
func (datos datosInversiones) grafica() []map[string]interface{} {

	fechas := []string{}
	for fecha := range datos.valorPorFecha {
		fechas = append(fechas, fecha)
	}
	sort.Strings(fechas)

	puntos := []map[string]interface{}{}
	for _, fecha := range fechas {
		puntos = append(puntos, map[string]interface{}{"fecha": fecha, "valor": datos.valorPorFecha[fecha]})
	}

	return puntos
}

// totales suma lo invertido (balance) y lo ganado (rendimiento).
func (datos datosInversiones) totales() (balance float64, rendimiento float64) {
	for _, inversion := range datos.inversiones {
		balance += inversion.Monto
		rendimiento += inversion.Rendimiento
	}
	return balance, rendimiento
}

// porcentajeDecimal: que % es "parte" de "total", con un decimal (0 si total es 0).
// Ejemplo: gano 50 sobre 1000 invertidos -> 5.0 %.
func porcentajeDecimal(parte float64, total float64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(parte/total*1000) / 10
}

// InversionesUsuario devuelve las inversiones del usuario con su riesgo,
// el balance (todo lo invertido), el rendimiento total, que porcentaje
// representa ese rendimiento y los puntos de la grafica de crecimiento.
func InversionesUsuario(idUsuario int) (map[string]interface{}, error) {

	datos, err := cargarInversiones(idUsuario)

	if err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}
	for _, inversion := range datos.inversiones {
		lista = append(lista, armarInversion(inversion, datos.riesgoPorId))
	}

	balance, rendimiento := datos.totales()

	resultado := map[string]interface{}{
		"inversiones":            lista,
		"balance":                balance,
		"rendimiento_total":      rendimiento,
		"rendimiento_porcentaje": porcentajeDecimal(rendimiento, balance),
		"grafica":                datos.grafica(),
	}

	// Solo si algo no se pudo cargar
	if len(datos.avisos) > 0 {
		resultado["avisos"] = datos.avisos
	}

	return resultado, nil
}
