package services

import (
	"net/http"

	"api_mid_financeup/helpers"
)

// ============================================================
// SERVICE GENERAL (lo usan todos para probar)
// ============================================================

// EstadoCruds intenta llamar a los 5 CRUD y dice cuales estan encendidos.
func EstadoCruds() map[string]string {

	cruds := map[string]string{
		"crud_auth":      helpers.UrlCrudAuth,
		"crud_finanzas":  helpers.UrlCrudFinanzas,
		"crud_educacion": helpers.UrlCrudEducacion,
		"crud_negocio":   helpers.UrlCrudNegocio,
		"crud_soporte":   helpers.UrlCrudSoporte,
	}

	estado := map[string]string{"mid": "encendido"}

	for nombre, direccion := range cruds {

		response, err := http.Get(direccion)

		if err != nil {
			estado[nombre] = "apagado"
			continue
		}

		response.Body.Close()
		estado[nombre] = "encendido"
	}

	return estado
}

// Pendiente: Dashboard (junta perfil, resumen financiero, progreso,
// solicitudes y PQR). Se arma cuando esten los services de todos.
