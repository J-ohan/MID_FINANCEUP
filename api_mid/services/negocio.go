package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE ALIANZAS (crud_negocio)
// Pendiente: cambiar las llamadas por los helpers cuando esten.
// ============================================================

// Ofertas: productos de credito con el nombre de su banco.
// Devuelve el resultado, el codigo HTTP y el error (si hubo).
func Ofertas() ([]map[string]interface{}, int, error) {

	// 1. Consultar los productos de credito activos
	responseProductos, err := http.Get("http://localhost:8084/v1/producto_crediticio?query=activo:true&limit=1000")

	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("No fue posible comunicarse con la API de negocio")
	}

	defer responseProductos.Body.Close()

	bodyProductos, _ := io.ReadAll(responseProductos.Body)

	var productos []models.ProductoCrediticio
	json.Unmarshal(bodyProductos, &productos)

	// 2. Consultar los bancos y guardarlos por su id
	responseBancos, err := http.Get("http://localhost:8084/v1/banco?limit=1000")

	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("No fue posible consultar los bancos")
	}

	defer responseBancos.Body.Close()

	bodyBancos, _ := io.ReadAll(responseBancos.Body)

	var bancos []models.Banco
	json.Unmarshal(bodyBancos, &bancos)

	bancoPorId := map[int]models.Banco{}
	for _, banco := range bancos {
		bancoPorId[banco.Id] = banco
	}

	// 3. Unir cada producto con su banco
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

	return ofertas, http.StatusOK, nil
}

// Pendiente: FichaOferta, RadicarSolicitud y SolicitudesUsuario
