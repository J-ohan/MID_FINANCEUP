package services

import (
	"net/http"

	"api_mid_financeup/helpers"
	"api_mid_financeup/services/finanzas"
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

// Dashboard junta perfil, resumen financiero, progreso, solicitudes y PQR.
func Dashboard(idUsuario int) (map[string]interface{}, error) {

	perfil, err := ObtenerPerfil(idUsuario)

	if err != nil {
		return nil, err
	}

	dashboard := map[string]interface{}{"perfil": perfil}
	avisos := []string{}

	if datos, err := finanzas.ResumenFinanciero(idUsuario); err == nil {
		dashboard["finanzas"] = datos
	} else {
		avisos = append(avisos, "finanzas: "+err.Error())
	}

	if datos, err := ProgresoEducativo(idUsuario); err == nil {
		dashboard["educacion"] = datos
	} else {
		avisos = append(avisos, "educacion: "+err.Error())
	}

	if datos, err := SolicitudesUsuario(idUsuario); err == nil {
		dashboard["solicitudes_credito"] = datos
	} else {
		avisos = append(avisos, "solicitudes_credito: "+err.Error())
	}

	if datos, err := PqrUsuario(idUsuario); err == nil {
		dashboard["pqr"] = datos
	} else {
		avisos = append(avisos, "pqr: "+err.Error())
	}

	if len(avisos) > 0 {
		dashboard["avisos"] = avisos
	}

	return dashboard, nil
}
