package services

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

func RegistrarActividad(idUsuario int, tipo string, descripcion string, entidad string) {

	actividad := models.RegistroActividad{
		IdUsuario:       &idUsuario,
		TipoActividad:   tipo,
		Descripcion:     descripcion,
		EntidadAfectada: entidad,
	}

	helpers.Crear(helpers.UrlCrudSoporte+"/registro_actividad", actividad, nil)
}

func estadosPorId() (map[int]string, error) {

	var estados []models.EstadoPqr

	if err := helpers.ConsultarLista(helpers.UrlCrudSoporte+"/estado_pqr?limit=1000", &estados); err != nil {
		return nil, err
	}

	nombres := map[int]string{}
	for _, estado := range estados {
		nombres[estado.Id] = estado.Nombre
	}

	return nombres, nil
}

func buscarEstado(nombre string) (models.EstadoPqr, bool) {

	var estados []models.EstadoPqr
	helpers.ConsultarLista(helpers.UrlCrudSoporte+"/estado_pqr?query=nombre:"+url.QueryEscape(nombre), &estados)

	if len(estados) == 0 {
		return models.EstadoPqr{}, false
	}

	return estados[0], true
}

func RadicarPqr(datos models.DatosPqr) (map[string]interface{}, error) {

	// 1. Validar
	if datos.Titulo == "" || datos.Descripcion == "" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El titulo y la descripcion son obligatorios")
	}

	if datos.Tipo != "peticion" && datos.Tipo != "queja" && datos.Tipo != "reclamo" && datos.Tipo != "sugerencia" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser: peticion, queja, reclamo o sugerencia")
	}

	if datos.Prioridad == "" {
		datos.Prioridad = "media"
	}

	// 2. Revisar el usuario y buscar el estado "Abierta"
	if _, err := BuscarUsuario(datos.IdUsuario); err != nil {
		return nil, err
	}

	estado, existe := buscarEstado("Abierta")

	if !existe {
		return nil, helpers.NuevoError(http.StatusInternalServerError, "No existe el estado Abierta en la base de datos")
	}

	// 3. Guardar la PQR (radicado: PQR- + fecha y hora, ej: PQR-20261009-153045)
	pqr := models.Pqr{
		IdUsuario:   datos.IdUsuario,
		Radicado:    "PQR-" + time.Now().Format("20060102-150405"),
		Titulo:      datos.Titulo,
		Tipo:        datos.Tipo,
		Categoria:   datos.Categoria,
		Prioridad:   datos.Prioridad,
		Descripcion: datos.Descripcion,
		IdEstado:    &models.Ref{Id: estado.Id},
		Activo:      true,
	}

	if err := helpers.Crear(helpers.UrlCrudSoporte+"/pqr", pqr, &pqr); err != nil {
		return nil, err
	}
}
