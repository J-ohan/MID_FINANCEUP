package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type Categoria struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
	Icono  string `json:"Icono"`
}

type Movimiento struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdCategoria   *Ref      `json:"IdCategoria"`
	Nombre        string    `json:"Nombre"`
	Monto         float64   `json:"Monto"`
	EsIngreso     bool      `json:"EsIngreso"`
	Fecha         time.Time `json:"Fecha"`
	MetodoPago    string    `json:"MetodoPago"`
	Observaciones string    `json:"Observaciones"`
	Activo        bool      `json:"Activo"`
}

type Meta struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	Nombre        string    `json:"Nombre"`
	Descripcion   string    `json:"Descripcion"`
	MontoObjetivo float64   `json:"MontoObjetivo"`
	MontoActual   float64   `json:"MontoActual"`
	FechaLimite   time.Time `json:"FechaLimite"`
	Icono         string    `json:"Icono"`
	Color         string    `json:"Color"`
	Activo        bool      `json:"Activo"`
}

type NivelRiesgo struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Inversion struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdNivelRiesgo *Ref      `json:"IdNivelRiesgo"`
	Nombre        string    `json:"Nombre"`
	Monto         float64   `json:"Monto"`
	Rendimiento   float64   `json:"Rendimiento"`
	Duracion      string    `json:"Duracion"`
	FechaInicio   time.Time `json:"FechaInicio"`
	FechaFin      time.Time `json:"FechaFin"`
	Activo        bool      `json:"Activo"`
}

type HistorialInversion struct {
	Id    int       `json:"Id"`
	Fecha time.Time `json:"Fecha"`
	Valor float64   `json:"Valor"`
}

type DatosImportar struct {
	IdUsuario   int `json:"id_usuario"`
	Movimientos []struct {
		Fecha         string  `json:"fecha"`
		Concepto      string  `json:"concepto"`
		Categoria     string  `json:"categoria"`
		Tipo          string  `json:"tipo"`
		Valor         float64 `json:"valor"`
		MetodoPago    string  `json:"metodo_pago"`
		Observaciones string  `json:"observaciones"`
	} `json:"movimientos"`
}

type DatosAbono struct {
	Valor float64 `json:"valor"`
}

// FinanzasController operations for Finanzas
type FinanzasController struct {
	beego.Controller
}

// ============================================================
// GET /v1/resumen-financiero/:id
// Ingresos, gastos, balance y gastos por categoria (crud_finanzas)
// ============================================================

