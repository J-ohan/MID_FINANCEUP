package services

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE FINANZAS (crud_finanzas)
// Aqui esta la logica: sumas, porcentajes, filtros, importar,
// validar y los cambios que tocan varias tablas.
// ============================================================

// ------------------------------------------------------------
// Funciones de apoyo
// ------------------------------------------------------------

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

// categoriasPorNombre: nombre de la categoria (en minuscula) -> id.
func categoriasPorNombre(categorias map[int]models.Categoria) map[string]int {

	mapa := map[string]int{}
	for _, categoria := range categorias {
		mapa[strings.ToLower(strings.TrimSpace(categoria.Nombre))] = categoria.Id
	}

	return mapa
}

// nombresUsuarios trae de crud_auth el nombre de cada usuario por su id.
// Es privada para no chocar con NombresUsuarios del service de auth.
func nombresUsuarios() (map[int]string, error) {

	var usuarios []struct {
		Id       int    `json:"Id"`
		Nombre   string `json:"Nombre"`
		Apellido string `json:"Apellido"`
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?limit=10000", &usuarios); err != nil {
		return nil, err
	}

	nombres := map[int]string{}
	for _, usuario := range usuarios {
		nombres[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	return nombres, nil
}

// usuarioExiste revisa en crud_auth que el usuario exista.
func usuarioExiste(idUsuario int) error {

	if idUsuario <= 0 {
		return helpers.NuevoError(http.StatusBadRequest, "El id del usuario es obligatorio")
	}

	var usuario struct {
		Id int `json:"Id"`
	}

	return helpers.ConsultarUno(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, idUsuario), &usuario, "El usuario no existe")
}

// idDe saca el id de un campo que llega del CRUD, sea un numero (3)
// o una llave foranea ({"Id": 3}).
func idDe(valor interface{}) int {

	switch dato := valor.(type) {
	case float64:
		return int(dato)
	case map[string]interface{}:
		id, _ := dato["Id"].(float64)
		return int(id)
	}

	return 0
}

// revisarRegistro: el registro debe estar activo y, si llega el
// usuario, debe ser suyo.
func revisarRegistro(registro map[string]interface{}, idUsuario int, mensajeNoExiste string) error {

	if activo, _ := registro["Activo"].(bool); !activo {
		return helpers.NuevoError(http.StatusNotFound, mensajeNoExiste)
	}

	if idUsuario > 0 && idDe(registro["IdUsuario"]) != idUsuario {
		return helpers.NuevoError(http.StatusForbidden, "No puedes modificar datos de otro usuario")
	}

	return nil
}

// hoy devuelve la fecha de hoy sin hora.
func hoy() time.Time {
	fecha, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	return fecha
}

// tipoMovimiento convierte EsIngreso en "ingreso" o "gasto".
func tipoMovimiento(movimiento models.Movimiento) string {
	if movimiento.EsIngreso {
		return "ingreso"
	}
	return "gasto"
}

// armarMovimiento pasa un movimiento al formato que usa la pagina.
func armarMovimiento(movimiento models.Movimiento, categorias map[int]models.Categoria) map[string]interface{} {

	categoria := models.Categoria{}
	if movimiento.IdCategoria != nil {
		categoria = categorias[movimiento.IdCategoria.Id]
	}

	return map[string]interface{}{
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
	}
}

// saludFinanciera: "Alto riesgo" si gasta mas de lo que gana,
// "Atencion" si ahorra menos del 10 % y si no "En orden".
func saludFinanciera(balance float64, porcentajeAhorro float64) map[string]string {

	if balance < 0 {
		return map[string]string{"estado": "Alto riesgo", "mensaje": "Estas gastando mas de lo que ganas."}
	}

	if porcentajeAhorro < 10 {
		return map[string]string{"estado": "Atencion", "mensaje": "Tu margen de ahorro es bajo, revisa tus gastos."}
	}

	return map[string]string{"estado": "En orden", "mensaje": "Tus finanzas van bien, sigue asi."}
}

// ------------------------------------------------------------
// Panel principal
// ------------------------------------------------------------

// ResumenFinanciero: ingresos, gastos, balance, salud, gastos por
// categoria y los ultimos 5 movimientos.
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

	// De la categoria con mas gasto a la de menos
	sort.Slice(distribucion, func(i, j int) bool {
		return distribucion[i]["total"].(float64) > distribucion[j]["total"].(float64)
	})

	// 4. Ultimos 5 movimientos
	ultimos := []map[string]interface{}{}

	for i, movimiento := range movimientos {

		if i == 5 {
			break
		}

		ultimos = append(ultimos, armarMovimiento(movimiento, categorias))
	}

	// 5. Salud financiera segun cuanto se ahorra de lo que entra
	balance := totalIngresos - totalGastos

	porcentajeAhorro := 0.0
	if totalIngresos > 0 {
		porcentajeAhorro = math.Round(balance / totalIngresos * 100)
	}

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

// Categorias: categorias activas para los selects y filtros.
func Categorias() ([]map[string]interface{}, error) {

	var categorias []models.Categoria

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/categoria?query=activo:true&sortby=Nombre&order=asc&limit=1000", &categorias); err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}

	for _, categoria := range categorias {
		lista = append(lista, map[string]interface{}{
			"id":     categoria.Id,
			"nombre": categoria.Nombre,
			"icono":  categoria.Icono,
		})
	}

	return lista, nil
}

// ------------------------------------------------------------
// Libro mayor (movimiento_ingreso_egreso)
// ------------------------------------------------------------

// validarFiltros revisa los filtros que llegan en la URL.
func validarFiltros(filtros models.FiltrosMovimientos) error {

	if filtros.Tipo != "" && filtros.Tipo != "todos" && filtros.Tipo != "ingreso" && filtros.Tipo != "gasto" {
		return helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser todos, ingreso o gasto")
	}

	for _, fecha := range []string{filtros.Desde, filtros.Hasta} {

		if fecha == "" {
			continue
		}

		if _, err := time.Parse("2006-01-02", fecha); err != nil {
			return helpers.NuevoError(http.StatusBadRequest, "Las fechas deben tener el formato AAAA-MM-DD")
		}
	}

	return nil
}

// movimientosFiltrados trae los movimientos del usuario que cumplen
// los filtros (mas nuevos primero) junto con las categorias.
func movimientosFiltrados(idUsuario int, filtros models.FiltrosMovimientos) ([]models.Movimiento, map[int]models.Categoria, error) {

	if err := validarFiltros(filtros); err != nil {
		return nil, nil, err
	}

	// 1. Filtros que hace el CRUD
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

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, nil, err
	}

	// 2. Filtros que el CRUD no sabe hacer: rango de fechas y texto
	texto := strings.ToLower(strings.TrimSpace(filtros.Texto))
	resultado := []models.Movimiento{}

	for _, movimiento := range movimientos {

		// Las fechas AAAA-MM-DD se pueden comparar como texto
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

// LibroMayor: movimientos del usuario con nombre e icono de su categoria.
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

// nuevoMovimiento revisa una fila (del Excel o del formulario) y la
// convierte en un movimiento listo para guardar.
func nuevoMovimiento(idUsuario int, fila models.FilaExcel, categoriaPorNombre map[string]int) (models.Movimiento, error) {

	concepto := strings.TrimSpace(fila.Concepto)
	tipo := strings.ToLower(strings.TrimSpace(fila.Tipo))
	textoFecha := strings.TrimSpace(fila.Fecha)

	// La plantilla de Excel no trae fecha: se usa la de hoy
	if textoFecha == "" {
		textoFecha = time.Now().Format("2006-01-02")
	}

	fecha, err := time.Parse("2006-01-02", textoFecha)

	if err != nil {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "La fecha debe ser AAAA-MM-DD")
	}

	if concepto == "" || fila.Valor <= 0 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "Falta el concepto o el valor")
	}

	if utf8.RuneCountInString(concepto) > 120 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El concepto no puede tener mas de 120 caracteres")
	}

	if tipo != "ingreso" && tipo != "gasto" {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser ingreso o gasto")
	}

	if utf8.RuneCountInString(fila.MetodoPago) > 50 {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "El metodo de pago no puede tener mas de 50 caracteres")
	}

	idCategoria, existe := categoriaPorNombre[strings.ToLower(strings.TrimSpace(fila.Categoria))]

	if !existe {
		return models.Movimiento{}, helpers.NuevoError(http.StatusBadRequest, "La categoria \""+fila.Categoria+"\" no existe")
	}

	return models.Movimiento{
		IdUsuario:     idUsuario,
		IdCategoria:   &models.Ref{Id: idCategoria},
		Nombre:        concepto,
		Monto:         fila.Valor,
		EsIngreso:     tipo == "ingreso",
		Fecha:         fecha,
		MetodoPago:    strings.TrimSpace(fila.MetodoPago),
		Observaciones: strings.TrimSpace(fila.Observaciones),
		Activo:        true,
	}, nil
}

