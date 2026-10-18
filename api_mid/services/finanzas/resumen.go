package finanzas

import (
	"fmt"
	"math"
	"sort"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// PANEL PRINCIPAL (pantalla Finanzas)
// GET /v1/resumen-financiero/:id
// ============================================================

// ResumenFinanciero arma todo lo que muestra el panel de un usuario:
// sus ingresos, sus gastos, el balance, la salud financiera, cuanto
// gasto en cada categoria (para la grafica de dona) y sus ultimos
// 5 movimientos.
func ResumenFinanciero(idUsuario int) (map[string]interface{}, error) {

	// 1. Pedir al CRUD los movimientos activos del usuario,
	//    del mas nuevo al mas viejo, y todas las categorias.
	var movimientos []models.Movimiento

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/movimiento_ingreso_egreso?query=id_usuario:%d,activo:true&sortby=Fecha&order=desc&limit=10000", helpers.UrlCrudFinanzas, idUsuario), &movimientos); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	// 2. Recorrer los movimientos sumando ingresos, gastos y
	//    cuanto se gasto en cada categoria.
	totalIngresos := 0.0
	totalGastos := 0.0
	gastoPorCategoria := map[int]float64{}

	for _, movimiento := range movimientos {

		if movimiento.EsIngreso {
			totalIngresos += movimiento.Monto
			continue // un ingreso no cuenta como gasto de ninguna categoria
		}

		totalGastos += movimiento.Monto

		if movimiento.IdCategoria != nil {
			gastoPorCategoria[movimiento.IdCategoria.Id] += movimiento.Monto
		}
	}

	// 3. Para cada categoria, que porcentaje es de todo lo gastado.
	//    Ejemplo: si gasto 300 en Vivienda de 1000 en total, es el 30 %.
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

	// Se ordena de la categoria con mas gasto a la de menos
	sort.Slice(distribucion, func(i, j int) bool {
		return distribucion[i]["total"].(float64) > distribucion[j]["total"].(float64)
	})

	// 4. Los ultimos 5 movimientos (ya vienen del mas nuevo al mas viejo)
	ultimos := []map[string]interface{}{}

	for i, movimiento := range movimientos {

		if i == 5 {
			break
		}

		ultimos = append(ultimos, armarMovimiento(movimiento, categorias))
	}

	// 5. Balance (lo que queda) y que porcentaje de lo que gana esta ahorrando
	balance := totalIngresos - totalGastos

	porcentajeAhorro := 0.0
	if totalIngresos > 0 {
		porcentajeAhorro = math.Round(balance / totalIngresos * 100)
	}

	// 6. Respuesta para la pagina
	return map[string]interface{}{
		"id_usuario":          idUsuario,
		"total_ingresos":      totalIngresos,
		"total_gastos":        totalGastos,
		"balance":             balance,
		"porcentaje_ahorro":   porcentajeAhorro,
		"salud":               saludFinanciera(balance, porcentajeAhorro),
		"gastos_categoria":    distribucion,
		"ultimos_movimientos": ultimos,
	}, nil
}
