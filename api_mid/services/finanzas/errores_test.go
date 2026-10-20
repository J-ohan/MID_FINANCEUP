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
// PRUEBAS DEL MANEJO DE ERRORES Y DE LA EXPORTACION DE INVERSIONES
// Se corren con: go test ./services/finanzas
//
// crudFinanzasDePrueba hace de cuenta que es crud_auth y crud_finanzas.
// Solo existe el usuario 1. La meta 10 esta activa y la 11 borrada.
// ============================================================

type crudFinanzasDePrueba struct {
	metas          map[string]map[string]interface{}
	fallaHistorial bool // true = el historial "se cae" (cierra la conexion)
}

func (c *crudFinanzasDePrueba) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	responder := func(datos interface{}) { json.NewEncoder(w).Encode(datos) }
	ruta := r.URL.Path

	switch {
	case ruta == "/auth/usuario/1":
		responder(map[string]interface{}{"Id": 1, "Nombre": "Sharith"})

	case strings.HasPrefix(ruta, "/auth/usuario/"):
		responder("<QuerySeter> no row found")

	case ruta == "/fin/categoria":
		responder([]map[string]interface{}{{"Id": 1, "Nombre": "Vivienda"}})

	case ruta == "/fin/nivel_riesgo":
		responder([]map[string]interface{}{{"Id": 1, "Nombre": "Bajo"}, {"Id": 2, "Nombre": "Medio"}, {"Id": 3, "Nombre": "Alto"}})

	case ruta == "/fin/inversion":
		responder([]map[string]interface{}{
			{"Id": 1, "Nombre": "Acciones", "Monto": 1000, "Rendimiento": 100, "IdNivelRiesgo": map[string]interface{}{"Id": 3}, "Duracion": "1 año", "Activo": true},
			{"Id": 2, "Nombre": "CDT", "Monto": 500, "Rendimiento": -50, "IdNivelRiesgo": map[string]interface{}{"Id": 1}, "Duracion": "6 meses", "Activo": true},
			{"Id": 3, "Nombre": "Bonos", "Monto": 500, "Rendimiento": 0, "Duracion": "Flexible", "Activo": true},
		})

	case ruta == "/fin/historial_inversion":
		if c.fallaHistorial {
			// Cerrar la conexion de golpe, como si el CRUD se cayera
			conexion, _, _ := w.(http.Hijacker).Hijack()
			conexion.Close()
			return
		}
		if r.URL.Query().Get("query") == "id_inversion:1,activo:true" {
			responder([]map[string]interface{}{{"Id": 1, "Fecha": "2026-10-01T00:00:00Z", "Valor": 1050}})
			return
		}
		responder([]map[string]interface{}{})

	case strings.HasPrefix(ruta, "/fin/meta/"):
		meta, existe := c.metas[strings.TrimPrefix(ruta, "/fin/meta/")]
		if !existe {
			responder("<QuerySeter> no row found")
			return
		}
		if r.Method == http.MethodPut {
			json.NewDecoder(r.Body).Decode(&meta)
			responder("OK")
			return
		}
		responder(meta)

	default: // listas vacias (movimientos, metas)
		responder([]map[string]interface{}{})
	}
}

func usarCrudFinanzas(t *testing.T) *crudFinanzasDePrueba {

	crud := &crudFinanzasDePrueba{metas: map[string]map[string]interface{}{
		"10": {"Id": 10, "Nombre": "Viaje", "MontoActual": 100, "MontoObjetivo": 1000, "Activo": true},
		"11": {"Id": 11, "Nombre": "Borrada", "MontoActual": 0, "MontoObjetivo": 500, "Activo": false},
	}}

	servidor := httptest.NewServer(crud)
	t.Cleanup(servidor.Close)

	helpers.UrlCrudAuth = servidor.URL + "/auth"
	helpers.UrlCrudFinanzas = servidor.URL + "/fin"

	return crud
}

// codigoDe saca el codigo HTTP de un error (0 si no hubo error)
func codigoDe(err error) int {
	if errorApi, ok := err.(*helpers.ErrorApi); ok {
		return errorApi.Codigo
	}
	return 0
}