// CrearMovimiento: registra un ingreso o un gasto.
func CrearMovimiento(datos models.DatosMovimiento) (map[string]interface{}, error) {

	if err := usuarioExiste(datos.IdUsuario); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	movimiento, err := nuevoMovimiento(datos.IdUsuario, datos.FilaExcel, categoriasPorNombre(categorias))

	if err != nil {
		return nil, err
	}

	// El CRUD responde el registro creado con su Id
	var creado struct {
		Id int `json:"Id"`
	}

	if err := helpers.Crear(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso", movimiento, &creado); err != nil {
		return nil, err
	}

	movimiento.Id = creado.Id

	return armarMovimiento(movimiento, categorias), nil
}

// EditarMovimiento: cambia los datos de un movimiento.
func EditarMovimiento(idMovimiento int, datos models.DatosMovimiento) (map[string]interface{}, error) {

	// El registro se guarda en un map para no perder ningun campo en el PUT
	var registro map[string]interface{}
	direccion := fmt.Sprintf("%s/movimiento_ingreso_egreso/%d", helpers.UrlCrudFinanzas, idMovimiento)

	if err := helpers.ConsultarUno(direccion, &registro, "El movimiento no existe"); err != nil {
		return nil, err
	}

	if err := revisarRegistro(registro, datos.IdUsuario, "El movimiento no existe"); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	movimiento, err := nuevoMovimiento(idDe(registro["IdUsuario"]), datos.FilaExcel, categoriasPorNombre(categorias))

	if err != nil {
		return nil, err
	}

	registro["IdCategoria"] = movimiento.IdCategoria
	registro["Nombre"] = movimiento.Nombre
	registro["Monto"] = movimiento.Monto
	registro["EsIngreso"] = movimiento.EsIngreso
	registro["Fecha"] = movimiento.Fecha
	registro["MetodoPago"] = movimiento.MetodoPago
	registro["Observaciones"] = movimiento.Observaciones

	if err := helpers.Actualizar(direccion, registro); err != nil {
		return nil, err
	}

	movimiento.Id = idMovimiento

	return armarMovimiento(movimiento, categorias), nil
}

