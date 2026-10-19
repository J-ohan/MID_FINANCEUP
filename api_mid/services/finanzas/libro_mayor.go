package finanzas

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// LIBRO MAYOR (tabla movimiento_ingreso_egreso)
// GET /v1/movimientos/usuario/:id
//
// Crear, editar o borrar UN movimiento no esta aqui: la pagina
// lo hace directo en crud_finanzas (POST/PUT/DELETE
// /v1/movimiento_ingreso_egreso), como dice la guia.
// ============================================================

// validarFiltros revisa que los filtros que llegan en la URL tengan
// sentido antes de usarlos. Si alguno esta mal, devuelve un error 400.
func validarFiltros(filtros models.FiltrosMovimientos) error {

	if filtros.Tipo != "" && filtros.Tipo != "todos" && filtros.Tipo != "ingreso" && filtros.Tipo != "gasto" {
		return helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser todos, ingreso o gasto")
	}

	for _, fecha := range []string{filtros.Desde, filtros.Hasta} {

		if fecha == "" {
			continue // la fecha es opcional
		}

		if _, err := time.Parse("2006-01-02", fecha); err != nil {
			return helpers.NuevoError(http.StatusBadRequest, "Las fechas deben tener el formato AAAA-MM-DD")
		}
	}

	return nil
}

// movimientosFiltrados trae los movimientos del usuario que cumplen los
// filtros (del mas nuevo al mas viejo) junto con las categorias.
// La usan el libro mayor y la exportacion a Excel.
func movimientosFiltrados(idUsuario int, filtros models.FiltrosMovimientos) ([]models.Movimiento, map[int]models.Categoria, error) {

	if err := validarFiltros(filtros); err != nil {
		return nil, nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, nil, err
	}

	// La pagina puede filtrar por el nombre de la categoria ("Vivienda")
	// en vez del Id: aqui se busca su Id.
	nombre := strings.ToLower(strings.TrimSpace(filtros.Categoria))

	if filtros.IdCategoria == 0 && nombre != "" && nombre != "todas" {

		idCategoria, existe := categoriasPorNombre(categorias)[nombre]

		if !existe {
			return nil, nil, helpers.NuevoError(http.StatusBadRequest, "La categoria \""+filtros.Categoria+"\" no existe")
		}

		filtros.IdCategoria = idCategoria
	}

	// 1. Los filtros que el CRUD sabe hacer van en el "query" de la URL:
	//    usuario, activos, tipo (es_ingreso) y categoria.
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

	var movimientos []models.Movimiento

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso?query="+filtro+"&sortby=Fecha&order=desc&limit=10000", &movimientos); err != nil {
		return nil, nil, err
	}

	// 2. Los filtros que el CRUD NO sabe hacer se hacen aqui:
	//    rango de fechas (desde / hasta) y buscar un texto en el concepto.
	texto := strings.ToLower(strings.TrimSpace(filtros.Texto))
	resultado := []models.Movimiento{}

	for _, movimiento := range movimientos {

		// Las fechas en formato AAAA-MM-DD se pueden comparar como texto:
		// "2026-10-05" es menor que "2026-10-20"
		fecha := movimiento.Fecha.Format("2006-01-02")

		if filtros.Desde != "" && fecha < filtros.Desde {
			continue
		}

		if filtros.Hasta != "" && fecha > filtros.Hasta {
			continue
		}

		if texto != "" && !strings.Contains(strings.ToLower(movimiento.Nombre), texto) {
			continue
		}

		resultado = append(resultado, movimiento)
	}

	return resultado, categorias, nil
}

// LibroMayor devuelve la lista de movimientos del usuario, ya filtrada y
// en el formato de la pagina (con el nombre y el icono de su categoria).
func LibroMayor(idUsuario int, filtros models.FiltrosMovimientos) ([]map[string]interface{}, error) {

	movimientos, categorias, err := movimientosFiltrados(idUsuario, filtros)

	if err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}

	for _, movimiento := range movimientos {
		lista = append(lista, armarMovimiento(movimiento, categorias))
	}

	return lista, nil
}

// nuevoMovimiento revisa una fila del Excel y, si esta bien, la convierte
// en un movimiento listo para mandarle al CRUD. Si algo esta mal devuelve
// un error 400 diciendo que.
func nuevoMovimiento(idUsuario int, fila models.FilaExcel, categoriaPorNombre map[string]int) (models.Movimiento, error) {

	// Quitar espacios de mas y pasar el tipo a minuscula ("Ingreso" -> "ingreso")
	concepto := strings.TrimSpace(fila.Concepto)
	tipo := strings.ToLower(strings.TrimSpace(fila.Tipo))

	// 1. La fecha debe venir como AAAA-MM-DD
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(fila.Fecha))

	if err != nil {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "La fecha debe ser AAAA-MM-DD")
	}

	// 2. Concepto y valor son obligatorios
	if concepto == "" || fila.Valor <= 0 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "Falta el concepto o el valor")
	}

	// 3. Los mismos limites que tiene la tabla en la base de datos
	if utf8.RuneCountInString(concepto) > 120 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El concepto no puede tener mas de 120 caracteres")
	}

	if utf8.RuneCountInString(fila.MetodoPago) > 50 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El metodo de pago no puede tener mas de 50 caracteres")
	}

	// 4. El tipo solo puede ser ingreso o gasto
	if tipo != "ingreso" && tipo != "gasto" {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser ingreso o gasto")
	}

	movimiento := models.Movimiento{
		IdUsuario:     idUsuario,
		Nombre:        concepto,
		Monto:         fila.Valor,
		EsIngreso:     tipo == "ingreso",
		Fecha:         fecha,
		MetodoPago:    strings.TrimSpace(fila.MetodoPago),
		Observaciones: strings.TrimSpace(fila.Observaciones),
		Activo:        true,
	}

	// 5. La categoria es opcional: se busca por su nombre y, si se
	//    encuentra, se le pone su Id. Si no, el movimiento queda sin categoria.
	if idCategoria, existe := categoriaPorNombre[strings.ToLower(strings.TrimSpace(fila.Categoria))]; existe {
		movimiento.IdCategoria = &models.Ref{Id: idCategoria}
	}

	return movimiento, nil
}
