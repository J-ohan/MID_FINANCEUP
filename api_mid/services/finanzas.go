package services

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE FINANZAS (crud_finanzas)
// Aqui esta la logica: sumas, porcentajes, filtros, importar.
// ============================================================

// categoriasPorId trae todas las categorias y las guarda por su id.
func categoriasPorId() (map[int]models.Categoria, error) {

	var categorias []models.Categoria

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/categoria?limit=1000", &categorias); err != nil {
		return nil, err
	}

	mapa := map[int]models.Categoria{}
	for _, categoria := range categorias {
		mapa[categoria.Id] = categoria
	}

	return mapa, nil
}

// tipoMovimiento convierte EsIngreso en "ingreso" o "gasto".
func tipoMovimiento(movimiento models.Movimiento) string {
	if movimiento.EsIngreso {
		return "ingreso"
	}
	return "gasto"
}

// ResumenFinanciero: ingresos, gastos, balance, gastos por categoria
// y los ultimos 5 movimientos.
func ResumenFinanciero(idUsuario int) (map[string]interface{}, error) {

	// 1. Movimientos del usuario (mas nuevos primero) y categorias
	var movimientos []models.Movimiento

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/movimiento_ingreso_egreso?query=id_usuario:%d,activo:true&sortby=Fecha&order=desc&limit=10000", helpers.UrlCrudFinanzas, idUsuario), &movimientos); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	// 2. Sumar ingresos, gastos y gastos por categoria
	totalIngresos := 0.0
	totalGastos := 0.0
	gastoPorCategoria := map[int]float64{}

	for _, movimiento := range movimientos {

		if movimiento.EsIngreso {
			totalIngresos += movimiento.Monto
			continue
		}

		totalGastos += movimiento.Monto

		if movimiento.IdCategoria != nil {
			gastoPorCategoria[movimiento.IdCategoria.Id] += movimiento.Monto
		}
	}

	// 3. Porcentaje de cada categoria sobre el total de gastos
	distribucion := []map[string]interface{}{}

	for idCategoria, total := range gastoPorCategoria {

		porcentaje := 0.0
		if totalGastos > 0 {
			porcentaje = math.Round(total / totalGastos * 100)
		}

		distribucion = append(distribucion, map[string]interface{}{
			"categoria":  categorias[idCategoria].Nombre,
			"icono":      categorias[idCategoria].Icono,
			"total":      total,
			"porcentaje": porcentaje,
		})
	}

	// 4. Ultimos 5 movimientos
	ultimos := []map[string]interface{}{}

	for i, movimiento := range movimientos {

		if i == 5 {
			break
		}

		nombreCategoria := ""
		if movimiento.IdCategoria != nil {
			nombreCategoria = categorias[movimiento.IdCategoria.Id].Nombre
		}

		ultimos = append(ultimos, map[string]interface{}{
			"id":        movimiento.Id,
			"fecha":     movimiento.Fecha.Format("2006-01-02"),
			"concepto":  movimiento.Nombre,
			"categoria": nombreCategoria,
			"tipo":      tipoMovimiento(movimiento),
			"valor":     movimiento.Monto,
		})
	}

	return map[string]interface{}{
		"id_usuario":          idUsuario,
		"total_ingresos":      totalIngresos,
		"total_gastos":        totalGastos,
		"balance":             totalIngresos - totalGastos,
		"gastos_categoria":    distribucion,
		"ultimos_movimientos": ultimos,
	}, nil
}

// LibroMayor: movimientos del usuario con nombre e icono de su categoria.
func LibroMayor(idUsuario int, filtros models.FiltrosMovimientos) ([]map[string]interface{}, error) {

	// 1. Armar el filtro para el CRUD
	filtro := fmt.Sprintf("id_usuario:%d,activo:true", idUsuario)

	if filtros.Tipo == "ingreso" {
		filtro += ",es_ingreso:true"
	}

	if filtros.Tipo == "gasto" {
		filtro += ",es_ingreso:false"
	}

	if filtros.IdCategoria > 0 {
		filtro += fmt.Sprintf(",id_categoria:%d", filtros.IdCategoria)
	}

	// 2. Consultar movimientos y categorias
	var movimientos []models.Movimiento

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso?query="+filtro+"&sortby=Fecha&order=desc&limit=10000", &movimientos); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	// 3. Armar la lista como la usa la pagina
	lista := []map[string]interface{}{}

	for _, movimiento := range movimientos {

		categoria := models.Categoria{}
		if movimiento.IdCategoria != nil {
			categoria = categorias[movimiento.IdCategoria.Id]
		}

		lista = append(lista, map[string]interface{}{
			"id":            movimiento.Id,
			"fecha":         movimiento.Fecha.Format("2006-01-02"),
			"concepto":      movimiento.Nombre,
			"id_categoria":  categoria.Id,
			"categoria":     categoria.Nombre,
			"icono":         categoria.Icono,
			"tipo":          tipoMovimiento(movimiento),
			"valor":         movimiento.Monto,
			"metodo_pago":   movimiento.MetodoPago,
			"observaciones": movimiento.Observaciones,
		})
	}

	return lista, nil
}