// desactivar hace el borrado logico: Activo = false.
// Los registros no se borran de verdad para no perder el historial.
func desactivar(tabla string, id int, idUsuario int, mensajeNoExiste string) error {

	var registro map[string]interface{}
	direccion := fmt.Sprintf("%s/%s/%d", helpers.UrlCrudFinanzas, tabla, id)

	if err := helpers.ConsultarUno(direccion, &registro, mensajeNoExiste); err != nil {
		return err
	}

	if err := revisarRegistro(registro, idUsuario, mensajeNoExiste); err != nil {
		return err
	}

	registro["Activo"] = false

	return helpers.Actualizar(direccion, registro)
}

// EliminarMovimiento: borra (desactiva) un movimiento.
func EliminarMovimiento(idMovimiento int, idUsuario int) (map[string]interface{}, error) {

	if err := desactivar("movimiento_ingreso_egreso", idMovimiento, idUsuario, "El movimiento no existe"); err != nil {
		return nil, err
	}

	return map[string]interface{}{"id": idMovimiento, "mensaje": "Movimiento eliminado"}, nil
}

// ------------------------------------------------------------
// Excel: importar y exportar
// ------------------------------------------------------------

// ImportarMovimientos: guarda las filas del Excel que esten bien.
// Las filas con errores no se guardan y se devuelven en "errores".
// Si el CRUD falla a la mitad, se desactivan las que ya se guardaron
// para que la importacion no quede a medias.
func ImportarMovimientos(datos models.DatosImportar) (map[string]interface{}, error) {

	if len(datos.Movimientos) == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "No llego ningun movimiento")
	}

	if err := usuarioExiste(datos.IdUsuario); err != nil {
		return nil, err
	}

	categorias, err := categoriasPorId()

	if err != nil {
		return nil, err
	}

	categoriaPorNombre := categoriasPorNombre(categorias)

	// 1. Revisar todas las filas antes de guardar
	validos := []models.Movimiento{}
	errores := []string{}

	for i, fila := range datos.Movimientos {

		movimiento, err := nuevoMovimiento(datos.IdUsuario, fila, categoriaPorNombre)

		if err != nil {
			errores = append(errores, fmt.Sprintf("Fila %d: %s", i+1, err.Error()))
			continue
		}

		validos = append(validos, movimiento)
	}

	// 2. Guardar las filas buenas; si una falla, deshacer las anteriores
	guardados := []models.Movimiento{}

	for _, movimiento := range validos {

		var creado struct {
			Id int `json:"Id"`
		}

		if err := helpers.Crear(helpers.UrlCrudFinanzas+"/movimiento_ingreso_egreso", movimiento, &creado); err != nil {

			for _, anterior := range guardados {
				anterior.Activo = false
				helpers.Actualizar(fmt.Sprintf("%s/movimiento_ingreso_egreso/%d", helpers.UrlCrudFinanzas, anterior.Id), anterior)
			}

			return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo guardar \""+movimiento.Nombre+"\", no se importo ninguna fila")
		}

		movimiento.Id = creado.Id
		guardados = append(guardados, movimiento)
	}

	return map[string]interface{}{"guardados": len(guardados), "errores": errores}, nil
}

