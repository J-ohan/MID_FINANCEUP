<<<<<<< HEAD
package helpers
=======
package helpers

import "os"

// ============================================================
// Direcciones de los 5 CRUD
// Tienen un valor por defecto (localhost) y se pueden cambiar
// en el archivo .env sin tocar el codigo.
// ============================================================

var (
	UrlCrudAuth      = "http://localhost:8081/v1"
	UrlCrudFinanzas  = "http://localhost:8082/v1"
	UrlCrudEducacion = "http://localhost:8083/v1"
	UrlCrudNegocio   = "http://localhost:8084/v1"
	UrlCrudSoporte   = "http://localhost:8085/v1"
)

// CargarUrls lee las direcciones del .env (si existen).
// Se llama una sola vez desde main.go al arrancar.
func CargarUrls() {
	if valor := os.Getenv("URL_CRUD_AUTH"); valor != "" {
		UrlCrudAuth = valor
	}
	if valor := os.Getenv("URL_CRUD_FINANZAS"); valor != "" {
		UrlCrudFinanzas = valor
	}
	if valor := os.Getenv("URL_CRUD_EDUCACION"); valor != "" {
		UrlCrudEducacion = valor
	}
	if valor := os.Getenv("URL_CRUD_NEGOCIO"); valor != "" {
		UrlCrudNegocio = valor
	}
	if valor := os.Getenv("URL_CRUD_SOPORTE"); valor != "" {
		UrlCrudSoporte = valor
	}
}
