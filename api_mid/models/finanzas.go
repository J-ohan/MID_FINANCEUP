package models

import "time"

// ------------------------------------------------------------
// Tablas de crud_finanzas (como llegan del CRUD)
// ------------------------------------------------------------

type Categoria struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
	Icono  string `json:"Icono"`
}

type Movimiento struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdCategoria   *Ref      `json:"IdCategoria"`
	Nombre        string    `json:"Nombre"`
	Monto         float64   `json:"Monto"`
	EsIngreso     bool      `json:"EsIngreso"`
	Fecha         time.Time `json:"Fecha"`
	MetodoPago    string    `json:"MetodoPago"`
	Observaciones string    `json:"Observaciones"`
	Activo        bool      `json:"Activo"`
}

type Meta struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	Nombre        string    `json:"Nombre"`
	Descripcion   string    `json:"Descripcion"`
	MontoObjetivo float64   `json:"MontoObjetivo"`
	MontoActual   float64   `json:"MontoActual"`
	FechaLimite   time.Time `json:"FechaLimite"`
	Icono         string    `json:"Icono"`
	Color         string    `json:"Color"`
	Activo        bool      `json:"Activo"`
}

type NivelRiesgo struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Inversion struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdNivelRiesgo *Ref      `json:"IdNivelRiesgo"`
	Nombre        string    `json:"Nombre"`
	Monto         float64   `json:"Monto"`
	Rendimiento   float64   `json:"Rendimiento"`
	Duracion      string    `json:"Duracion"`
	FechaInicio   time.Time `json:"FechaInicio"`
	FechaFin      time.Time `json:"FechaFin"`
	Activo        bool      `json:"Activo"`
}

type HistorialInversion struct {
	Id          int       `json:"Id"`
	IdInversion *Ref      `json:"IdInversion"`
	Fecha       time.Time `json:"Fecha"`
	Valor       float64   `json:"Valor"`
	Activo      bool      `json:"Activo"`
}

// ------------------------------------------------------------
// Datos que manda la pagina
// ------------------------------------------------------------

type FilaExcel struct {
	Fecha         string  `json:"fecha"`
	Concepto      string  `json:"concepto"`
	Categoria     string  `json:"categoria"`
	Tipo          string  `json:"tipo"`
	Valor         float64 `json:"valor"`
	MetodoPago    string  `json:"metodo_pago"`
	Observaciones string  `json:"observaciones"`
}

type DatosImportar struct {
	IdUsuario   int         `json:"id_usuario"`
	Movimientos []FilaExcel `json:"movimientos"`
}

type DatosAbono struct {
	Valor float64 `json:"valor"`
}

// DatosMovimiento: crear o editar un movimiento del libro mayor.
// Trae los mismos campos de una fila del Excel mas el usuario.
type DatosMovimiento struct {
	IdUsuario int `json:"id_usuario"`
	FilaExcel
}

// DatosMeta: crear o editar una meta de ahorro.
type DatosMeta struct {
	IdUsuario int     `json:"id_usuario"`
	Nombre    string  `json:"nombre"`
	Icono     string  `json:"icono"`
	Actual    float64 `json:"actual"`
	Objetivo  float64 `json:"objetivo"`
}

// DatosInversion: crear o editar una inversion.
// El riesgo llega como texto: "Bajo", "Medio" o "Alto".
type DatosInversion struct {
	IdUsuario   int     `json:"id_usuario"`
	Nombre      string  `json:"nombre"`
	Monto       float64 `json:"monto"`
	Rendimiento float64 `json:"rendimiento"`
	Riesgo      string  `json:"riesgo"`
	Duracion    string  `json:"duracion"`
}

// FiltrosMovimientos son los filtros opcionales del libro mayor
type FiltrosMovimientos struct {
	Tipo        string // "ingreso", "gasto" o vacio
	IdCategoria int    // 0 = todas
	Desde       string // AAAA-MM-DD o vacio
	Hasta       string // AAAA-MM-DD o vacio
	Texto       string // busca en el concepto
}
