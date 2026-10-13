package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// ============================================================
// Funciones para hablar con los CRUD.
// Todas hacen lo mismo que el ejemplo: llamar por HTTP,
// leer la respuesta y convertir el JSON a una estructura.
// ============================================================

// ConsultarUno trae UN registro (GET /v1/tabla/5).
// Si el CRUD no responde: error 502.
// Si el registro no existe (el CRUD responde un texto): error 404 con mensajeNoExiste.
func ConsultarUno(url string, destino interface{}, mensajeNoExiste string) error {

	response, err := http.Get(url)

	if err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible comunicarse con un CRUD: "+url)
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	if err := json.Unmarshal(body, destino); err != nil {
		return NuevoError(http.StatusNotFound, mensajeNoExiste)
	}

	return nil
}

// ConsultarLista trae una LISTA (GET /v1/tabla?query=...).
// Si el CRUD no responde: error 502.
// Si no hay datos, la lista queda vacia (no es error).
func ConsultarLista(url string, destino interface{}) error {

	response, err := http.Get(url)

	if err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible comunicarse con un CRUD: "+url)
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	// Si el CRUD manda null o un texto de error, la lista queda vacia
	json.Unmarshal(body, destino)

	return nil
}

// Crear guarda un registro nuevo (POST /v1/tabla).
// El CRUD responde 201 y el registro creado (con su Id), que
// se guarda en "respuesta". Si no se necesita, se pasa nil.
func Crear(url string, datos interface{}, respuesta interface{}) error {

	jsonDatos, _ := json.Marshal(datos)

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonDatos))

	if err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible comunicarse con un CRUD: "+url)
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusCreated {
		return NuevoError(http.StatusBadRequest, "No se pudo guardar: "+string(body))
	}

	if respuesta != nil {
		json.Unmarshal(body, respuesta)
	}

	return nil
}

// Actualizar reemplaza un registro (PUT /v1/tabla/5).
// OJO: el PUT del CRUD reemplaza TODO el registro, por eso
// siempre se manda el registro completo.
func Actualizar(url string, datos interface{}) error {

	jsonDatos, _ := json.Marshal(datos)

	// Go no tiene http.Put, se arma la peticion a mano
	peticion, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonDatos))
	peticion.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(peticion)

	if err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible comunicarse con un CRUD: "+url)
	}

	response.Body.Close()

	return nil
}