// ImportarMovimientos: guarda las filas del Excel una por una.
// Devuelve cuantas se guardaron y que filas fallaron.
func ImportarMovimientos(datos models.DatosImportar) (map[string]interface{}, error) {

	if len(datos.Movimientos) == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "No llego ningun movimiento")
	}

	// 1. Categorias por nombre (en minuscula)
	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	categoriaPorNombre := map[string]int{}
	for _, categoria := range categorias {
		categoriaPorNombre[strings.ToLower(categoria.Nombre)] = categoria.Id
	}

	// 2. Revisar y guardar cada fila
	guardados := 0
	errores := []string{}

	for i, fila := range datos.Movimientos {

		numeroFila := i + 1

		fecha, err := time.Parse("2006-01-02", fila.Fecha)

		if err != nil {
			errores = append(errores, fmt.Sprintf("Fila %d: la fecha debe ser AAAA-MM-DD", numeroFila))
			continue
		}

		if fila.Concepto == "" || fila.Valor <= 0 {
			errores = append(errores, fmt.Sprintf("Fila %d: falta el concepto o el valor", numeroFila))
			continue
		}

		if fila.Tipo != "ingreso" && fila.Tipo != "gasto" {
			errores = append(errores, fmt.Sprintf("Fila %d: el tipo debe ser ingreso o gasto", numeroFila))
			continue
		}

		movimiento := models.Movimiento{
			IdUsuario:     datos.IdUsuario,
			Nombre:        fila.Concepto,
			Monto:         fila.Valor,
			EsIngreso:     fila.Tipo == "ingreso",
			Fecha:         fecha,
			MetodoPago:    fila.MetodoPago,
			Observaciones: fila.Observaciones,
			Activo:        true,
		}

		if idCategoria, existe := categoriaPorNombre[strings.ToLower(fila.Categoria)]; existe {
			movimiento.IdCategoria = &models.Ref{Id: idCategoria}
		}

		if err := helpers.Crear(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso", movimiento, nil); err != nil {
			errores = append(errores, fmt.Sprintf("Fila %d: no se pudo guardar", numeroFila))
			continue
		}

		guardados++
	}

	return map[string]interface{}{"guardados": guardados, "errores": errores}, nil
}

// MetasUsuario: % de cada meta, cuales estan cumplidas y totales.
func MetasUsuario(idUsuario int) (map[string]interface{}, error) {

	var metas []models.Meta

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/meta?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &metas); err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}
	totalAhorrado := 0.0
	totalObjetivo := 0.0
	cumplidas := 0

	for _, meta := range metas {

		// porcentaje = ahorrado / objetivo * 100 (maximo 100)
		porcentaje := 0.0
		if meta.MontoObjetivo > 0 {
			porcentaje = math.Min(100, math.Round(meta.MontoActual/meta.MontoObjetivo*100))
		}

		cumplida := porcentaje >= 100
		if cumplida {
			cumplidas++
		}

		totalAhorrado += meta.MontoActual
		totalObjetivo += meta.MontoObjetivo

		lista = append(lista, map[string]interface{}{
			"id":         meta.Id,
			"nombre":     meta.Nombre,
			"icono":      meta.Icono,
			"actual":     meta.MontoActual,
			"objetivo":   meta.MontoObjetivo,
			"porcentaje": porcentaje,
			"cumplida":   cumplida,
		})
	}

	return map[string]interface{}{
		"metas":          lista,
		"total_ahorrado": totalAhorrado,
		"total_objetivo": totalObjetivo,
		"cumplidas":      cumplidas,
	}, nil
}

// AbonarMeta: suma dinero a lo ahorrado en una meta.
func AbonarMeta(idMeta int, valor float64) (map[string]interface{}, error) {

	if valor <= 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El valor del abono debe ser mayor a 0")
	}

	// La meta se guarda en un map para no perder ningun campo en el PUT
	var meta map[string]interface{}
	direccion := fmt.Sprintf("%s/meta/%d", helpers.UrlCrudFinanzas, idMeta)

	if err := helpers.ConsultarUno(direccion, &meta, "La meta no existe"); err != nil {
		return nil, err
	}

	actual, _ := meta["MontoActual"].(float64)
	meta["MontoActual"] = actual + valor

	if err := helpers.Actualizar(direccion, meta); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       idMeta,
		"actual":   meta["MontoActual"],
		"objetivo": meta["MontoObjetivo"],
	}, nil
}

