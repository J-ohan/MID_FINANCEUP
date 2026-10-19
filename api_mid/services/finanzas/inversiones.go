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

	riesgo := ""
	if inversion.IdNivelRiesgo != nil {
		riesgo = riesgoPorId[inversion.IdNivelRiesgo.Id]
	}

	return map[string]interface{}{
		"id":          inversion.Id,
		"nombre":      inversion.Nombre,
		"monto":       inversion.Monto,
		"rendimiento": inversion.Rendimiento,
		"riesgo":      riesgo,
		"duracion":    inversion.Duracion,
	}
}

// InversionesUsuario devuelve las inversiones del usuario con su riesgo,
// el balance (todo lo invertido), el rendimiento total, que porcentaje
// representa ese rendimiento y los puntos de la grafica de crecimiento.
func InversionesUsuario(idUsuario int) (map[string]interface{}, error) {

	// 1. Pedir al CRUD las inversiones activas del usuario y los riesgos
	var inversiones []models.Inversion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/inversion?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &inversiones); err != nil {
		return nil, err
	}

	riesgoPorId, err := riesgosPorId()

	if err != nil {
		return nil, err
	}

	// 2. Recorrer las inversiones: armar la lista, sumar el balance y el
	//    rendimiento, y juntar el historial de todas por fecha.
	lista := []map[string]interface{}{}
	balance := 0.0
	rendimientoTotal := 0.0
	valorPorFecha := map[string]float64{}

	for _, inversion := range inversiones {

		balance += inversion.Monto
		rendimientoTotal += inversion.Rendimiento

		lista = append(lista, armarInversion(inversion, riesgoPorId))

		// El historial de esta inversion (puede estar vacio)
		var historial []models.HistorialInversion
		helpers.ConsultarLista(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, inversion.Id), &historial)

		// Si dos inversiones tienen un valor el mismo dia, se suman
		for _, punto := range historial {
			valorPorFecha[punto.Fecha.Format("2006-01-02")] += punto.Valor
		}
	}

	// 3. Los puntos de la grafica, ordenados de la fecha mas vieja a la mas nueva
	fechas := []string{}
	for fecha := range valorPorFecha {
		fechas = append(fechas, fecha)
	}
	sort.Strings(fechas)

	grafica := []map[string]interface{}{}
	for _, fecha := range fechas {
		grafica = append(grafica, map[string]interface{}{"fecha": fecha, "valor": valorPorFecha[fecha]})
	}

	// 4. Porcentaje de rendimiento = rendimiento / balance * 100 (con un decimal).
	//    Ejemplo: gano 50 sobre 1000 invertidos -> 5 %.
	porcentaje := 0.0
	if balance > 0 {
		porcentaje = math.Round(rendimientoTotal/balance*1000) / 10
	}

	return map[string]interface{}{
		"inversiones":            lista,
		"balance":                balance,
		"rendimiento_total":      rendimientoTotal,
		"rendimiento_porcentaje": porcentaje,
		"grafica":                grafica,
	}, nil
}
