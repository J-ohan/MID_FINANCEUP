package finanzas

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// PRUEBAS DEL EXCEL
// Se corren con: go test ./services/finanzas
//
// No necesitan los CRUD encendidos: crudDePrueba hace de cuenta que
// es crud_auth y crud_finanzas, guardando todo en memoria.
// ============================================================

// crudDePrueba simula los CRUD. fallarEnPost dice que numero de POST
// debe fallar (0 = ninguno), para probar que la importacion se deshace.
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
	// Solo existe el usuario 1; cualquier otro responde como bee cuando no encuentra
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

// activos cuenta cuantos movimientos quedaron sin borrar
func (c *crudDePrueba) activos() int {
	total := 0
	for _, movimiento := range c.movimientos {
		if movimiento["Activo"] == true {
			total++
		}
	}
	return total
}

// usarCrudDePrueba enciende el CRUD de prueba y hace que el MID le hable a el
func usarCrudDePrueba(t *testing.T, fallarEnPost int) *crudDePrueba {
	crud := &crudDePrueba{movimientos: map[float64]map[string]interface{}{}, fallarEnPost: fallarEnPost}
	servidor := httptest.NewServer(crud)
	t.Cleanup(servidor.Close)

	helpers.UrlCrudAuth = servidor.URL + "/auth"
	helpers.UrlCrudFinanzas = servidor.URL + "/fin"

	return crud
}

// Cuatro filas: tres buenas y una sin concepto (la fila 2)
var filasExcel = []models.FilaExcel{
	{Fecha: "2026-10-01", Concepto: "Salario", Categoria: "Otros ingresos", Tipo: "ingreso", Valor: 1000},
	{Fecha: "2026-10-01", Concepto: "", Categoria: "Otros", Tipo: "gasto", Valor: 50},
	{Fecha: "2026-10-02", Concepto: "Arriendo", Categoria: "Otros", Tipo: "gasto", Valor: 400},
	{Fecha: "2026-10-03", Concepto: "Cafe", Categoria: "Gastos hormiga", Tipo: "gasto", Valor: 30},
}

func TestImportarEjemploDeLaGuia(t *testing.T) {

	// El mismo body de la guia (Paso 6, Postman): sin categoria
	crud := usarCrudDePrueba(t, 0)

	resultado, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: []models.FilaExcel{
		{Fecha: "2026-10-01", Concepto: "Sueldo", Tipo: "ingreso", Valor: 3000000},
	}})

	if err != nil || resultado["guardados"] != 1 || crud.activos() != 1 {
		t.Errorf("la guia espera guardados: 1, llego %v (%v)", resultado, err)
	}
}

func TestImportarLeeElMetodoDePagoDelCliente(t *testing.T) {

	// Una fila escrita igual que la arma el cliente (con "metodoPago")
	crud := usarCrudDePrueba(t, 0)

	var datos models.DatosImportar
	json.Unmarshal([]byte(`{"id_usuario": 1, "movimientos": [
		{"fecha": "2026-10-01", "concepto": "Mercado", "categoria": "Otros", "tipo": "gasto", "valor": 150000, "metodoPago": "Efectivo"}]}`), &datos)

	if _, err := ImportarMovimientos(datos); err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	if guardado := crud.movimientos[1]; guardado["MetodoPago"] != "Efectivo" {
		t.Errorf("el metodo de pago no llego al CRUD: %v", guardado)
	}
}

func TestImportarGuardaLasFilasBuenas(t *testing.T) {

	crud := usarCrudDePrueba(t, 0)

	resultado, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: filasExcel})

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	errores := resultado["errores"].([]models.ErrorFila)

	if resultado["guardados"] != 3 || len(errores) != 1 || crud.activos() != 3 {
		t.Fatalf("esperaba 3 guardados y 1 error, llego %v", resultado)
	}

	// La fila 2 del Excel es la que no tiene concepto
	if errores[0].Fila != 2 || errores[0].Motivo == "" {
		t.Errorf("error mal armado: %+v", errores[0])
	}
}

func TestImportarModoValidarNoGuarda(t *testing.T) {

	crud := usarCrudDePrueba(t, 0)

	resultado, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Modo: "validar", Movimientos: filasExcel})

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	if resultado["validos"] != 3 || resultado["guardados"] != 0 || crud.activos() != 0 {
		t.Errorf("validar no debe guardar nada: %v (activos %d)", resultado, crud.activos())
	}
}

func TestImportarModoInvalido(t *testing.T) {

	usarCrudDePrueba(t, 0)

	_, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Modo: "borrar", Movimientos: filasExcel})

	if errorApi, ok := err.(*helpers.ErrorApi); !ok || errorApi.Codigo != http.StatusBadRequest {
		t.Errorf("esperaba 400, llego %v", err)
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
