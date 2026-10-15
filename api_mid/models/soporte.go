package models

import "time"

// ------------------------------------------------------------
// Tablas de crud_soporte (como llegan del CRUD)
// ------------------------------------------------------------

type EstadoPqr struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Pqr struct {
	Id                int       `json:"Id"`
	IdUsuario         int       `json:"IdUsuario"`
	Radicado          string    `json:"Radicado"`
	Titulo            string    `json:"Titulo"`
	Tipo              string    `json:"Tipo"`
	Categoria         string    `json:"Categoria"`
	Prioridad         string    `json:"Prioridad"`
	Asesor            string    `json:"Asesor"`
	MensajeRespuesta  string    `json:"MensajeRespuesta"`
	Progreso          int       `json:"Progreso"`
	Descripcion       string    `json:"Descripcion"`
	IdEstado          *Ref      `json:"IdEstado"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type Adjunto struct {
	Id            int    `json:"Id"`
	IdPqr         *Ref   `json:"IdPqr"`
	NombreArchivo string `json:"NombreArchivo"`
	RutaArchivo   string `json:"RutaArchivo"`
	TipoMime      string `json:"TipoMime"`
	TamanoBytes   int    `json:"TamanoBytes"`
	Activo        bool   `json:"Activo"`
}

type RespuestaPqr struct {
	Id             int       `json:"Id"`
	IdPqr          *Ref      `json:"IdPqr"`
	IdUsuario      *int      `json:"IdUsuario"`
	Autor          string    `json:"Autor"`
	Mensaje        string    `json:"Mensaje"`
	FechaRespuesta time.Time `json:"FechaRespuesta"`
	Activo         bool      `json:"Activo"`
}

type RegistroActividad struct {
	IdUsuario       *int   `json:"IdUsuario"`
	TipoActividad   string `json:"TipoActividad"`
	Descripcion     string `json:"Descripcion"`
	EntidadAfectada string `json:"EntidadAfectada"`
}

// ------------------------------------------------------------
// Datos que manda la pagina
// ------------------------------------------------------------

type ArchivoAdjunto struct {
	NombreArchivo string `json:"nombre_archivo"`
	RutaArchivo   string `json:"ruta_archivo"`
	TipoMime      string `json:"tipo_mime"`
	TamanoBytes   int    `json:"tamano_bytes"`
}

type DatosPqr struct {
	IdUsuario   int              `json:"id_usuario"`
	Titulo      string           `json:"titulo"`
	Tipo        string           `json:"tipo"`
	Categoria   string           `json:"categoria"`
	Prioridad   string           `json:"prioridad"`
	Descripcion string           `json:"descripcion"`
	Adjuntos    []ArchivoAdjunto `json:"adjuntos"`
}

type DatosRespuestaPqr struct {
	IdUsuario int    `json:"id_usuario"`
	Autor     string `json:"autor"`
	Mensaje   string `json:"mensaje"`
	Estado    string `json:"estado"`
	Progreso  int    `json:"progreso"`
}