func (c *FinanzasController) GetResumenFinanciero() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar los movimientos del usuario (mas nuevos primero)
	// --------------------------------------------------------

	responseMovimientos, err := http.Get(fmt.Sprintf("%s/movimiento_ingreso_egreso?query=id_usuario:%d,activo:true&sortby=Fecha&order=desc&limit=10000", urlCrudFinanzas, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseMovimientos.Body.Close()

	bodyMovimientos, _ := io.ReadAll(responseMovimientos.Body)

	var movimientos []Movimiento
	json.Unmarshal(bodyMovimientos, &movimientos)

	// --------------------------------------------------------
	// 3. Consultar las categorias (para saber nombre e icono)
	// --------------------------------------------------------

	responseCategorias, err := http.Get(urlCrudFinanzas + "/categoria?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las categorias"}
		c.ServeJSON()
		return
	}

	defer responseCategorias.Body.Close()

	bodyCategorias, _ := io.ReadAll(responseCategorias.Body)

	var categorias []Categoria
	json.Unmarshal(bodyCategorias, &categorias)

	categoriaPorId := map[int]Categoria{}
	for _, categoria := range categorias {
		categoriaPorId[categoria.Id] = categoria
	}

	// --------------------------------------------------------
	// 4. Sumar ingresos y gastos, y los gastos de cada categoria
	// --------------------------------------------------------

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

	// --------------------------------------------------------
	// 5. Armar la lista de gastos por categoria con su porcentaje
	// --------------------------------------------------------

	distribucion := []map[string]interface{}{}

	for idCategoria, total := range gastoPorCategoria {

		porcentaje := 0.0
		if totalGastos > 0 {
			porcentaje = math.Round(total / totalGastos * 100)
		}

		distribucion = append(distribucion, map[string]interface{}{
			"categoria":  categoriaPorId[idCategoria].Nombre,
			"icono":      categoriaPorId[idCategoria].Icono,
			"total":      total,
			"porcentaje": porcentaje,
		})
	}

	// --------------------------------------------------------
	// 6. Tomar los ultimos 5 movimientos
	// --------------------------------------------------------

	ultimos := []map[string]interface{}{}

	for i, movimiento := range movimientos {

		if i == 5 {
			break
		}

		tipo := "gasto"
		if movimiento.EsIngreso {
			tipo = "ingreso"
		}

		nombreCategoria := ""
		if movimiento.IdCategoria != nil {
			nombreCategoria = categoriaPorId[movimiento.IdCategoria.Id].Nombre
		}

		ultimos = append(ultimos, map[string]interface{}{
			"id":        movimiento.Id,
			"fecha":     movimiento.Fecha.Format("2006-01-02"),
			"concepto":  movimiento.Nombre,
			"categoria": nombreCategoria,
			"tipo":      tipo,
			"valor":     movimiento.Monto,
		})
	}

	// --------------------------------------------------------
	// 7. Devolver el resumen
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_usuario":          id,
		"total_ingresos":      totalIngresos,
		"total_gastos":        totalGastos,
		"balance":             totalIngresos - totalGastos,
		"gastos_categoria":    distribucion,
		"ultimos_movimientos": ultimos,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/movimientos/usuario/:id
// Libro mayor de un usuario con nombre e icono de la categoria.
// Filtros opcionales: ?tipo=ingreso|gasto  ?categoria=<id>
// ============================================================

func (c *FinanzasController) GetMovimientos() {

	// --------------------------------------------------------
	// 1. Obtener el ID y los filtros enviados en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	filtro := fmt.Sprintf("id_usuario:%d,activo:true", id)

	if c.GetString("tipo") == "ingreso" {
		filtro += ",es_ingreso:true"
	}

	if c.GetString("tipo") == "gasto" {
		filtro += ",es_ingreso:false"
	}

	if idCategoria, err := c.GetInt("categoria"); err == nil {
		filtro += fmt.Sprintf(",id_categoria:%d", idCategoria)
	}

	// --------------------------------------------------------
	// 2. Consultar los movimientos con esos filtros
	// --------------------------------------------------------

	responseMovimientos, err := http.Get(urlCrudFinanzas + "/movimiento_ingreso_egreso?query=" + filtro + "&sortby=Fecha&order=desc&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseMovimientos.Body.Close()

	bodyMovimientos, _ := io.ReadAll(responseMovimientos.Body)

	var movimientos []Movimiento
	json.Unmarshal(bodyMovimientos, &movimientos)

	// --------------------------------------------------------
	// 3. Consultar las categorias
	// --------------------------------------------------------

	responseCategorias, err := http.Get(urlCrudFinanzas + "/categoria?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las categorias"}
		c.ServeJSON()
		return
	}

	defer responseCategorias.Body.Close()

	bodyCategorias, _ := io.ReadAll(responseCategorias.Body)

	var categorias []Categoria
	json.Unmarshal(bodyCategorias, &categorias)

	categoriaPorId := map[int]Categoria{}
	for _, categoria := range categorias {
		categoriaPorId[categoria.Id] = categoria
	}

	// --------------------------------------------------------
	// 4. Construir la lista como la usa la pagina
	// --------------------------------------------------------

	lista := []map[string]interface{}{}

	for _, movimiento := range movimientos {

		tipo := "gasto"
		if movimiento.EsIngreso {
			tipo = "ingreso"
		}

		categoria := Categoria{}
		if movimiento.IdCategoria != nil {
			categoria = categoriaPorId[movimiento.IdCategoria.Id]
		}

		lista = append(lista, map[string]interface{}{
			"id":            movimiento.Id,
			"fecha":         movimiento.Fecha.Format("2006-01-02"),
			"concepto":      movimiento.Nombre,
			"id_categoria":  categoria.Id,
			"categoria":     categoria.Nombre,
			"icono":         categoria.Icono,
			"tipo":          tipo,
			"valor":         movimiento.Monto,
			"metodo_pago":   movimiento.MetodoPago,
			"observaciones": movimiento.Observaciones,
		})
	}

	// --------------------------------------------------------
	// 5. Devolver la lista
	// --------------------------------------------------------

	c.Data["json"] = lista
	c.ServeJSON()
}

// ============================================================
// POST /v1/movimientos/importar
// Guarda varios movimientos de una vez (plantilla de Excel).
// Devuelve cuantos se guardaron y que filas tuvieron error.
// ============================================================

func (c *FinanzasController) PostImportarMovimientos() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosImportar

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	if len(datos.Movimientos) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "No llego ningun movimiento"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar las categorias para buscarlas por nombre
	// --------------------------------------------------------

	responseCategorias, err := http.Get(urlCrudFinanzas + "/categoria?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseCategorias.Body.Close()

	bodyCategorias, _ := io.ReadAll(responseCategorias.Body)

	var categorias []Categoria
	json.Unmarshal(bodyCategorias, &categorias)

	categoriaPorNombre := map[string]int{}
	for _, categoria := range categorias {
		categoriaPorNombre[strings.ToLower(categoria.Nombre)] = categoria.Id
	}

	// --------------------------------------------------------
	// 3. Revisar y guardar cada fila
	// --------------------------------------------------------

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

		movimiento := Movimiento{
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
			movimiento.IdCategoria = &Ref{Id: idCategoria}
		}

		jsonMovimiento, _ := json.Marshal(movimiento)

		responseCrear, err := http.Post(urlCrudFinanzas+"/movimiento_ingreso_egreso", "application/json", bytes.NewBuffer(jsonMovimiento))

		if err != nil {
			errores = append(errores, fmt.Sprintf("Fila %d: no se pudo guardar", numeroFila))
			continue
		}

		responseCrear.Body.Close()

		if responseCrear.StatusCode != http.StatusCreated {
			errores = append(errores, fmt.Sprintf("Fila %d: no se pudo guardar", numeroFila))
			continue
		}

		guardados++
	}

	// --------------------------------------------------------
	// 4. Devolver el resultado
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"guardados": guardados,
		"errores":   errores,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/metas/usuario/:id
// Metas de un usuario con su porcentaje y si ya se cumplio
// ============================================================

func (c *FinanzasController) GetMetas() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar las metas del usuario
	// --------------------------------------------------------

	responseMetas, err := http.Get(fmt.Sprintf("%s/meta?query=id_usuario:%d,activo:true&limit=1000", urlCrudFinanzas, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseMetas.Body.Close()

	bodyMetas, _ := io.ReadAll(responseMetas.Body)

	var metas []Meta
	json.Unmarshal(bodyMetas, &metas)

	// --------------------------------------------------------
	// 3. Calcular el porcentaje de cada meta
	//    porcentaje = ahorrado / objetivo * 100 (maximo 100)
	// --------------------------------------------------------

	lista := []map[string]interface{}{}
	totalAhorrado := 0.0
	totalObjetivo := 0.0
	cumplidas := 0

	for _, meta := range metas {

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

	// --------------------------------------------------------
	// 4. Devolver las metas con los totales
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"metas":          lista,
		"total_ahorrado": totalAhorrado,
		"total_objetivo": totalObjetivo,
		"cumplidas":      cumplidas,
	}

	c.ServeJSON()
}

// ============================================================
// PUT /v1/metas/:id/abono
// Suma dinero a lo ahorrado en una meta
// ============================================================

func (c *FinanzasController) PutAbonoMeta() {

	// --------------------------------------------------------
	// 1. Obtener el ID y el valor a abonar
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	var datos DatosAbono

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil || datos.Valor <= 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El valor del abono debe ser mayor a 0"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar la meta completa
	//    Se guarda en un map para no perder ningun campo
	//    cuando se haga el PUT al CRUD.
	// --------------------------------------------------------

	responseMeta, err := http.Get(fmt.Sprintf("%s/meta/%d", urlCrudFinanzas, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseMeta.Body.Close()

	bodyMeta, _ := io.ReadAll(responseMeta.Body)

	var meta map[string]interface{}

	if err := json.Unmarshal(bodyMeta, &meta); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "La meta no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Sumar el abono y guardar
	// --------------------------------------------------------

	actual, _ := meta["MontoActual"].(float64)
	meta["MontoActual"] = actual + datos.Valor

	jsonMeta, _ := json.Marshal(meta)

	peticionPut, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/meta/%d", urlCrudFinanzas, id), bytes.NewBuffer(jsonMeta))
	peticionPut.Header.Set("Content-Type", "application/json")

	responsePut, err := http.DefaultClient.Do(peticionPut)

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo guardar el abono"}
		c.ServeJSON()
		return
	}

	responsePut.Body.Close()

	// --------------------------------------------------------
	// 4. Devolver el nuevo valor ahorrado
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id":       id,
		"actual":   meta["MontoActual"],
		"objetivo": meta["MontoObjetivo"],
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/inversiones/usuario/:id
// Inversiones con su riesgo, balance, rendimiento y la grafica
// ============================================================

func (c *FinanzasController) GetInversiones() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar las inversiones del usuario
	// --------------------------------------------------------

	responseInversiones, err := http.Get(fmt.Sprintf("%s/inversion?query=id_usuario:%d,activo:true&limit=1000", urlCrudFinanzas, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseInversiones.Body.Close()

	bodyInversiones, _ := io.ReadAll(responseInversiones.Body)

	var inversiones []Inversion
	json.Unmarshal(bodyInversiones, &inversiones)

	// --------------------------------------------------------
	// 3. Consultar los niveles de riesgo (Bajo, Medio, Alto)
	// --------------------------------------------------------

	responseNiveles, err := http.Get(urlCrudFinanzas + "/nivel_riesgo?limit=100")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los niveles de riesgo"}
		c.ServeJSON()
		return
	}

	defer responseNiveles.Body.Close()

	bodyNiveles, _ := io.ReadAll(responseNiveles.Body)

	var niveles []NivelRiesgo
	json.Unmarshal(bodyNiveles, &niveles)

	riesgoPorId := map[int]string{}
	for _, nivel := range niveles {
		riesgoPorId[nivel.Id] = nivel.Nombre
	}

	// --------------------------------------------------------
	// 4. Armar la lista y sumar balance y rendimiento
	// --------------------------------------------------------

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

		// Historial de esta inversion (puntos de la grafica)
		responseHistorial, err := http.Get(fmt.Sprintf("%s/historial_inversion?query=id_inversion:%d,activo:true&limit=1000", urlCrudFinanzas, inversion.Id))

		if err != nil {
			continue
		}

		bodyHistorial, _ := io.ReadAll(responseHistorial.Body)
		responseHistorial.Body.Close()

		var historial []HistorialInversion
		json.Unmarshal(bodyHistorial, &historial)

		for _, punto := range historial {
			valorPorFecha[punto.Fecha.Format("2006-01-02")] += punto.Valor
		}
	}

	// --------------------------------------------------------
	// 5. Ordenar los puntos de la grafica por fecha
	// --------------------------------------------------------

	fechas := []string{}
	for fecha := range valorPorFecha {
		fechas = append(fechas, fecha)
	}
	sort.Strings(fechas)

	grafica := []map[string]interface{}{}
	for _, fecha := range fechas {
		grafica = append(grafica, map[string]interface{}{"fecha": fecha, "valor": valorPorFecha[fecha]})
	}

	// --------------------------------------------------------
	// 6. Calcular el porcentaje de rendimiento y devolver todo
	//    porcentaje = rendimiento / balance * 100
	// --------------------------------------------------------

	porcentaje := 0.0
	if balance > 0 {
		porcentaje = math.Round(rendimientoTotal/balance*1000) / 10
	}

	c.Data["json"] = map[string]interface{}{
		"inversiones":            lista,
		"balance":                balance,
		"rendimiento_total":      rendimientoTotal,
		"rendimiento_porcentaje": porcentaje,
		"grafica":                grafica,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/admin/finanzas
// Para el administrador: movimientos, metas e inversiones de
// todos los usuarios, con el nombre de cada usuario
// (crud_auth + crud_finanzas)
// ============================================================

func (c *FinanzasController) GetAdminFinanzas() {

	// --------------------------------------------------------
	// 1. Consultar los usuarios (para poner su nombre)
	// --------------------------------------------------------

	responseUsuarios, err := http.Get(urlCrudAuth + "/usuario?limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuarios.Body.Close()

	bodyUsuarios, _ := io.ReadAll(responseUsuarios.Body)

	var usuarios []Usuario
	json.Unmarshal(bodyUsuarios, &usuarios)

	nombrePorId := map[int]string{}
	for _, usuario := range usuarios {
		nombrePorId[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	// --------------------------------------------------------
	// 2. Consultar los movimientos de todos
	// --------------------------------------------------------

	responseMovimientos, err := http.Get(urlCrudFinanzas + "/movimiento_ingreso_egreso?query=activo:true&sortby=Fecha&order=desc&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de finanzas"}
		c.ServeJSON()
		return
	}

	defer responseMovimientos.Body.Close()

	bodyMovimientos, _ := io.ReadAll(responseMovimientos.Body)

	var movimientos []Movimiento
	json.Unmarshal(bodyMovimientos, &movimientos)

	// --------------------------------------------------------
	// 3. Consultar las metas de todos
	// --------------------------------------------------------

	responseMetas, err := http.Get(urlCrudFinanzas + "/meta?query=activo:true&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las metas"}
		c.ServeJSON()
		return
	}

	defer responseMetas.Body.Close()

	bodyMetas, _ := io.ReadAll(responseMetas.Body)

	var metas []Meta
	json.Unmarshal(bodyMetas, &metas)

	// --------------------------------------------------------
	// 4. Consultar las inversiones de todos
	// --------------------------------------------------------

	responseInversiones, err := http.Get(urlCrudFinanzas + "/inversion?query=activo:true&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las inversiones"}
		c.ServeJSON()
		return
	}

	defer responseInversiones.Body.Close()

	bodyInversiones, _ := io.ReadAll(responseInversiones.Body)

	var inversiones []Inversion
	json.Unmarshal(bodyInversiones, &inversiones)

	// --------------------------------------------------------
	// 5. Armar las tres listas con el nombre del usuario
	// --------------------------------------------------------

	listaMovimientos := []map[string]interface{}{}
	for _, movimiento := range movimientos {
		tipo := "gasto"
		if movimiento.EsIngreso {
			tipo = "ingreso"
		}
		listaMovimientos = append(listaMovimientos, map[string]interface{}{
			"id":       movimiento.Id,
			"usuario":  nombrePorId[movimiento.IdUsuario],
			"fecha":    movimiento.Fecha.Format("2006-01-02"),
			"concepto": movimiento.Nombre,
			"tipo":     tipo,
			"valor":    movimiento.Monto,
		})
	}

	listaMetas := []map[string]interface{}{}
	for _, meta := range metas {
		listaMetas = append(listaMetas, map[string]interface{}{
			"id":       meta.Id,
			"usuario":  nombrePorId[meta.IdUsuario],
			"nombre":   meta.Nombre,
			"actual":   meta.MontoActual,
			"objetivo": meta.MontoObjetivo,
		})
	}

	listaInversiones := []map[string]interface{}{}
	for _, inversion := range inversiones {
		listaInversiones = append(listaInversiones, map[string]interface{}{
			"id":          inversion.Id,
			"usuario":     nombrePorId[inversion.IdUsuario],
			"nombre":      inversion.Nombre,
			"monto":       inversion.Monto,
			"rendimiento": inversion.Rendimiento,
		})
	}

	// --------------------------------------------------------
	// 6. Devolver todo
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"movimientos": listaMovimientos,
		"metas":       listaMetas,
		"inversiones": listaInversiones,
	}

	c.ServeJSON()
}
