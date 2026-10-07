package services

import (
	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// Dashboard junta en una sola respuesta la informacion de los 5 CRUD.
// Solo el perfil es obligatorio: si otra seccion falla (por ejemplo porque
// ese CRUD esta apagado) se agrega un aviso y se entrega el resto.
func Dashboard(idUsuario int) (models.Dashboard, error) {
	panel := models.Dashboard{
		SolicitudesCredito: []models.SolicitudCredito{},
		Pqr:                []models.PqrVista{},
	}

	perfil, err := ObtenerPerfil(idUsuario)
	if err != nil {
		return panel, err
	}
	panel.Perfil = perfil

	if finanzas, err := ResumenFinanciero(idUsuario); err == nil {
		panel.Finanzas = &finanzas
	} else {
		panel.Avisos = append(panel.Avisos, "finanzas: "+err.Error())
	}

	if educacion, err := ProgresoEducativo(idUsuario); err == nil {
		panel.Educacion = &educacion
	} else {
		panel.Avisos = append(panel.Avisos, "educacion: "+err.Error())
	}

	if solicitudes, err := SolicitudesDeUsuario(idUsuario); err == nil {
		panel.SolicitudesCredito = solicitudes
	} else {
		panel.Avisos = append(panel.Avisos, "negocio: "+err.Error())
	}

	if pqrs, err := PqrDeUsuario(idUsuario); err == nil {
		panel.Pqr = pqrs
	} else {
		panel.Avisos = append(panel.Avisos, "soporte: "+err.Error())
	}

	return panel, nil
}

// EstadoServicios indica que CRUD estan encendidos.
func EstadoServicios() map[string]string {
	estado := map[string]string{"mid": "encendido"}
	for _, conjunto := range []string{"auth", "finanzas", "educacion", "negocio", "soporte"} {
		if helpers.EstaEncendido(conjunto) {
			estado["crud_"+conjunto] = "encendido"
		} else {
			estado["crud_"+conjunto] = "apagado"
		}
	}
	return estado
}