// InversionesUsuario: inversiones con su riesgo, balance, rendimiento y grafica.
func InversionesUsuario(idUsuario int) (map[string]interface{}, error) {

	// 1. Inversiones y niveles de riesgo
	var inversiones []models.Inversion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/inversion?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &inversiones); err != nil {
		return nil, err
	}

	var niveles []models.NivelRiesgo

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/nivel_riesgo?limit=100", &niveles); err != nil {
		return nil, err
	}

	riesgoPorId := map[int]string{}
	for _, nivel := range niveles {
		riesgoPorId[nivel.Id] = nivel.Nombre
	}

	// 2. Armar la lista, sumar y juntar el historial por fecha
	lista := []map[string]interface{}{}
	balance := 0.0
	rendimientoTotal := 0.0
	valorPorFecha := map[string]float64{}

	for _, inversion := range inversiones {

		balance += inversion.Monto
		rendimientoTotal += inversion.Rendimiento

		riesgo := ""
		if inversion.IdNivelRiesgo != nil {
			riesgo = riesgoPorId[inversion.IdNivelRiesgo.Id]
		}

		lista = append(lista, map[string]interface{}{
			"id":          inversion.Id,
			"nombre":      inversion.Nombre,
			"monto":       inversion.Monto,
			"rendimiento": inversion.Rendimiento,
			"riesgo":      riesgo,
			"duracion":    inversion.Duracion,
		})

		var historial []models.HistorialInversion
		helpers.ConsultarLista(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, inversion.Id), &historial)

		for _, punto := range historial {
			valorPorFecha[punto.Fecha.Format("2006-01-02")] += punto.Valor
		}
	}

	// 3. Puntos de la grafica ordenados por fecha
	fechas := []string{}
	for fecha := range valorPorFecha {
		fechas = append(fechas, fecha)
	}
	sort.Strings(fechas)

	grafica := []map[string]interface{}{}
	for _, fecha := range fechas {
		grafica = append(grafica, map[string]interface{}{"fecha": fecha, "valor": valorPorFecha[fecha]})
	}

	// 4. porcentaje = rendimiento / balance * 100 (con un decimal)
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

// FinanzasAdmin: movimientos, metas e inversiones de todos los usuarios
// con el nombre de cada uno (crud_auth + crud_finanzas).
func FinanzasAdmin() (map[string]interface{}, error) {

	// 1. Nombres de los usuarios
	nombres, err := NombresUsuarios()

	if err != nil {
		return nil, err
	}

	// 2. Movimientos, metas e inversiones de todos
	var movimientos []models.Movimiento
	var metas []models.Meta
	var inversiones []models.Inversion

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso?query=activo:true&sortby=Fecha&order=desc&limit=10000", &movimientos); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/meta?query=activo:true&limit=10000", &metas); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/inversion?query=activo:true&limit=10000", &inversiones); err != nil {
		return nil, err
	}

	// 3. Armar las tres listas con el nombre del usuario
	listaMovimientos := []map[string]interface{}{}
	for _, movimiento := range movimientos {
		listaMovimientos = append(listaMovimientos, map[string]interface{}{
			"id":       movimiento.Id,
			"usuario":  nombres[movimiento.IdUsuario],
			"fecha":    movimiento.Fecha.Format("2006-01-02"),
			"concepto": movimiento.Nombre,
			"tipo":     tipoMovimiento(movimiento),
			"valor":    movimiento.Monto,
		})
	}

	listaMetas := []map[string]interface{}{}
	for _, meta := range metas {
		listaMetas = append(listaMetas, map[string]interface{}{
			"id":       meta.Id,
			"usuario":  nombres[meta.IdUsuario],
			"nombre":   meta.Nombre,
			"actual":   meta.MontoActual,
			"objetivo": meta.MontoObjetivo,
		})
	}

	listaInversiones := []map[string]interface{}{}
	for _, inversion := range inversiones {
		listaInversiones = append(listaInversiones, map[string]interface{}{
			"id":          inversion.Id,
			"usuario":     nombres[inversion.IdUsuario],
			"nombre":      inversion.Nombre,
			"monto":       inversion.Monto,
			"rendimiento": inversion.Rendimiento,
		})
	}

	return map[string]interface{}{
		"movimientos": listaMovimientos,
		"metas":       listaMetas,
		"inversiones": listaInversiones,
	}, nil
}
