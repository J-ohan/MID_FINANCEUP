package models

import "time"

// ============================================================
// MODELOS DE FINANZAS
// Un "modelo" es solo la forma que tienen los datos: que campos
// trae y de que tipo es cada uno. No hace nada por si solo.
//
// Este archivo tiene dos partes:
//   1. Las tablas de crud_finanzas, tal como el CRUD las manda.
//      Los nombres van en mayuscula (Id, IdUsuario, Monto...)
//      porque asi los escribe el CRUD.
//   2. Los datos que manda la pagina (Angular).
//      Los nombres van en minuscula (id_usuario, valor...)
//      porque asi los escribe la pagina.
//
// Lo que esta entre comillas invertidas (`json:"..."`) le dice
// a Go con que nombre viene cada campo en el JSON.
// ============================================================

// ------------------------------------------------------------
// 1. Tablas de crud_finanzas (como llegan del CRUD)
// ------------------------------------------------------------

// Categoria: tabla finanzas.categoria (Vivienda, Salario, Gastos hormiga...)
type Categoria struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
	Icono  string `json:"Icono"`
}

// Movimiento: tabla finanzas.movimiento_ingreso_egreso.
// Cada fila es un ingreso o un gasto del libro mayor.
type Movimiento struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdCategoria   *Ref      `json:"IdCategoria"` // llave foranea: llega como {"Id": 3}; puede venir vacia
	Nombre        string    `json:"Nombre"`      // el concepto: "Arriendo", "Sueldo"...
	Monto         float64   `json:"Monto"`
	EsIngreso     bool      `json:"EsIngreso"` // true = ingreso, false = gasto
	Fecha         time.Time `json:"Fecha"`
	MetodoPago    string    `json:"MetodoPago"`
	Observaciones string    `json:"Observaciones"`
	Activo        bool      `json:"Activo"` // false = borrado
}

// Meta: tabla finanzas.meta (metas de ahorro).
type Meta struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	Nombre        string    `json:"Nombre"`
	Descripcion   string    `json:"Descripcion"`
	MontoObjetivo float64   `json:"MontoObjetivo"` // cuanto quiere ahorrar
	MontoActual   float64   `json:"MontoActual"`   // cuanto lleva ahorrado
	FechaLimite   time.Time `json:"FechaLimite"`
	Icono         string    `json:"Icono"`
	Color         string    `json:"Color"`
	Activo        bool      `json:"Activo"`
}

// NivelRiesgo: tabla finanzas.nivel_riesgo (Bajo, Medio, Alto).
type NivelRiesgo struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

// Inversion: tabla finanzas.inversion.
type Inversion struct {
	Id            int       `json:"Id"`
	IdUsuario     int       `json:"IdUsuario"`
	IdNivelRiesgo *Ref      `json:"IdNivelRiesgo"` // llave foranea a nivel_riesgo
	Nombre        string    `json:"Nombre"`
	Monto         float64   `json:"Monto"`       // lo que se invirtio
	Rendimiento   float64   `json:"Rendimiento"` // lo que se gano (o perdio, si es negativo)
	Duracion      string    `json:"Duracion"`    // texto: "1 año", "6 meses"...
	FechaInicio   time.Time `json:"FechaInicio"`
	FechaFin      time.Time `json:"FechaFin"`
	Activo        bool      `json:"Activo"`
}

// HistorialInversion: tabla finanzas.historial_inversion.
// Guarda el valor de una inversion en cada fecha (para la grafica).
type HistorialInversion struct {
	Id          int       `json:"Id"`
	IdInversion *Ref      `json:"IdInversion"`
	Fecha       time.Time `json:"Fecha"`
	Valor       float64   `json:"Valor"`
	Activo      bool      `json:"Activo"`
}

// ------------------------------------------------------------
// 2. Datos que manda la pagina
// ------------------------------------------------------------

// FilaExcel: una fila de la plantilla de Excel que leyo la pagina.
type FilaExcel struct {
	Fecha         string  `json:"fecha"` // AAAA-MM-DD
	Concepto      string  `json:"concepto"`
	Categoria     string  `json:"categoria"` // el nombre; es opcional
	Tipo          string  `json:"tipo"`      // "ingreso" o "gasto"
	Valor         float64 `json:"valor"`
	MetodoPago    string  `json:"metodoPago"` // asi lo llama el cliente (services/finanzas.ts)
	Observaciones string  `json:"observaciones"`
}

// DatosImportar: lo que llega a POST /v1/movimientos/importar.
// Modo "validar" solo revisa las filas; "confirmar" (o vacio) las guarda.
type DatosImportar struct {
	IdUsuario   int         `json:"id_usuario"`
	Modo        string      `json:"modo"`
	Movimientos []FilaExcel `json:"movimientos"`
}

// ErrorFila: una fila del Excel que no se pudo importar y por que.
// Es la misma forma que usa la pagina: {fila, motivo}.
type ErrorFila struct {
	Fila     int    `json:"fila"`
	Concepto string `json:"concepto"`
	Motivo   string `json:"motivo"`
}

// DatosAbono: lo que llega a PUT /v1/metas/:id/abono.
type DatosAbono struct {
	Valor float64 `json:"valor"`
}

// FiltrosMovimientos: los filtros opcionales del libro mayor.
// Llegan en la URL, por ejemplo: ?tipo=gasto&categoria=3&desde=2026-10-01
type FiltrosMovimientos struct {
	Tipo        string // "ingreso", "gasto", "todos" o vacio
	IdCategoria int    // 0 = todas
	Categoria   string // el nombre de la categoria, si no llega el id
	Desde       string // AAAA-MM-DD o vacio
	Hasta       string // AAAA-MM-DD o vacio
	Texto       string // busca en el concepto
}