// ExportarMovimientos: datos de la plantilla "Control de tus finanzas":
// ingresos, gastos fijos y gastos hormiga (concepto y valor), sus
// totales y el resumen con las mismas formulas del Excel.
func ExportarMovimientos(idUsuario int, filtros models.FiltrosMovimientos) (map[string]interface{}, error) {

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

	// Resumen: 10 % de ahorro sugerido, gasto total, balance del mes
	// y lo que queda despues de apartar el ahorro
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

// ------------------------------------------------------------
// Metas (meta)
// ------------------------------------------------------------

// armarMeta: % de avance (maximo 100) y si ya esta cumplida.
func armarMeta(meta models.Meta) map[string]interface{} {

	// porcentaje = ahorrado / objetivo * 100 (maximo 100)
	porcentaje := 0.0
	if meta.MontoObjetivo > 0 {
		porcentaje = math.Min(100, math.Round(meta.MontoActual/meta.MontoObjetivo*100))
	}

	return map[string]interface{}{
		"id":         meta.Id,
		"nombre":     meta.Nombre,
		"icono":      meta.Icono,
		"actual":     meta.MontoActual,
		"objetivo":   meta.MontoObjetivo,
		"porcentaje": porcentaje,
		"cumplida":   meta.MontoObjetivo > 0 && meta.MontoActual >= meta.MontoObjetivo,
	}
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

		item := armarMeta(meta)

		if item["cumplida"].(bool) {
			cumplidas++
		}

		totalAhorrado += meta.MontoActual
		totalObjetivo += meta.MontoObjetivo

		lista = append(lista, item)
	}

	return map[string]interface{}{
		"metas":          lista,
		"total_ahorrado": totalAhorrado,
		"total_objetivo": totalObjetivo,
		"cumplidas":      cumplidas,
	}, nil
}

// validarMeta revisa y limpia los datos de una meta.
func validarMeta(datos *models.DatosMeta) error {

	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Icono = strings.TrimSpace(datos.Icono)

	if datos.Nombre == "" {
		return helpers.NuevoError(http.StatusBadRequest, "El nombre de la meta es obligatorio")
	}

	if utf8.RuneCountInString(datos.Nombre) > 120 {
		return helpers.NuevoError(http.StatusBadRequest, "El nombre no puede tener mas de 120 caracteres")
	}

	if datos.Objetivo <= 0 {
		return helpers.NuevoError(http.StatusBadRequest, "El objetivo debe ser mayor a 0")
	}

	if datos.Actual < 0 {
		return helpers.NuevoError(http.StatusBadRequest, "Lo ahorrado no puede ser negativo")
	}

	// Si no se elige icono se usa el de diana
	if datos.Icono == "" {
		datos.Icono = "🎯"
	}

	if utf8.RuneCountInString(datos.Icono) > 50 {
		return helpers.NuevoError(http.StatusBadRequest, "El icono no puede tener mas de 50 caracteres")
	}

	return nil
}

