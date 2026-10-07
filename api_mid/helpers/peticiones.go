package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

// cliente HTTP que usa el MID para hablar con los CRUD.
// Si un CRUD no responde en 10 segundos se corta la peticion.
var cliente = &http.Client{Timeout: 10 * time.Second}

// Conjuntos validos y la llave de app.conf donde esta su URL.
var conjuntos = map[string]string{
	"auth":      "url_crud_auth",
	"finanzas":  "url_crud_finanzas",
	"educacion": "url_crud_educacion",
	"negocio":   "url_crud_negocio",
	"soporte":   "url_crud_soporte",
}

// ErrorMid es un error con el codigo HTTP que debe devolver el MID.
type ErrorMid struct {
	Status  int
	Mensaje string
}

func (e *ErrorMid) Error() string {
	return e.Mensaje
}

// NuevoError crea un ErrorMid, por ejemplo NuevoError(400, "El correo es obligatorio").
func NuevoError(status int, mensaje string) *ErrorMid {
	return &ErrorMid{Status: status, Mensaje: mensaje}
}

// ExisteConjunto indica si el nombre corresponde a uno de los 5 CRUD.
func ExisteConjunto(conjunto string) bool {
	_, ok := conjuntos[conjunto]
	return ok
}

// UrlCrud arma la URL de un recurso de un CRUD.
// Ejemplo: UrlCrud("auth", "usuario/1") -> http://localhost:8081/v1/usuario/1
func UrlCrud(conjunto, recurso string) string {
	base, _ := beego.AppConfig.String(conjuntos[conjunto])
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(recurso, "/")
}

// Filtro arma el parametro "query" que entienden los CRUD de bee.
// Ejemplo: Filtro("id_usuario", 1, "activo", true) -> "query=id_usuario:1,activo:true&limit=1000"
func Filtro(pares ...interface{}) string {
	condiciones := []string{}
	for i := 0; i+1 < len(pares); i += 2 {
		condiciones = append(condiciones, fmt.Sprintf("%v:%v", pares[i], pares[i+1]))
	}
	return "query=" + url.QueryEscape(strings.Join(condiciones, ",")) + "&limit=1000"
}

// Consultar hace un GET a un CRUD y guarda el JSON recibido en "destino".
func Consultar(conjunto, recurso string, destino interface{}) error {
	return Enviar("GET", conjunto, recurso, nil, destino)
}

// Enviar hace una peticion (GET, POST, PUT o DELETE) a un CRUD.
// "cuerpo" se envia como JSON y la respuesta se guarda en "destino" (puede ser nil).
func Enviar(metodo, conjunto, recurso string, cuerpo interface{}, destino interface{}) error {
	direccion := UrlCrud(conjunto, recurso)

	var lector io.Reader
	if cuerpo != nil {
		datos, err := json.Marshal(cuerpo)
		if err != nil {
			return NuevoError(500, "No fue posible preparar los datos para el CRUD de "+conjunto)
		}
		lector = bytes.NewReader(datos)
	}

	peticion, err := http.NewRequest(metodo, direccion, lector)
	if err != nil {
		return NuevoError(500, "Direccion invalida: "+direccion)
	}
	peticion.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return NuevoError(502, "No fue posible comunicarse con el CRUD de "+conjunto+". Verifica que este encendido.")
	}
	defer respuesta.Body.Close()

	datos, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return NuevoError(502, "Error leyendo la respuesta del CRUD de "+conjunto)
	}

	if respuesta.StatusCode == http.StatusNotFound {
		return NuevoError(404, "La ruta "+recurso+" no existe en el CRUD de "+conjunto)
	}
	if respuesta.StatusCode >= 300 {
		return NuevoError(502, fmt.Sprintf("El CRUD de %s respondio con error %d", conjunto, respuesta.StatusCode))
	}

	// Los CRUD que genera bee responden los errores como un texto entre comillas
	// (con codigo 200). Ejemplo: "<QuerySeter> no row found". Aqui se detectan.
	// (Una lista vacia llega como null; esa no es un error.)
	var texto string
	if bytes.HasPrefix(bytes.TrimSpace(datos), []byte(`"`)) && json.Unmarshal(datos, &texto) == nil {
		switch {
		case texto == "OK":
			return nil
		case strings.Contains(texto, "no row found"):
			return NuevoError(404, "No se encontro el registro en el CRUD de "+conjunto)
		default:
			return NuevoError(400, "El CRUD de "+conjunto+" rechazo la operacion: "+texto)
		}
	}

	if destino != nil {
		if err := json.Unmarshal(datos, destino); err != nil {
			return NuevoError(502, "El CRUD de "+conjunto+" devolvio datos que no se pudieron leer")
		}
	}
	return nil
}

// Reenviar pasa una peticion tal cual a un CRUD y devuelve el codigo y el cuerpo
// de su respuesta. Lo usa la pasarela /v1/crud/... del MID.
func Reenviar(metodo, conjunto, recurso, consulta string, cuerpo []byte) (int, []byte, error) {
	direccion := UrlCrud(conjunto, recurso)
	if consulta != "" {
		direccion += "?" + consulta
	}
	peticion, err := http.NewRequest(metodo, direccion, bytes.NewReader(cuerpo))
	if err != nil {
		return 0, nil, NuevoError(400, "Direccion invalida: "+direccion)
	}
	peticion.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return 0, nil, NuevoError(502, "No fue posible comunicarse con el CRUD de "+conjunto+". Verifica que este encendido.")
	}
	defer respuesta.Body.Close()

	datos, err := io.ReadAll(respuesta.Body)
	return respuesta.StatusCode, datos, err
}

// EstaEncendido revisa si un CRUD responde.
func EstaEncendido(conjunto string) bool {
	base, _ := beego.AppConfig.String(conjuntos[conjunto])
	respuesta, err := cliente.Get(base)
	if err != nil {
		return false
	}
	respuesta.Body.Close()
	return true
}
