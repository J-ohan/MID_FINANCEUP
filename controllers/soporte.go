package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type EstadoPqr struct {
	ID                 int        `json:"id_estado"`
	Nombre             string     `json:"nombre"`
	Descripcion        *string    `json:"descripcion"`
}

type Pqr struct {
	ID                 int       `json:"id_pqr"`
	IDUsuario          int       `json:"id_usuario"`
	Descripcion        string    `json:"descripcion"`
	IDEstado           int       `json:"id_estado"`
}

type Adjunto struct {
	ID                 int        `json:"id_adjunto"`
	IDPqr              int        `json:"id_pqr"`
	NombreArchivo      string     `json:"nombre_archivo"`
	RutaArchivo        string     `json:"ruta_archivo"`
	TipoMime           *string    `json:"tipo_mime"`
	TamanoBytes        *int       `json:"tamano_bytes"`
}


type RegistroActividad struct {
	ID               int        `json:"id_actividad"`
	IDUsuario        *int       `json:"id_usuario"`
	TipoActividad    *string    `json:"tipo_actividad"`
	Descripcion      *string    `json:"descripcion"`
	EntidadAfectada  *string    `json:"entidad_afectada"`
}

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
