package services

import (
	"fmt"
	"strings"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

var (
	tiposPqr       = map[string]bool{"peticion": true, "queja": true, "reclamo": true, "sugerencia": true}
	prioridadesPqr = map[string]bool{"alta": true, "media": true, "baja": true}
)

const estadoInicialPqr = "Abierta"

// PqrDeUsuario lista las PQR de un usuario con su estado y sus adjuntos.
func PqrDeUsuario(idUsuario int) ([]models.PqrVista, error) {
	lista := []models.PqrVista{}
	if err := ExisteUsuario(idUsuario); err != nil {
		return lista, err
	}

	var pqrs []models.Pqr
	if err := helpers.Consultar("soporte", "pqr?"+helpers.Filtro("id_usuario", idUsuario, "activo", true)+"&sortby=fecha_creacion&order=desc", &pqrs); err != nil {
		return lista, err
	}
	estados, err := estadosPorId()
	if err != nil {
		return lista, err
	}

	for _, p := range pqrs {
		lista = append(lista, armarPqr(p, estados))
	}
	return lista, nil
}

// CrearPqr valida los datos, genera el numero de radicado y guarda la PQR
// con el estado "Abierta".
func CrearPqr(datos models.DatosPqr) (models.PqrVista, error) {
	datos.Tipo = strings.ToLower(strings.TrimSpace(datos.Tipo))
	datos.Prioridad = strings.ToLower(strings.TrimSpace(datos.Prioridad))
	if datos.Prioridad == "" {
		datos.Prioridad = "media"
	}
	switch {
	case datos.IdUsuario <= 0:
		return models.PqrVista{}, helpers.NuevoError(400, "id_usuario es obligatorio")
	case strings.TrimSpace(datos.Titulo) == "":
		return models.PqrVista{}, helpers.NuevoError(400, "El titulo es obligatorio")
	case strings.TrimSpace(datos.Descripcion) == "":
		return models.PqrVista{}, helpers.NuevoError(400, "La descripcion es obligatoria")
	case !tiposPqr[datos.Tipo]:
		return models.PqrVista{}, helpers.NuevoError(400, "El tipo debe ser: peticion, queja, reclamo o sugerencia")
	case !prioridadesPqr[datos.Prioridad]:
		return models.PqrVista{}, helpers.NuevoError(400, "La prioridad debe ser: alta, media o baja")
	}
	if err := ExisteUsuario(datos.IdUsuario); err != nil {
		return models.PqrVista{}, err
	}

	var estados []models.EstadoPqr
	if err := helpers.Consultar("soporte", "estado_pqr?"+helpers.Filtro("nombre", estadoInicialPqr), &estados); err != nil {
		return models.PqrVista{}, err
	}
	if len(estados) == 0 {
		return models.PqrVista{}, helpers.NuevoError(500, "No existe el estado '"+estadoInicialPqr+"' en la base de datos")
	}

	pqr := models.Pqr{
		IdUsuario:   datos.IdUsuario,
		Radicado:    nuevoRadicado(),
		Titulo:      strings.TrimSpace(datos.Titulo),
		Tipo:        datos.Tipo,
		Categoria:   datos.Categoria,
		Prioridad:   datos.Prioridad,
		Descripcion: strings.TrimSpace(datos.Descripcion),
		IdEstado:    &models.Ref{Id: estados[0].Id},
		Activo:      true,
	}
	if err := helpers.Enviar("POST", "soporte", "pqr", pqr, &pqr); err != nil {
		return models.PqrVista{}, err
	}

	RegistrarActividad(datos.IdUsuario, "CREAR_PQR", "Se radico la PQR "+pqr.Radicado, "soporte.pqr")

	return armarPqr(pqr, map[int]string{estados[0].Id: estados[0].Nombre}), nil
}

// RegistrarActividad deja una huella en soporte.registro_actividad.
// Si falla no detiene la operacion principal: es solo un registro.
func RegistrarActividad(idUsuario int, tipo, descripcion, entidad string) {
	actividad := models.RegistroActividad{
		IdUsuario:       &idUsuario,
		TipoActividad:   tipo,
		Descripcion:     descripcion,
		EntidadAfectada: entidad,
	}
	helpers.Enviar("POST", "soporte", "registro_actividad", actividad, nil)
}

func armarPqr(p models.Pqr, estados map[int]string) models.PqrVista {
	vista := models.PqrVista{
		IdPqr:               p.Id,
		Radicado:            p.Radicado,
		Titulo:              p.Titulo,
		Tipo:                p.Tipo,
		Categoria:           p.Categoria,
		Prioridad:           p.Prioridad,
		Descripcion:         p.Descripcion,
		Asesor:              p.Asesor,
		Respuesta:           p.MensajeRespuesta,
		FechaCreacion:       p.FechaCreacion,
		UltimaActualizacion: p.FechaModificacion,
		Adjuntos:            []models.AdjuntoVista{},
	}
	if p.IdEstado != nil {
		vista.Estado = estados[p.IdEstado.Id]
	}

	var adjuntos []models.Adjunto
	helpers.Consultar("soporte", "adjunto?"+helpers.Filtro("id_pqr", p.Id, "activo", true), &adjuntos)
	for _, a := range adjuntos {
		vista.Adjuntos = append(vista.Adjuntos, models.AdjuntoVista{
			Nombre:      a.NombreArchivo,
			Ruta:        a.RutaArchivo,
			Tipo:        a.TipoMime,
			TamanoBytes: a.TamanoBytes,
		})
	}
	return vista
}

func estadosPorId() (map[int]string, error) {
	var estados []models.EstadoPqr
	if err := helpers.Consultar("soporte", "estado_pqr?limit=1000", &estados); err != nil {
		return nil, err
	}
	resultado := map[int]string{}
	for _, e := range estados {
		resultado[e.Id] = e.Nombre
	}
	return resultado, nil
}

// nuevoRadicado genera un numero unico, por ejemplo PQR-20261007-154233-512.
func nuevoRadicado() string {
	ahora := time.Now()
	return fmt.Sprintf("PQR-%s-%03d", ahora.Format("20060102-150405"), ahora.Nanosecond()/1e6)
}
