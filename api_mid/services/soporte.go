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

// estadosPorId devuelve un mapa id -> nombre del estado.
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

// buscarEstado busca un estado por su nombre (ej: "Abierta").
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
	

	adjuntosGuaradados := 0

	for _, archivo :=range datos.Adjuntos{

		adjunto := models.Adjunto{
			IdPqr: &models.Ref{Id: pqr.Id},
			NombreArchivo: archivo.NombreArchivo,
			RutaArchivo: archivo.RutaArchivo,
			TipoMime: archivo.TipoMime,
			TamanoBytes: archivo.TamanoBytes,
			Activo: true,

		}
		if helpers.Crear(helpers.UrlCrudSoporte+"/abjunto", adjunto, nil)== nil{  
			adjuntosGuaradados++
		}
	}

	RegistrarActividad(datos.IdUsuario, "CREAR_PQR", "se radico la pqr"+pqr.Radicado, "soporte.pqr")

	return map[string]interface{}{
		"id_pqr": pqr.Id,
		"radicado": pqr.Radicado,
		"titulo": pqr.Titulo,
		"tipo": pqr.Tipo,
		"prioridad": pqr.Prioridad,
		"estado": pqr.IdEstado,
		"adjuntos": adjuntosGuaradados,
	}, nil
}

func PqrUsuario(idUsuario int) ([]map[string]interface{}, error) {

	var pqrs []models.Pqr

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/pqr?query=id_usuario:%d,activo:true&sortby=Id&order=desc&limit=1000", helpers.UrlCrudSoporte, idUsuario), &pqrs); err != nil {
		return nil, err
	}

	estados, err := estadosPorId()

	if err != nil{
		return nil, err
	}

	lista := []map[string]interface{}{}

	for _, pqr := range pqrs {
		lista = append(lista, map[string]interface{}{
			"id_pqr":         pqr.Id,
			"radicado":       pqr.Radicado,
			"titulo":         pqr.Titulo,
			"tipo":           pqr.Tipo,
			"categoria":      pqr.Categoria,
			"prioridad":      pqr.Prioridad,
			"estado":         estados[pqr.IdEstado.Id],
			"progreso":       pqr.Progreso,
			"descripcion":    pqr.Descripcion,
			"fecha_creacion": pqr.FechaCreacion,
		})
	}

	return lista, nil
}

func VerPqr(radicado string) (map[string]interface{}, error) {

	// 1. Buscar la PQR
	var pqrs []models.Pqr

	if err := helpers.ConsultarLista(helpers.UrlCrudSoporte+"/pqr?query=radicado:"+url.QueryEscape(radicado)+"&limit=1", &pqrs); err != nil {
		return nil, err
	}

	if len(pqrs) == 0 {
		return nil, helpers.NuevoError(http.StatusNotFound, "No existe una PQR con ese radicado")
	}

	pqr := pqrs[0]

	// 2. Su estado
	estados, err := estadosPorId()

	if err != nil {
		return nil, err
	}

	// 3. Respuestas (de la mas vieja a la mas nueva) y adjuntos
	var respuestas []models.RespuestaPqr
	var adjuntos []models.Adjunto

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/respuesta_pqr?query=id_pqr:%d,activo:true&sortby=Id&order=asc&limit=1000", helpers.UrlCrudSoporte, pqr.Id), &respuestas); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/adjunto?query=id_pqr:%d,activo:true&limit=1000", helpers.UrlCrudSoporte, pqr.Id), &adjuntos); err != nil {
		return nil, err
	}

	listaRespuestas := []map[string]interface{}{}
	for _, respuesta := range respuestas {
		listaRespuestas = append(listaRespuestas, map[string]interface{}{"autor": respuesta.Autor, "mensaje": respuesta.Mensaje, "fecha": respuesta.FechaRespuesta})
	}

	listaAdjuntos := []map[string]interface{}{}
	for _, adjunto := range adjuntos {
		listaAdjuntos = append(listaAdjuntos, map[string]interface{}{"nombre": adjunto.NombreArchivo, "ruta": adjunto.RutaArchivo, "tipo": adjunto.TipoMime, "tamano": adjunto.TamanoBytes})
	}

	return map[string]interface{}{
		"id_pqr":               pqr.Id,
		"radicado":             pqr.Radicado,
		"titulo":               pqr.Titulo,
		"tipo":                 pqr.Tipo,
		"categoria":            pqr.Categoria,
		"prioridad":            pqr.Prioridad,
		"descripcion":          pqr.Descripcion,
		"estado":               estados[pqr.IdEstado.Id],
		"progreso":             pqr.Progreso,
		"asesor":               pqr.Asesor,
		"mensaje_respuesta":    pqr.MensajeRespuesta,
		"fecha_creacion":       pqr.FechaCreacion,
		"ultima_actualizacion": pqr.FechaModificacion,
		"respuestas":           listaRespuestas,
		"adjuntos":             listaAdjuntos,
	}, nil
}