// Las funciones que reciben un id de usuario
var funcionesConUsuario = map[string]func(id int) error{
	"ResumenFinanciero":   func(id int) error { _, err := ResumenFinanciero(id); return err },
	"LibroMayor":          func(id int) error { _, err := LibroMayor(id, models.FiltrosMovimientos{}); return err },
	"ExportarMovimientos": func(id int) error { _, err := ExportarMovimientos(id, models.FiltrosMovimientos{}); return err },
	"MetasUsuario":        func(id int) error { _, err := MetasUsuario(id); return err },
	"InversionesUsuario":  func(id int) error { _, err := InversionesUsuario(id); return err },
	"ExportarInversiones": func(id int) error { _, err := ExportarInversiones(id); return err },
}

func TestErroresIdDeUsuarioInvalido(t *testing.T) {

	usarCrudFinanzas(t)

	for nombre, funcion := range funcionesConUsuario {
		for _, id := range []int{0, -5} {
			if codigo := codigoDe(funcion(id)); codigo != http.StatusBadRequest {
				t.Errorf("%s(%d): esperaba 400, llego %d", nombre, id, codigo)
			}
		}
	}
}

func TestErroresUsuarioQueNoExiste(t *testing.T) {

	usarCrudFinanzas(t)

	for nombre, funcion := range funcionesConUsuario {
		if codigo := codigoDe(funcion(99)); codigo != http.StatusNotFound {
			t.Errorf("%s(99): esperaba 404, llego %d", nombre, codigo)
		}
		if err := funcion(1); err != nil {
			t.Errorf("%s(1): el usuario 1 existe, no esperaba error: %v", nombre, err)
		}
	}
}

func TestErroresCrudApagado(t *testing.T) {

	// Un servidor que se apaga enseguida: el CRUD "no responde"
	apagado := httptest.NewServer(http.NotFoundHandler())
	apagado.Close()

	helpers.UrlCrudAuth = apagado.URL
	helpers.UrlCrudFinanzas = apagado.URL

	for nombre, funcion := range funcionesConUsuario {
		if codigo := codigoDe(funcion(1)); codigo != http.StatusBadGateway {
			t.Errorf("%s con el CRUD apagado: esperaba 502, llego %d", nombre, codigo)
		}
	}
}

func TestErroresFiltrosDelLibroMayor(t *testing.T) {

	usarCrudFinanzas(t)

	casos := map[string]models.FiltrosMovimientos{
		"tipo que no existe":      {Tipo: "otro"},
		"fecha mal escrita":       {Desde: "01/10/2026"},
		"desde despues de hasta":  {Desde: "2026-10-31", Hasta: "2026-10-01"},
		"categoria que no existe": {Categoria: "NoExiste"},
	}

	for nombre, filtros := range casos {
		if _, err := LibroMayor(1, filtros); codigoDe(err) != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, llego %v", nombre, err)
		}
	}

	// Filtros correctos (incluido "todas" y el mismo dia en desde y hasta)
	if _, err := LibroMayor(1, models.FiltrosMovimientos{Tipo: "todos", Categoria: "todas", Desde: "2026-10-01", Hasta: "2026-10-01"}); err != nil {
		t.Errorf("filtros correctos no debian dar error: %v", err)
	}
}

func TestErroresAbonarMeta(t *testing.T) {

	crud := usarCrudFinanzas(t)

	casos := []struct {
		nombre   string
		idMeta   int
		valor    float64
		esperado int
	}{
		{"id de meta en 0", 0, 50, http.StatusBadRequest},
		{"abono en 0", 10, 0, http.StatusBadRequest},
		{"abono negativo", 10, -20, http.StatusBadRequest},
		{"meta que no existe", 99, 50, http.StatusNotFound},
		{"meta borrada", 11, 50, http.StatusNotFound},
	}

	for _, caso := range casos {
		if _, err := AbonarMeta(caso.idMeta, caso.valor); codigoDe(err) != caso.esperado {
			t.Errorf("%s: esperaba %d, llego %v", caso.nombre, caso.esperado, err)
		}
	}

	// La meta borrada no debe haber cambiado
	if crud.metas["11"]["MontoActual"] != 0 && crud.metas["11"]["MontoActual"] != float64(0) {
		t.Errorf("la meta borrada recibio un abono: %v", crud.metas["11"])
	}

	// Un abono correcto suma
	resultado, err := AbonarMeta(10, 50)

	if err != nil || resultado["actual"] != 150.0 {
		t.Errorf("el abono correcto debia dejar 150, llego %v %v", resultado, err)
	}
}

