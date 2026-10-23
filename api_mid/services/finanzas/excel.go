package finanzas

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// EXCEL (plantilla "Control de tus finanzas")
// POST /v1/movimientos/importar
// GET  /v1/movimientos/usuario/:id/exportar
//
// El archivo .xlsx NO llega al MID: la pagina lo lee en el
// navegador y manda las filas en JSON. Al exportar pasa lo mismo
// al reves: el MID manda los datos y la pagina arma el archivo.
// ============================================================

// ImportarMovimientos recibe las filas del Excel y las revisa una por una.
//   - modo "validar": solo revisa, NO guarda nada. Sirve para mostrarle
//     al usuario que filas tienen errores antes de confirmar.
//   - modo "confirmar" (o sin modo): guarda las filas que estan bien.
//
// Las filas con errores nunca se guardan y se devuelven en "errores".
func ImportarMovimientos(datos models.DatosImportar) (map[string]interface{}, error) {

	// 1. Revisar el modo (si no llega, es "confirmar")
	modo := strings.ToLower(strings.TrimSpace(datos.Modo))

	if modo == "" {
		modo = "confirmar"
	}

	if modo != "validar" && modo != "confirmar" {
		return nil, errModoImportar
	}

	// 2. Debe llegar entre 1 y 1000 filas, y el usuario debe existir
	if len(datos.Movimientos) == 0 {
		return nil, errSinMovimientos
	}

	if len(datos.Movimientos) > LimiteFilasExcel {
		return nil, errDemasiadasFilas
	}

	if err := usuarioExiste(datos.IdUsuario); err != nil {
		return nil, err
	}

	// 3. Las categorias por nombre, porque el Excel trae el nombre
	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	// 4. Separar las filas buenas de las que tienen errores
	validos, errores := revisarFilas(datos.IdUsuario, datos.Movimientos, categoriasPorNombre(categorias))

	resultado := map[string]interface{}{
		"modo":      modo,
		"validos":   len(validos),
		"guardados": 0,
		"errores":   errores,
	}

	// 5. En modo "validar" se termina aqui, sin guardar nada
	if modo == "validar" {
		return resultado, nil
	}

	// 6. En modo "confirmar" se guardan las filas buenas
	guardados, err := guardarMovimientos(validos)

	if err != nil {
		return nil, err
	}

	resultado["guardados"] = guardados

	return resultado, nil
}

// revisarFilas pasa cada fila por nuevoMovimiento (que la valida).
// Las buenas quedan en "validos" y las malas en "errores", con su
// numero de fila (empieza en 1) y el motivo.
func revisarFilas(idUsuario int, filas []models.FilaExcel, categoriaPorNombre map[string]int) ([]models.Movimiento, []models.ErrorFila) {

	validos := []models.Movimiento{}
	errores := []models.ErrorFila{}

	for i, fila := range filas {

		movimiento, err := nuevoMovimiento(idUsuario, fila, categoriaPorNombre)

		if err != nil {
			errores = append(errores, models.ErrorFila{
				Fila:     i + 1,
				Concepto: strings.TrimSpace(fila.Concepto),
				Motivo:   err.Error(),
			})
			continue
		}

		validos = append(validos, movimiento)
	}

	return validos, errores
}

// guardarMovimientos manda los movimientos al CRUD uno por uno.
// Si el CRUD falla a la mitad (por ejemplo, se apago), desactiva los que
// ya se habian guardado. Asi la importacion no queda a medias: o se
// guardan todas las filas buenas, o ninguna.
func guardarMovimientos(movimientos []models.Movimiento) (int, error) {

	guardados := []models.Movimiento{}

	for _, movimiento := range movimientos {

		// El CRUD responde el registro creado; de ahi se saca su Id
		var creado struct {
			Id int `json:"Id"`
		}

		if err := helpers.Crear(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso", movimiento, &creado); err != nil {

			// Deshacer: los que ya se guardaron quedan como borrados (Activo = false)
			for _, anterior := range guardados {
				anterior.Activo = false
				helpers.Actualizar(fmt.Sprintf("%s/movimiento_ingreso_egreso/%d", helpers.UrlCrudFinanzas, anterior.Id), anterior)
			}

			return 0, errNoSeImporto(movimiento.Nombre)
		}

		movimiento.Id = creado.Id
		guardados = append(guardados, movimiento)
	}

	return len(guardados), nil
}

