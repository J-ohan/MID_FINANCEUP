package finanzas

import (
	"sort"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// VISTAS DEL ADMINISTRADOR (datos de todos los usuarios)
// GET /v1/admin/finanzas
// GET /v1/admin/libro-mayor
//
// Juntan dos CRUD: crud_finanzas (los datos) y crud_auth (el
// nombre de cada usuario). Por eso pasan por el MID.
// ============================================================

// FinanzasAdmin devuelve los movimientos, las metas y las inversiones de
// todos los usuarios, cada uno con el nombre de su dueno.
func FinanzasAdmin() (map[string]interface{}, error) {

	// 1. El nombre de cada usuario (de crud_auth)
	nombres, err := nombresUsuarios()

	if err != nil {
		return nil, err
	}

	// 2. Movimientos, metas e inversiones activos de todos (de crud_finanzas)
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

	// 3. Armar las tres listas poniendo el nombre del usuario
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

// LibroMayorAdmin arma un libro mayor por cada usuario, como la pantalla
// admin/libro-mayor: el nombre del usuario, sus filas (cada una con
// ingreso o egreso) de la mas vieja a la mas nueva, y sus totales.
func LibroMayorAdmin() ([]map[string]interface{}, error) {

	// 1. Nombres de los usuarios, movimientos de todos y categorias
	nombres, err := nombresUsuarios()

	if err != nil {
		return nil, err
	}

	var movimientos []models.Movimiento

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso?query=activo:true&sortby=Fecha&order=asc&limit=10000", &movimientos); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	// 2. Un "libro" por usuario: sus filas y sus totales
	type libro struct {
		filas    []map[string]interface{}
		ingresos float64
		egresos  float64
	}

	libros := map[int]*libro{}
	usuarios := []int{} // el orden en que aparecen los usuarios

	for _, movimiento := range movimientos {

		// Si es el primer movimiento de este usuario, se le crea su libro
		actual, existe := libros[movimiento.IdUsuario]

		if !existe {
			actual = &libro{filas: []map[string]interface{}{}}
			libros[movimiento.IdUsuario] = actual
			usuarios = append(usuarios, movimiento.IdUsuario)
		}

		nombreCategoria := ""
		if movimiento.IdCategoria != nil {
			nombreCategoria = categorias[movimiento.IdCategoria.Id].Nombre
		}

		// Cada fila separa el valor en la columna de ingreso o de egreso
		fila := map[string]interface{}{
			"id":        movimiento.Id,
			"fecha":     movimiento.Fecha.Format("2006-01-02"),
			"concepto":  movimiento.Nombre,
			"categoria": nombreCategoria,
			"ingreso":   0.0,
			"egreso":    0.0,
		}

		if movimiento.EsIngreso {
			fila["ingreso"] = movimiento.Monto
			actual.ingresos += movimiento.Monto
		} else {
			fila["egreso"] = movimiento.Monto
			actual.egresos += movimiento.Monto
		}

		actual.filas = append(actual.filas, fila)
	}

	// 3. Ordenar por Id de usuario y armar la respuesta
	sort.Ints(usuarios)

	resultado := []map[string]interface{}{}

	for _, idUsuario := range usuarios {

		actual := libros[idUsuario]

		// Los nombres son los de LibroDeUsuario en el cliente
		// (pages/admin/libro-mayor/libro-mayor.ts)
		resultado = append(resultado, map[string]interface{}{
			"usuarioId":     idUsuario,
			"nombreUsuario": nombres[idUsuario],
			"filas":         actual.filas,
			"totalIngresos": actual.ingresos,
			"totalEgresos":  actual.egresos,
			"resultado":     actual.ingresos - actual.egresos, // lo que le queda
		})
	}

	return resultado, nil
}
