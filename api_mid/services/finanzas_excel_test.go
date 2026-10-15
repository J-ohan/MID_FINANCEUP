package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// crudDePrueba simula crud_auth y crud_finanzas en memoria.
// fallarEnPost: numero del POST de movimientos que falla (0 = ninguno).
type crudDePrueba struct {
	movimientos  map[float64]map[string]interface{}
	posts        int
	fallarEnPost int
}

func (c *crudDePrueba) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	responder := func(codigo int, datos interface{}) {
		w.WriteHeader(codigo)
		json.NewEncoder(w).Encode(datos)
	}

	switch {
	case strings.HasPrefix(r.URL.Path, "/auth/usuario/1"):
		responder(200, map[string]interface{}{"Id": 1, "Nombre": "Sharith"})

	case strings.HasPrefix(r.URL.Path, "/auth/usuario/"):
		responder(200, "<QuerySeter> no row found")

	case r.URL.Path == "/fin/categoria":
		responder(200, []map[string]interface{}{
			{"Id": 1, "Nombre": "Otros ingresos"},
			{"Id": 2, "Nombre": "Otros"},
			{"Id": 3, "Nombre": "Gastos hormiga"},
		})

	case r.Method == http.MethodGet && r.URL.Path == "/fin/movimiento_ingreso_egreso":
		lista := []map[string]interface{}{}
		for _, movimiento := range c.movimientos {
			if movimiento["Activo"] == true {
				lista = append(lista, movimiento)
			}
		}
		responder(200, lista)

	case r.Method == http.MethodPost:
		c.posts++
		if c.posts == c.fallarEnPost {
			responder(400, "error de prueba")
			return
		}
		var movimiento map[string]interface{}
		json.NewDecoder(r.Body).Decode(&movimiento)
		movimiento["Id"] = float64(len(c.movimientos) + 1)
		c.movimientos[movimiento["Id"].(float64)] = movimiento
		responder(201, movimiento)

	case r.Method == http.MethodPut:
		var movimiento map[string]interface{}
		json.NewDecoder(r.Body).Decode(&movimiento)
		c.movimientos[movimiento["Id"].(float64)] = movimiento
		responder(200, "OK")
	}
}

func (c *crudDePrueba) activos() int {
	total := 0
	for _, movimiento := range c.movimientos {
		if movimiento["Activo"] == true {
			total++
		}
	}
	return total
}

func usarCrudDePrueba(t *testing.T, fallarEnPost int) *crudDePrueba {
	crud := &crudDePrueba{movimientos: map[float64]map[string]interface{}{}, fallarEnPost: fallarEnPost}
	servidor := httptest.NewServer(crud)
	t.Cleanup(servidor.Close)

	helpers.UrlCrudAuth = servidor.URL + "/auth"
	helpers.UrlCrudFinanzas = servidor.URL + "/fin"

	return crud
}

var filasExcel = []models.FilaExcel{
	{Concepto: "Salario", Categoria: "Otros ingresos", Tipo: "ingreso", Valor: 1000},
	{Concepto: "", Categoria: "Otros", Tipo: "gasto", Valor: 50},
	{Concepto: "Arriendo", Categoria: "Otros", Tipo: "gasto", Valor: 400},
	{Concepto: "Cafe", Categoria: "Gastos hormiga", Tipo: "gasto", Valor: 30},
}

func TestImportarGuardaLasFilasBuenas(t *testing.T) {

	crud := usarCrudDePrueba(t, 0)

	resultado, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: filasExcel})

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	if resultado["guardados"] != 3 || len(resultado["errores"].([]string)) != 1 || crud.activos() != 3 {
		t.Errorf("esperaba 3 guardados y 1 error, llego %v", resultado)
	}
}

func TestImportarDeshaceSiElCrudFalla(t *testing.T) {

	// El segundo POST falla: el primero que ya se guardo se debe desactivar
	crud := usarCrudDePrueba(t, 2)

	_, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: filasExcel})

	if err == nil || crud.activos() != 0 {
		t.Errorf("esperaba error y 0 movimientos activos, quedaron %d (%v)", crud.activos(), err)
	}
}

func TestImportarUsuarioInexistente(t *testing.T) {

	usarCrudDePrueba(t, 0)

	_, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 9, Movimientos: filasExcel})

	if errorApi, ok := err.(*helpers.ErrorApi); !ok || errorApi.Codigo != http.StatusNotFound {
		t.Errorf("esperaba 404, llego %v", err)
	}
}

func TestExportarArmaLaPlantilla(t *testing.T) {

	usarCrudDePrueba(t, 0)

	if _, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: filasExcel}); err != nil {
		t.Fatalf("no esperaba error al importar: %v", err)
	}

	resultado, err := ExportarMovimientos(1, models.FiltrosMovimientos{})

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	esperado := map[string]float64{
		"total_ingresos":       1000,
		"total_gastos_fijos":   400,
		"total_gastos_hormiga": 30,
		"proyeccion_ahorro":    100,
		"balance_mes":          570,
		"ahorro_mensual":       470,
	}

	for campo, valor := range esperado {
		if resultado[campo] != valor {
			t.Errorf("%s = %v, esperaba %v", campo, resultado[campo], valor)
		}
	}
}