// CrearMeta: crea una meta de ahorro.
func CrearMeta(datos models.DatosMeta) (map[string]interface{}, error) {

	if err := usuarioExiste(datos.IdUsuario); err != nil {
		return nil, err
	}

	if err := validarMeta(&datos); err != nil {
		return nil, err
	}

	meta := models.Meta{
		IdUsuario:     datos.IdUsuario,
		Nombre:        datos.Nombre,
		Icono:         datos.Icono,
		MontoObjetivo: datos.Objetivo,
		MontoActual:   datos.Actual,
		Activo:        true,
	}

	var creada struct {
		Id int `json:"Id"`
	}

	if err := helpers.Crear(helpers.UrlCrudFinanzas+"/meta", meta, &creada); err != nil {
		return nil, err
	}

	meta.Id = creada.Id

	return armarMeta(meta), nil
}

// EditarMeta: cambia nombre, icono, lo ahorrado y el objetivo.
func EditarMeta(idMeta int, datos models.DatosMeta) (map[string]interface{}, error) {

	if err := validarMeta(&datos); err != nil {
		return nil, err
	}

	var registro map[string]interface{}
	direccion := fmt.Sprintf("%s/meta/%d", helpers.UrlCrudFinanzas, idMeta)

	if err := helpers.ConsultarUno(direccion, &registro, "La meta no existe"); err != nil {
		return nil, err
	}

	if err := revisarRegistro(registro, datos.IdUsuario, "La meta no existe"); err != nil {
		return nil, err
	}

	registro["Nombre"] = datos.Nombre
	registro["Icono"] = datos.Icono
	registro["MontoActual"] = datos.Actual
	registro["MontoObjetivo"] = datos.Objetivo

	if err := helpers.Actualizar(direccion, registro); err != nil {
		return nil, err
	}

	return armarMeta(models.Meta{
		Id:            idMeta,
		Nombre:        datos.Nombre,
		Icono:         datos.Icono,
		MontoActual:   datos.Actual,
		MontoObjetivo: datos.Objetivo,
	}), nil
}

// EliminarMeta: borra (desactiva) una meta.
func EliminarMeta(idMeta int, idUsuario int) (map[string]interface{}, error) {

	if err := desactivar("meta", idMeta, idUsuario, "La meta no existe"); err != nil {
		return nil, err
	}

	return map[string]interface{}{"id": idMeta, "mensaje": "Meta eliminada"}, nil
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

// ------------------------------------------------------------
// Inversiones (inversion + historial_inversion)
// ------------------------------------------------------------

// nivelesRiesgo: nombre en minuscula -> id, e id -> nombre.
func nivelesRiesgo() (map[string]int, map[int]string, error) {

	var niveles []models.NivelRiesgo

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/nivel_riesgo?limit=100", &niveles); err != nil {
		return nil, nil, err
	}

	porNombre := map[string]int{}
	porId := map[int]string{}

	for _, nivel := range niveles {
		porNombre[strings.ToLower(nivel.Nombre)] = nivel.Id
		porId[nivel.Id] = nivel.Nombre
	}

	return porNombre, porId, nil
}

// armarInversion pasa una inversion al formato que usa la pagina.
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

