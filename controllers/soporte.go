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