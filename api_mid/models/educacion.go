package models

import "time"

// ------------------------------------------------------------
// Tablas de crud_educacion (como llegan del CRUD)
// ------------------------------------------------------------

type Escuela struct {
	Id          int    `json:"Id"`
	Slug        string `json:"Slug"`
	Nombre      string `json:"Nombre"`
	Descripcion string `json:"Descripcion"`
	Orden       int    `json:"Orden"`
}

type Instructor struct {
	Id        int    `json:"Id"`
	Nombre    string `json:"Nombre"`
	Cargo     string `json:"Cargo"`
	Iniciales string `json:"Iniciales"`
}

type ModuloEducativo struct {
	Id                int       `json:"Id"`
	IdEscuela         *Ref      `json:"IdEscuela"`
	IdInstructor      *Ref      `json:"IdInstructor"`
	Slug              string    `json:"Slug"`
	Titulo            string    `json:"Titulo"`
	Descripcion       string    `json:"Descripcion"`
	Nivel             string    `json:"Nivel"`
	Categoria         string    `json:"Categoria"`
	Formato           string    `json:"Formato"`
	Duracion          string    `json:"Duracion"`
	ResumenLecciones  string    `json:"ResumenLecciones"`
	OtorgaCertificado bool      `json:"OtorgaCertificado"`
	Estudiantes       int       `json:"Estudiantes"`
	Calificacion      float64   `json:"Calificacion"`
	UrlThumbnail      string    `json:"UrlThumbnail"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type ModuloDetalle struct {
	Id    int    `json:"Id"`
	Tipo  string `json:"Tipo"`
	Texto string `json:"Texto"`
}

type Leccion struct {
	Id              int    `json:"Id"`
	IdModulo        *Ref   `json:"IdModulo"`
	IdContenido     *Ref   `json:"IdContenido"`
	Titulo          string `json:"Titulo"`
	Tipo            string `json:"Tipo"`
	Descripcion     string `json:"Descripcion"`
	DuracionMinutos int    `json:"DuracionMinutos"`
	UrlVideo        string `json:"UrlVideo"`
	NumeroLeccion   int    `json:"NumeroLeccion"`
}

type Contenido struct {
	Id          int    `json:"Id"`
	Titulo      string `json:"Titulo"`
	Descripcion string `json:"Descripcion"`
	UrlVideo    string `json:"UrlVideo"`
}

type ProgresoEducativo struct {
	Id                   int       `json:"Id"`
	IdUsuario            int       `json:"IdUsuario"`
	IdModulo             *Ref      `json:"IdModulo"`
	PorcentajeCompletado int       `json:"PorcentajeCompletado"`
	FechaCompletado      time.Time `json:"FechaCompletado"`
	Calificacion         int       `json:"Calificacion"`
	Activo               bool      `json:"Activo"`
}

type ProgresoLeccion struct {
	Id              int       `json:"Id"`
	IdUsuario       int       `json:"IdUsuario"`
	IdLeccion       *Ref      `json:"IdLeccion"`
	Completado      bool      `json:"Completado"`
	FechaCompletado time.Time `json:"FechaCompletado"`
	Activo          bool      `json:"Activo"`
}

type Certificado struct {
	Id                 int       `json:"Id"`
	IdUsuario          int       `json:"IdUsuario"`
	IdModulo           *Ref      `json:"IdModulo"`
	CodigoVerificacion string    `json:"CodigoVerificacion"`
	FechaEmision       time.Time `json:"FechaEmision"`
	Activo             bool      `json:"Activo"`
}

// ------------------------------------------------------------
// Datos que manda la pagina
// ------------------------------------------------------------

type DatosUsuarioCurso struct {
	IdUsuario int `json:"id_usuario"`
	IdModulo  int `json:"id_modulo"`
}
