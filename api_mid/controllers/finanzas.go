package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type Categoria struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Movimiento struct {
	Id          int       `json:"Id"`
	IdCategoria *Ref      `json:"IdCategoria"`
	Nombre      string    `json:"Nombre"`
	Monto       float64   `json:"Monto"`
	EsIngreso   bool      `json:"EsIngreso"`
	Fecha       time.Time `json:"Fecha"`
}

type Meta struct {
	Id            int     `json:"Id"`
	Nombre        string  `json:"Nombre"`
	MontoObjetivo float64 `json:"MontoObjetivo"`
	MontoActual   float64 `json:"MontoActual"`
	Icono         string  `json:"Icono"`
}

type Inversion struct {
	Id           int     `json:"Id"`
	Nombre       string  `json:"Nombre"`
	Monto        float64 `json:"Monto"`
	Rentabilidad float64 `json:"Rentabilidad"`
}

type SolicitudConsolidacion struct {
	Id             int     `json:"Id"`
	SaldoTotal     float64 `json:"SaldoTotal"`
	CuotaActual    float64 `json:"CuotaActual"`
	CuotaPropuesta float64 `json:"CuotaPropuesta"`
	Estado         string  `json:"Estado"`
}

// FinanzasController operations for Finanzas
type FinanzasController struct {
	beego.Controller
}

// ============================================================
// GET /v1/resumen-financiero/:id
// Ingresos, gastos, balance, metas, inversiones y deudas
// (crud_auth + crud_finanzas)
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
	// 2. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario

	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar las categorias (para saber su nombre)
	// --------------------------------------------------------

	responseCategorias, err := http.Get("http://localhost:8082/v1/categoria?limit=1000")

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

	nombreCategoria := map[int]string{}
	for _, categoria := range categorias {
		nombreCategoria[categoria.Id] = categoria.Nombre
	}

	// --------------------------------------------------------
	// 4. Consultar los movimientos y sumar ingresos y gastos
	// --------------------------------------------------------

	responseMovimientos, err := http.Get(fmt.Sprintf("http://localhost:8082/v1/movimiento_ingreso_egreso?query=id_usuario:%d,activo:true&limit=1000", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los movimientos"}
		c.ServeJSON()
		return
	}

	defer responseMovimientos.Body.Close()

	bodyMovimientos, _ := io.ReadAll(responseMovimientos.Body)

	var movimientos []Movimiento
	json.Unmarshal(bodyMovimientos, &movimientos)

	totalIngresos := 0.0
	totalGastos := 0.0
	gastosPorCategoria := map[string]float64{}

	for _, movimiento := range movimientos {

		if movimiento.EsIngreso {
			totalIngresos += movimiento.Monto
			continue
		}

		totalGastos += movimiento.Monto

		categoria := "Sin categoria"
		if movimiento.IdCategoria != nil {
			categoria = nombreCategoria[movimiento.IdCategoria.Id]
		}
		gastosPorCategoria[categoria] += movimiento.Monto
	}

	// --------------------------------------------------------
	// 5. Consultar las metas y calcular su porcentaje
	// --------------------------------------------------------

	responseMetas, err := http.Get(fmt.Sprintf("http://localhost:8082/v1/meta?query=id_usuario:%d,activo:true&limit=1000", id))

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

	listaMetas := []map[string]interface{}{}

	for _, meta := range metas {

		porcentaje := 0.0
		if meta.MontoObjetivo > 0 {
			porcentaje = meta.MontoActual * 100 / meta.MontoObjetivo
		}
		if porcentaje > 100 {
			porcentaje = 100
		}

		listaMetas = append(listaMetas, map[string]interface{}{
			"id":         meta.Id,
			"nombre":     meta.Nombre,
			"icono":      meta.Icono,
			"actual":     meta.MontoActual,
			"objetivo":   meta.MontoObjetivo,
			"porcentaje": porcentaje,
			"cumplida":   meta.MontoActual >= meta.MontoObjetivo,
		})
	}

	// --------------------------------------------------------
	// 6. Consultar las inversiones y su ganancia estimada
	// --------------------------------------------------------

	responseInversiones, err := http.Get(fmt.Sprintf("http://localhost:8082/v1/inversion?query=id_usuario:%d,activo:true&limit=1000", id))

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

	totalInvertido := 0.0
	gananciaEstimada := 0.0
	listaInversiones := []map[string]interface{}{}

	for _, inversion := range inversiones {

		totalInvertido += inversion.Monto
		gananciaEstimada += inversion.Monto * inversion.Rentabilidad / 100

		listaInversiones = append(listaInversiones, map[string]interface{}{
			"id":           inversion.Id,
			"nombre":       inversion.Nombre,
			"monto":        inversion.Monto,
			"rentabilidad": inversion.Rentabilidad,
		})
	}

	// --------------------------------------------------------
	// 7. Consultar las solicitudes de "Resuelve tu deuda"
	// --------------------------------------------------------

	responseDeudas, err := http.Get(fmt.Sprintf("http://localhost:8082/v1/solicitud_consolidacion?query=id_usuario:%d,activo:true&limit=1000", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las solicitudes de deuda"}
		c.ServeJSON()
		return
	}

	defer responseDeudas.Body.Close()

	bodyDeudas, _ := io.ReadAll(responseDeudas.Body)

	var deudas []SolicitudConsolidacion
	json.Unmarshal(bodyDeudas, &deudas)

	listaDeudas := []map[string]interface{}{}

	for _, deuda := range deudas {
		listaDeudas = append(listaDeudas, map[string]interface{}{
			"id":              deuda.Id,
			"saldo_total":     deuda.SaldoTotal,
			"cuota_actual":    deuda.CuotaActual,
			"cuota_propuesta": deuda.CuotaPropuesta,
			"ahorro_mensual":  deuda.CuotaActual - deuda.CuotaPropuesta,
			"estado":          deuda.Estado,
		})
	}

	// --------------------------------------------------------
	// 8. Construir el resumen y devolverlo
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_usuario":           usuario.Id,
		"total_ingresos":       totalIngresos,
		"total_gastos":         totalGastos,
		"balance":              totalIngresos - totalGastos,
		"gastos_por_categoria": gastosPorCategoria,
		"metas":                listaMetas,
		"total_invertido":      totalInvertido,
		"ganancia_estimada":    gananciaEstimada,
		"inversiones":          listaInversiones,
		"solicitudes_deuda":    listaDeudas,
	}

	c.ServeJSON()
}
