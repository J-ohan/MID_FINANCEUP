package services

import (
	"sort"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

const movimientosRecientes = 5

// ResumenFinanciero calcula ingresos, gastos, balance, avance de metas,
// inversiones y solicitudes de deuda de un usuario.
func ResumenFinanciero(idUsuario int) (models.ResumenFinanciero, error) {
	resumen := models.ResumenFinanciero{
		IdUsuario:          idUsuario,
		GastosPorCategoria: []models.GastoCategoria{},
		UltimosMovimientos: []models.MovimientoVista{},
		Metas:              []models.MetaVista{},
		SolicitudesDeuda:   []models.SolicitudDeuda{},
		Inversiones:        models.ResumenInversiones{Detalle: []models.InversionVista{}},
	}
	if err := ExisteUsuario(idUsuario); err != nil {
		return resumen, err
	}
	filtroUsuario := helpers.Filtro("id_usuario", idUsuario, "activo", true)

	// 1. Movimientos: suma de ingresos y gastos, y gastos por categoria.
	var categorias []models.Categoria
	if err := helpers.Consultar("finanzas", "categoria?limit=1000", &categorias); err != nil {
		return resumen, err
	}
	nombreCategoria := map[int]string{}
	for _, c := range categorias {
		nombreCategoria[c.Id] = c.Nombre
	}

	var movimientos []models.MovimientoIngresoEgreso
	if err := helpers.Consultar("finanzas", "movimiento_ingreso_egreso?"+filtroUsuario, &movimientos); err != nil {
		return resumen, err
	}

	gastoCategoria := map[string]float64{}
	for _, m := range movimientos {
		categoria := "Sin categoria"
		if m.IdCategoria != nil {
			categoria = nombreCategoria[m.IdCategoria.Id]
		}
		if m.EsIngreso {
			resumen.TotalIngresos += m.Monto
		} else {
			resumen.TotalGastos += m.Monto
			gastoCategoria[categoria] += m.Monto
		}
	}
	resumen.Balance = resumen.TotalIngresos - resumen.TotalGastos

	for categoria, total := range gastoCategoria {
		resumen.GastosPorCategoria = append(resumen.GastosPorCategoria, models.GastoCategoria{Categoria: categoria, Total: total})
	}
	sort.Slice(resumen.GastosPorCategoria, func(i, j int) bool {
		return resumen.GastosPorCategoria[i].Total > resumen.GastosPorCategoria[j].Total
	})

	// Los mas recientes primero.
	sort.Slice(movimientos, func(i, j int) bool { return movimientos[i].Fecha.After(movimientos[j].Fecha) })
	for i, m := range movimientos {
		if i == movimientosRecientes {
			break
		}
		vista := models.MovimientoVista{
			Id:            m.Id,
			Fecha:         fecha(m.Fecha),
			Concepto:      m.Nombre,
			Categoria:     "Sin categoria",
			Tipo:          "gasto",
			Valor:         m.Monto,
			MetodoPago:    m.MetodoPago,
			Observaciones: m.Observaciones,
		}
		if m.EsIngreso {
			vista.Tipo = "ingreso"
		}
		if m.IdCategoria != nil {
			vista.Categoria = nombreCategoria[m.IdCategoria.Id]
		}
		resumen.UltimosMovimientos = append(resumen.UltimosMovimientos, vista)
	}

	// 2. Metas: porcentaje de avance.
	var metas []models.Meta
	if err := helpers.Consultar("finanzas", "meta?"+filtroUsuario, &metas); err != nil {
		return resumen, err
	}
	for _, m := range metas {
		porcentaje := 0.0
		if m.MontoObjetivo > 0 {
			porcentaje = m.MontoActual / m.MontoObjetivo * 100
		}
		if porcentaje > 100 {
			porcentaje = 100
		}
		resumen.Metas = append(resumen.Metas, models.MetaVista{
			Id:          m.Id,
			Nombre:      m.Nombre,
			Icono:       m.Icono,
			Color:       m.Color,
			Actual:      m.MontoActual,
			Objetivo:    m.MontoObjetivo,
			Porcentaje:  redondear(porcentaje),
			Cumplida:    m.MontoActual >= m.MontoObjetivo,
			FechaLimite: fecha(m.FechaLimite),
		})
	}

	// 3. Inversiones: total invertido y ganancia estimada segun la rentabilidad.
	var inversiones []models.Inversion
	if err := helpers.Consultar("finanzas", "inversion?"+filtroUsuario, &inversiones); err != nil {
		return resumen, err
	}
	var riesgos []models.NivelRiesgo
	var tipos []models.TipoInversion
	helpers.Consultar("finanzas", "nivel_riesgo?limit=1000", &riesgos)
	helpers.Consultar("finanzas", "tipo_inversion?limit=1000", &tipos)
	nombreRiesgo := map[int]string{}
	for _, r := range riesgos {
		nombreRiesgo[r.Id] = r.Nombre
	}
	nombreTipo := map[int]string{}
	for _, t := range tipos {
		nombreTipo[t.Id] = t.Nombre
	}

	for _, inv := range inversiones {
		resumen.Inversiones.TotalInvertido += inv.Monto
		resumen.Inversiones.GananciaEstimada += inv.Monto * inv.Rentabilidad / 100
		vista := models.InversionVista{
			Id:           inv.Id,
			Nombre:       inv.Nombre,
			Monto:        inv.Monto,
			Rentabilidad: inv.Rentabilidad,
			FechaInicio:  fecha(inv.FechaInicio),
			FechaFin:     fecha(inv.FechaFin),
		}
		if inv.IdTipoInversion != nil {
			vista.Tipo = nombreTipo[inv.IdTipoInversion.Id]
		}
		if inv.IdNivelRiesgo != nil {
			vista.Riesgo = nombreRiesgo[inv.IdNivelRiesgo.Id]
		}
		resumen.Inversiones.Detalle = append(resumen.Inversiones.Detalle, vista)
	}
	resumen.Inversiones.Cantidad = len(inversiones)
	resumen.Inversiones.GananciaEstimada = redondear(resumen.Inversiones.GananciaEstimada)

	// 4. Solicitudes de "Resuelve tu deuda": cuanto se ahorraria al mes.
	var solicitudes []models.SolicitudConsolidacion
	if err := helpers.Consultar("finanzas", "solicitud_consolidacion?"+filtroUsuario, &solicitudes); err != nil {
		return resumen, err
	}
	for _, s := range solicitudes {
		resumen.SolicitudesDeuda = append(resumen.SolicitudesDeuda, models.SolicitudDeuda{
			Id:             s.Id,
			SaldoTotal:     s.SaldoTotal,
			CuotaActual:    s.CuotaActual,
			CuotaPropuesta: s.CuotaPropuesta,
			AhorroMensual:  s.CuotaActual - s.CuotaPropuesta,
			Estado:         s.Estado,
		})
	}

	return resumen, nil
}
