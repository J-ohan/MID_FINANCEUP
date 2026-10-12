package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
)

type Banco struct {
	Id          int    `json:"Id"`
	NombreBanco string `json:"NombreBanco"`
	UrlLogo     string `json:"UrlLogo"`
}

type ProductoCrediticio struct {
	Id             int         `json:"Id"`
	IdBanco        *Referencia `json:"IdBanco"`
	NombreProducto string      `json:"NombreProducto"`
	Descripcion    string      `json:"Descripcion"`
	MontoMinimo    float64     `json:"MontoMinimo"`
	MontoMaximo    float64     `json:"MontoMaximo"`
	TasaMinima     float64     `json:"TasaMinima"`
	TasaMaxima     float64     `json:"TasaMaxima"`
	PlazoMinimo    int         `json:"PlazoMinimo"`
	PlazoMaximo    int         `json:"PlazoMaximo"`
}

// NegocioController operations for Negocio (alianzas)
type NegocioController struct {
	beego.Controller
}

// ============================================================
// GET /v1/ofertas
// Productos de credito con el nombre de su banco (crud_negocio)
// ============================================================

func (c *NegocioController) GetOfertas() {

	// --------------------------------------------------------
	// 1. Consultar los productos de credito activos
	// --------------------------------------------------------

	responseProductos, err := http.Get("http://localhost:8084/v1/producto_crediticio?query=activo:true&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de negocio"}
		c.ServeJSON()
		return
	}

	defer responseProductos.Body.Close()

	bodyProductos, _ := io.ReadAll(responseProductos.Body)

	var productos []ProductoCrediticio
	json.Unmarshal(bodyProductos, &productos)

	// --------------------------------------------------------
	// 2. Consultar los bancos
	// --------------------------------------------------------

	responseBancos, err := http.Get("http://localhost:8084/v1/banco?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los bancos"}
		c.ServeJSON()
		return
	}

	defer responseBancos.Body.Close()

	bodyBancos, _ := io.ReadAll(responseBancos.Body)

	var bancos []Banco
	json.Unmarshal(bodyBancos, &bancos)

	bancoPorId := map[int]Banco{}
	for _, banco := range bancos {
		bancoPorId[banco.Id] = banco
	}

	// --------------------------------------------------------
	// 3. Unir cada producto con su banco
	// --------------------------------------------------------

	ofertas := []map[string]interface{}{}

	for _, producto := range productos {

		banco := bancoPorId[producto.IdBanco.Id]

		ofertas = append(ofertas, map[string]interface{}{
			"id_producto":  producto.Id,
			"producto":     producto.NombreProducto,
			"descripcion":  producto.Descripcion,
			"banco":        banco.NombreBanco,
			"monto_minimo": producto.MontoMinimo,
			"monto_maximo": producto.MontoMaximo,
			"tasa_minima":  producto.TasaMinima,
			"plazo_minimo": producto.PlazoMinimo,
			"plazo_maximo": producto.PlazoMaximo,
		})
	}

	// --------------------------------------------------------
	// 4. Devolver las ofertas
	// --------------------------------------------------------

	c.Data["json"] = ofertas
	c.ServeJSON()
}

// Pendiente:
// POST /v1/solicitud-credito            -> crear la solicitud (lead)
// GET  /v1/solicitud-credito/usuario/:id -> solicitudes de un usuario