// InversionesUsuario: inversiones con su riesgo, balance, rendimiento y grafica.
func InversionesUsuario(idUsuario int) (map[string]interface{}, error) {

	// 1. Inversiones y niveles de riesgo
	var inversiones []models.Inversion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/inversion?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &inversiones); err != nil {
		return nil, err
	}

	_, riesgoPorId, err := nivelesRiesgo()

	if err != nil {
		return nil, err
	}

	// 2. Armar la lista, sumar y juntar el historial por fecha
	lista := []map[string]interface{}{}
	balance := 0.0
	rendimientoTotal := 0.0
	valorPorFecha := map[string]float64{}

	for _, inversion := range inversiones {

		balance += inversion.Monto
		rendimientoTotal += inversion.Rendimiento

		lista = append(lista, armarInversion(inversion, riesgoPorId))

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

// validarInversion revisa y limpia los datos de una inversion.
// Devuelve el id del nivel de riesgo.
func validarInversion(datos *models.DatosInversion, riesgoPorNombre map[string]int) (int, error) {

	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Duracion = strings.TrimSpace(datos.Duracion)

	if datos.Nombre == "" || datos.Duracion == "" {
		return 0, helpers.NuevoError(http.StatusBadRequest, "El nombre y la duracion son obligatorios")
	}

	if utf8.RuneCountInString(datos.Nombre) > 120 {
		return 0, helpers.NuevoError(http.StatusBadRequest, "El nombre no puede tener mas de 120 caracteres")
	}

	if utf8.RuneCountInString(datos.Duracion) > 30 {
		return 0, helpers.NuevoError(http.StatusBadRequest, "La duracion no puede tener mas de 30 caracteres")
	}

	// El rendimiento si puede ser negativo (perdida)
	if datos.Monto <= 0 {
		return 0, helpers.NuevoError(http.StatusBadRequest, "El monto debe ser mayor a 0")
	}

	idRiesgo, existe := riesgoPorNombre[strings.ToLower(strings.TrimSpace(datos.Riesgo))]

	if !existe {
		return 0, helpers.NuevoError(http.StatusBadRequest, "El riesgo debe ser Bajo, Medio o Alto")
	}

	return idRiesgo, nil
}

// guardarPuntoHistorial deja en historial_inversion el valor de hoy.
// Si ya hay un punto de hoy lo actualiza (solo puede haber uno por dia).
func guardarPuntoHistorial(idInversion int, valor float64) error {

	var historial []models.HistorialInversion
	helpers.ConsultarLista(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idInversion), &historial)

	fechaHoy := hoy()

	for _, punto := range historial {

		if punto.Fecha.Format("2006-01-02") == fechaHoy.Format("2006-01-02") {
			punto.Valor = valor
			return helpers.Actualizar(fmt.Sprintf("%s/historial_inversion/%d", helpers.UrlCrudFinanzas, punto.Id), punto)
		}
	}

	punto := models.HistorialInversion{
		IdInversion: &models.Ref{Id: idInversion},
		Fecha:       fechaHoy,
		Valor:       valor,
		Activo:      true,
	}

	return helpers.Crear(helpers.UrlCrudFinanzas+"/historial_inversion", punto, nil)
}

// CrearInversion: guarda la inversion y su primer punto en el historial
// (para la grafica de crecimiento). Si el historial falla, se deshace.
func CrearInversion(datos models.DatosInversion) (map[string]interface{}, error) {

	if err := usuarioExiste(datos.IdUsuario); err != nil {
		return nil, err
	}

	riesgoPorNombre, riesgoPorId, err := nivelesRiesgo()

	if err != nil {
		return nil, err
	}

	idRiesgo, err := validarInversion(&datos, riesgoPorNombre)

	if err != nil {
		return nil, err
	}

	inversion := models.Inversion{
		IdUsuario:     datos.IdUsuario,
		IdNivelRiesgo: &models.Ref{Id: idRiesgo},
		Nombre:        datos.Nombre,
		Monto:         datos.Monto,
		Rendimiento:   datos.Rendimiento,
		Duracion:      datos.Duracion,
		FechaInicio:   hoy(),
		Activo:        true,
	}

	var creada struct {
		Id int `json:"Id"`
	}

	if err := helpers.Crear(helpers.UrlCrudFinanzas+"/inversion", inversion, &creada); err != nil {
		return nil, err
	}

	inversion.Id = creada.Id

	// Valor actual de la inversion = lo invertido + lo que ha ganado
	if err := guardarPuntoHistorial(inversion.Id, inversion.Monto+inversion.Rendimiento); err != nil {

		inversion.Activo = false
		helpers.Actualizar(fmt.Sprintf("%s/inversion/%d", helpers.UrlCrudFinanzas, inversion.Id), inversion)

		return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo guardar el historial, no se creo la inversion")
	}

	return armarInversion(inversion, riesgoPorId), nil
}

