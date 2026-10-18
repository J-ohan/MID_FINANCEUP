package finanzas

import (
	"fmt"
	"net/http"
	"strings"

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
		return nil, helpers.NuevoError(http.StatusBadRequest, "El modo debe ser validar o confirmar")
	}

	// 2. Debe llegar al menos una fila y el usuario debe existir
	if len(datos.Movimientos) == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "No llego ningun movimiento")
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

			return 0, helpers.NuevoError(http.StatusBadGateway, "No se pudo guardar \""+movimiento.Nombre+"\", no se importo ninguna fila")
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