// ExportarMovimientos arma los datos de la plantilla "Control de tus
// finanzas": tres columnas (ingresos, gastos fijos y gastos hormiga) con
// concepto y valor, el total de cada una y el resumen con las mismas
// formulas del Excel. Acepta los mismos filtros del libro mayor.
func ExportarMovimientos(idUsuario int, filtros models.FiltrosMovimientos) (map[string]interface{}, error) {

	// 1. Los movimientos filtrados, igual que en el libro mayor
	movimientos, categorias, err := movimientosFiltrados(idUsuario, filtros)

	if err != nil {
		return nil, err
	}

	ingresos := []map[string]interface{}{}
	gastosFijos := []map[string]interface{}{}
	gastosHormiga := []map[string]interface{}{}

	totalIngresos := 0.0
	totalFijos := 0.0
	totalHormiga := 0.0

	// 2. Repartir cada movimiento en su columna:
	//    ingreso -> Ingresos; gasto de "Gastos hormiga" -> Gastos hormiga;
	//    cualquier otro gasto -> Gastos fijos.
	for _, movimiento := range movimientos {

		fila := map[string]interface{}{"concepto": movimiento.Nombre, "valor": movimiento.Monto}

		nombreCategoria := ""
		if movimiento.IdCategoria != nil {
			nombreCategoria = strings.ToLower(categorias[movimiento.IdCategoria.Id].Nombre)
		}

		if movimiento.EsIngreso {
			ingresos = append(ingresos, fila)
			totalIngresos += movimiento.Monto
		} else if nombreCategoria == "gastos hormiga" {
			gastosHormiga = append(gastosHormiga, fila)
			totalHormiga += movimiento.Monto
		} else {
			gastosFijos = append(gastosFijos, fila)
			totalFijos += movimiento.Monto
		}
	}

	// 3. El resumen, con las mismas formulas de la plantilla:
	//    ahorro sugerido = 10 % de los ingresos
	//    balance del mes = ingresos - todos los gastos
	//    ahorro mensual  = balance - ahorro sugerido
	proyeccionAhorro := totalIngresos * 0.1
	totalGastos := totalFijos + totalHormiga
	balanceMes := totalIngresos - totalGastos

	return map[string]interface{}{
		"ingresos":             ingresos,
		"gastos_fijos":         gastosFijos,
		"gastos_hormiga":       gastosHormiga,
		"total_ingresos":       totalIngresos,
		"total_gastos_fijos":   totalFijos,
		"total_gastos_hormiga": totalHormiga,
		"proyeccion_ahorro":    proyeccionAhorro,
		"total_gastos":         totalGastos,
		"balance_mes":          balanceMes,
		"ahorro_mensual":       balanceMes - proyeccionAhorro,
	}, nil
}

// ExportarInversiones arma los datos del Excel "Mis inversiones":
//   - resumen: total invertido, rendimiento, valor actual, rentabilidad
//     y cuantas inversiones hay
//   - por_riesgo: cuanto dinero hay en Bajo, Medio y Alto, y que % del total es
//   - inversiones: el detalle de cada una con su rentabilidad, su valor
//     actual y que parte de la cartera es (de la mas grande a la mas pequena)
//   - crecimiento: los puntos del historial (los de la grafica)
//   - alerta: si la mitad o mas del dinero esta en riesgo Alto
//     (la misma regla que muestra la pagina)
//
// Usa cargarInversiones, asi el Excel muestra lo mismo que la pantalla.
func ExportarInversiones(idUsuario int) (map[string]interface{}, error) {

	datos, err := cargarInversiones(idUsuario)

	if err != nil {
		return nil, err
	}

	balance, rendimiento := datos.totales()

	// 1. Ordenar de la inversion mas grande a la mas pequena
	ordenadas := append([]models.Inversion{}, datos.inversiones...)
	sort.SliceStable(ordenadas, func(i, j int) bool { return ordenadas[i].Monto > ordenadas[j].Monto })

	// 2. Detalle de cada inversion y cuanto hay en cada riesgo
	detalle := []map[string]interface{}{}
	montoPorRiesgo := map[string]float64{}
	cantidadPorRiesgo := map[string]int{}

	for _, inversion := range ordenadas {

		riesgo := nombreRiesgo(inversion, datos.riesgoPorId)
		if riesgo == "" {
			riesgo = "Sin definir"
		}

		montoPorRiesgo[riesgo] += inversion.Monto
		cantidadPorRiesgo[riesgo]++

		detalle = append(detalle, map[string]interface{}{
			"nombre":        inversion.Nombre,
			"monto":         inversion.Monto,
			"rendimiento":   inversion.Rendimiento,
			"rentabilidad":  porcentajeDecimal(inversion.Rendimiento, inversion.Monto), // % que gano sobre lo invertido
			"valor_actual":  inversion.Monto + inversion.Rendimiento,
			"participacion": porcentajeDecimal(inversion.Monto, balance), // % de toda la cartera
			"riesgo":        riesgo,
			"duracion":      inversion.Duracion,
		})
	}

	// 3. Reparto por riesgo: siempre Bajo, Medio y Alto (aunque esten en 0)
	porRiesgo := []map[string]interface{}{}

	for _, riesgo := range []string{"Bajo", "Medio", "Alto", "Sin definir"} {

		if riesgo == "Sin definir" && cantidadPorRiesgo[riesgo] == 0 {
			continue
		}

		porRiesgo = append(porRiesgo, map[string]interface{}{
			"riesgo":     riesgo,
			"monto":      montoPorRiesgo[riesgo],
			"porcentaje": porcentajeDecimal(montoPorRiesgo[riesgo], balance),
			"cantidad":   cantidadPorRiesgo[riesgo],
		})
	}

	// 4. Alerta si la mitad o mas del dinero esta en riesgo Alto
	alerta := ""

	if balance > 0 {
		if enRiesgoAlto := math.Round(montoPorRiesgo["Alto"] / balance * 100); enRiesgoAlto >= 50 {
			alerta = fmt.Sprintf("El %.0f%% de tu dinero invertido esta en riesgo alto. Considera diversificar en opciones de menor riesgo.", enRiesgoAlto)
		}
	}

	resultado := map[string]interface{}{
		"fecha_exportacion": time.Now().Format("2006-01-02 15:04"),
		"resumen": map[string]interface{}{
			"total_invertido":    balance,
			"rendimiento_total":  rendimiento,
			"valor_actual":       balance + rendimiento,
			"rentabilidad":       porcentajeDecimal(rendimiento, balance),
			"numero_inversiones": len(datos.inversiones),
		},
		"por_riesgo":  porRiesgo,
		"inversiones": detalle,
		"crecimiento": datos.grafica(),
		"alerta":      alerta,
	}

	if len(datos.avisos) > 0 {
		resultado["avisos"] = datos.avisos
	}

	return resultado, nil
}