func TestErroresImportarDemasiadasFilas(t *testing.T) {

	usarCrudFinanzas(t)

	filas := make([]models.FilaExcel, LimiteFilasExcel+1)

	if _, err := ImportarMovimientos(models.DatosImportar{IdUsuario: 1, Movimientos: filas}); codigoDe(err) != http.StatusBadRequest {
		t.Errorf("mas de %d filas debia dar 400, llego %v", LimiteFilasExcel, err)
	}
}

func TestInversionesAvisaSiFallaElHistorial(t *testing.T) {

	crud := usarCrudFinanzas(t)
	crud.fallaHistorial = true

	resultado, err := InversionesUsuario(1)

	if err != nil {
		t.Fatalf("si solo falla el historial, la pantalla debe responder igual: %v", err)
	}

	if avisos, ok := resultado["avisos"].([]string); !ok || len(avisos) != 1 {
		t.Errorf("debia traer un aviso de la grafica: %v", resultado["avisos"])
	}

	if len(resultado["inversiones"].([]map[string]interface{})) != 3 {
		t.Errorf("las inversiones debian salir aunque falle el historial")
	}
}

func TestExportarInversiones(t *testing.T) {

	usarCrudFinanzas(t)

	resultado, err := ExportarInversiones(1)

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	// Resumen: 1000 + 500 + 500 invertidos, 100 - 50 + 0 de rendimiento
	resumen := resultado["resumen"].(map[string]interface{})
	esperado := map[string]interface{}{"total_invertido": 2000.0, "rendimiento_total": 50.0, "valor_actual": 2050.0, "rentabilidad": 2.5, "numero_inversiones": 3}

	for campo, valor := range esperado {
		if resumen[campo] != valor {
			t.Errorf("resumen.%s = %v, esperaba %v", campo, resumen[campo], valor)
		}
	}

	// Detalle ordenado de mayor a menor monto, con su rentabilidad y participacion
	detalle := resultado["inversiones"].([]map[string]interface{})

	if detalle[0]["nombre"] != "Acciones" || detalle[0]["rentabilidad"] != 10.0 || detalle[0]["valor_actual"] != 1100.0 || detalle[0]["participacion"] != 50.0 {
		t.Errorf("la primera fila esta mal: %v", detalle[0])
	}

	if detalle[1]["rentabilidad"] != -10.0 {
		t.Errorf("una perdida debe salir con rentabilidad negativa: %v", detalle[1])
	}

	// Reparto por riesgo: Bajo 25 %, Medio 0 %, Alto 50 % y la que no tiene riesgo
	porRiesgo := resultado["por_riesgo"].([]map[string]interface{})
	porcentajes := map[string]interface{}{}
	for _, fila := range porRiesgo {
		porcentajes[fila["riesgo"].(string)] = fila["porcentaje"]
	}

	if porcentajes["Bajo"] != 25.0 || porcentajes["Medio"] != 0.0 || porcentajes["Alto"] != 50.0 || porcentajes["Sin definir"] != 25.0 {
		t.Errorf("reparto por riesgo mal: %v", porcentajes)
	}

	// 50 % en riesgo Alto: debe salir la alerta
	if !strings.Contains(resultado["alerta"].(string), "50%") {
		t.Errorf("debia salir la alerta de riesgo alto: %q", resultado["alerta"])
	}

	// Crecimiento: el punto del historial de "Acciones"
	if crecimiento := resultado["crecimiento"].([]map[string]interface{}); len(crecimiento) != 1 || crecimiento[0]["valor"] != 1050.0 {
		t.Errorf("crecimiento mal: %v", crecimiento)
	}
}