// EditarInversion: cambia los datos de la inversion y deja el valor
// de hoy en el historial. Si el historial falla, se deshace el cambio.
func EditarInversion(idInversion int, datos models.DatosInversion) (map[string]interface{}, error) {

	riesgoPorNombre, riesgoPorId, err := nivelesRiesgo()

	if err != nil {
		return nil, err
	}

	idRiesgo, err := validarInversion(&datos, riesgoPorNombre)

	if err != nil {
		return nil, err
	}

	var registro map[string]interface{}
	direccion := fmt.Sprintf("%s/inversion/%d", helpers.UrlCrudFinanzas, idInversion)

	if err := helpers.ConsultarUno(direccion, &registro, "La inversion no existe"); err != nil {
		return nil, err
	}

	if err := revisarRegistro(registro, datos.IdUsuario, "La inversion no existe"); err != nil {
		return nil, err
	}

	// Copia de como estaba, por si hay que deshacer
	anterior := map[string]interface{}{}
	for campo, valor := range registro {
		anterior[campo] = valor
	}

	registro["Nombre"] = datos.Nombre
	registro["Monto"] = datos.Monto
	registro["Rendimiento"] = datos.Rendimiento
	registro["Duracion"] = datos.Duracion
	registro["IdNivelRiesgo"] = models.Ref{Id: idRiesgo}

	if err := helpers.Actualizar(direccion, registro); err != nil {
		return nil, err
	}

	if err := guardarPuntoHistorial(idInversion, datos.Monto+datos.Rendimiento); err != nil {

		helpers.Actualizar(direccion, anterior)

		return nil, helpers.NuevoError(http.StatusBadGateway, "No se pudo guardar el historial, no se cambio la inversion")
	}

	return armarInversion(models.Inversion{
		Id:            idInversion,
		IdNivelRiesgo: &models.Ref{Id: idRiesgo},
		Nombre:        datos.Nombre,
		Monto:         datos.Monto,
		Rendimiento:   datos.Rendimiento,
		Duracion:      datos.Duracion,
	}, riesgoPorId), nil
}

// EliminarInversion: desactiva primero su historial y despues la inversion.
func EliminarInversion(idInversion int, idUsuario int) (map[string]interface{}, error) {

	var registro map[string]interface{}
	direccion := fmt.Sprintf("%s/inversion/%d", helpers.UrlCrudFinanzas, idInversion)

	if err := helpers.ConsultarUno(direccion, &registro, "La inversion no existe"); err != nil {
		return nil, err
	}

	if err := revisarRegistro(registro, idUsuario, "La inversion no existe"); err != nil {
		return nil, err
	}

	// 1. Historial de la inversion
	var historial []models.HistorialInversion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idInversion), &historial); err != nil {
		return nil, err
	}

	for _, punto := range historial {

		punto.Activo = false

		if err := helpers.Actualizar(fmt.Sprintf("%s/historial_inversion/%d", helpers.UrlCrudFinanzas, punto.Id), punto); err != nil {
			return nil, err
		}
	}

	// 2. La inversion
	registro["Activo"] = false

	if err := helpers.Actualizar(direccion, registro); err != nil {
		return nil, err
	}

	return map[string]interface{}{"id": idInversion, "mensaje": "Inversion eliminada"}, nil
}

// ------------------------------------------------------------
// Admin (datos de todos los usuarios)
// ------------------------------------------------------------

// FinanzasAdmin: movimientos, metas e inversiones de todos los usuarios
// con el nombre de cada uno (crud_auth + crud_finanzas).
func FinanzasAdmin() (map[string]interface{}, error) {

	// 1. Nombres de los usuarios
	nombres, err := nombresUsuarios()

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

// LibroMayorAdmin: el libro mayor de cada usuario con su nombre.
// Cada fila separa ingreso y egreso, de la mas antigua a la mas nueva.
func LibroMayorAdmin() ([]map[string]interface{}, error) {

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

	// Un libro por usuario
	type libro struct {
		filas    []map[string]interface{}
		ingresos float64
		egresos  float64
	}

	libros := map[int]*libro{}
	usuarios := []int{}

	for _, movimiento := range movimientos {

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

	sort.Ints(usuarios)

	resultado := []map[string]interface{}{}

	for _, idUsuario := range usuarios {

		actual := libros[idUsuario]

		resultado = append(resultado, map[string]interface{}{
			"id_usuario":     idUsuario,
			"nombre_usuario": nombres[idUsuario],
			"filas":          actual.filas,
			"total_ingresos": actual.ingresos,
			"total_egresos":  actual.egresos,
			"resultado":      actual.ingresos - actual.egresos,
		})
	}

	return resultado, nil
}
