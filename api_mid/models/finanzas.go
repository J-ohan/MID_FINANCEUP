package models

import "time"

// ============================================================
// TABLAS: JSON que el MID recibe de los CRUD
// Los nombres de campo son los mismos de los modelos de los
// CRUD. Se incluyen todas las columnas porque el PUT del CRUD
// reemplaza el registro completo.
// ============================================================

// Llave foránea tal como la manda el CRUD: {"Id": 3}
type Ref struct {
	Id int `json:"Id"`
}

// auth.usuario (crud_auth)

// finanzas.categoria
type Categoria struct {
	Id          int    `json:"Id"`
	Nombre      string `json:"Nombre"`
	Descripcion string `json:"Descripcion"`
	Icono       string `json:"Icono"`
	Activo      bool   `json:"Activo"`
}

// finanzas.movimiento_ingreso_egreso (Libro mayor)
type MovimientoIngresoEgreso struct {
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

// finanzas.meta
type Meta struct {
	Id                 int       `json:"Id"`
	IdUsuario          int       `json:"IdUsuario"`
	IdEditarMeta       *Ref      `json:"IdEditarMeta"`
	IdMovimientoDinero *Ref      `json:"IdMovimientoDinero"`
	Nombre             string    `json:"Nombre"`
	Descripcion        string    `json:"Descripcion"`
	MontoObjetivo      float64   `json:"MontoObjetivo"`
	MontoActual        float64   `json:"MontoActual"`
	FechaLimite        time.Time `json:"FechaLimite"`
	Color              string    `json:"Color"`
	Icono              string    `json:"Icono"`
	Activo             bool      `json:"Activo"`
}

// finanzas.nivel_riesgo
type NivelRiesgo struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
	Activo bool   `json:"Activo"`
}

// finanzas.inversion
type Inversion struct {
	Id                 int       `json:"Id"`
	IdUsuario          int       `json:"IdUsuario"`
	IdTipoInversion    *Ref      `json:"IdTipoInversion"`
	IdNivelRiesgo      *Ref      `json:"IdNivelRiesgo"`
	IdMovimientoDinero *Ref      `json:"IdMovimientoDinero"`
	Nombre             string    `json:"Nombre"`
	Monto              float64   `json:"Monto"`
	Rentabilidad       float64   `json:"Rentabilidad"`
	Rendimiento        float64   `json:"Rendimiento"`
	Duracion           string    `json:"Duracion"`
	FechaInicio        time.Time `json:"FechaInicio"`
	FechaFin           time.Time `json:"FechaFin"`
	Activo             bool      `json:"Activo"`
}

// finanzas.historial_inversion
type HistorialInversion struct {
	Id          int       `json:"Id"`
	IdInversion *Ref      `json:"IdInversion"`
	Fecha       time.Time `json:"Fecha"`
	Valor       float64   `json:"Valor"`
	Activo      bool      `json:"Activo"`
}

// ============================================================
// SOLICITUDES: JSON que el MID recibe del frontend
// ============================================================

// POST /v1/movimientos y PUT /v1/movimientos/:id
type SolicitudMovimiento struct {
	IdUsuario     int     `json:"idUsuario"`
	Fecha         string  `json:"fecha"`
	Concepto      string  `json:"concepto"`
	Categoria     string  `json:"categoria"`
	Tipo          string  `json:"tipo"`
	Valor         float64 `json:"valor"`
	MetodoPago    *string `json:"metodoPago"`
	Observaciones *string `json:"observaciones"`
}

// POST /v1/movimientos/importar
type SolicitudImportacion struct {
	IdUsuario int                   `json:"idUsuario"`
	Modo      string                `json:"modo"`
	Filas     []SolicitudMovimiento `json:"filas"`
}

// POST /v1/metas y PUT /v1/metas/:id
type SolicitudMeta struct {
	IdUsuario int     `json:"idUsuario"`
	Nombre    string  `json:"nombre"`
	Icono     string  `json:"icono"`
	Actual    float64 `json:"actual"`
	Objetivo  float64 `json:"objetivo"`
}

// PUT /v1/metas/:id/abono
type SolicitudAbono struct {
	Monto float64 `json:"monto"`
}

// POST /v1/inversiones y PUT /v1/inversiones/:id
type SolicitudInversion struct {
	IdUsuario   int     `json:"idUsuario"`
	Nombre      string  `json:"nombre"`
	Monto       float64 `json:"monto"`
	Rendimiento float64 `json:"rendimiento"`
	Riesgo      string  `json:"riesgo"`
	Duracion    string  `json:"duracion"`
}

// Filtros del Libro mayor (llegan en la URL: ?tipo=&categoria=&desde=&hasta=&q=)
type FiltrosMovimientos struct {
	Tipo      string
	Categoria string
	Desde     string
	Hasta     string
	Texto     string
}

// ============================================================
// RESULTADOS: datos ya calculados para el frontend
// ============================================================

type ResultadoCategoria struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
	Icono  string `json:"icono"`
}

type ResultadoMovimiento struct {
	Id            int     `json:"id"`
	IdUsuario     int     `json:"idUsuario"`
	Fecha         string  `json:"fecha"`
	Concepto      string  `json:"concepto"`
	Categoria     string  `json:"categoria"`
	Icono         string  `json:"icono"`
	Tipo          string  `json:"tipo"`
	Valor         float64 `json:"valor"`
	MetodoPago    string  `json:"metodoPago"`
	Observaciones string  `json:"observaciones"`
}

type ResultadoTotales struct {
	TotalIngresos         float64 `json:"totalIngresos"`
	TotalGastos           float64 `json:"totalGastos"`
	Ahorro                float64 `json:"ahorro"`
	PorcentajeAhorro      int     `json:"porcentajeAhorro"`
	PorcentajeGastado     int     `json:"porcentajeGastado"`
	GastosHormiga         float64 `json:"gastosHormiga"`
	TotalAhorros          float64 `json:"totalAhorros"`
	CapacidadAhorro       float64 `json:"capacidadAhorro"`
	ProyeccionAhorroAnual float64 `json:"proyeccionAhorroAnual"`
}

type ResultadoErrorFila struct {
	Fila     int    `json:"fila"`
	Concepto string `json:"concepto"`
	Motivo   string `json:"motivo"`
}

type ResultadoMeta struct {
	Id         int     `json:"id"`
	IdUsuario  int     `json:"idUsuario"`
	Nombre     string  `json:"nombre"`
	Icono      string  `json:"icono"`
	Actual     float64 `json:"actual"`
	Objetivo   float64 `json:"objetivo"`
	Porcentaje int     `json:"porcentaje"`
	Faltante   float64 `json:"faltante"`
	Cumplida   bool    `json:"cumplida"`
}

type ResultadoMetas struct {
	Metas           []ResultadoMeta `json:"metas"`
	TotalMetas      int             `json:"totalMetas"`
	MetasAlcanzadas int             `json:"metasAlcanzadas"`
	TotalAhorrado   float64         `json:"totalAhorrado"`
	TotalObjetivo   float64         `json:"totalObjetivo"`
	ProgresoGeneral int             `json:"progresoGeneral"`
}

type ResultadoInversion struct {
	Id                    int     `json:"id"`
	IdUsuario             int     `json:"idUsuario"`
	Nombre                string  `json:"nombre"`
	Monto                 float64 `json:"monto"`
	Rentabilidad          float64 `json:"rentabilidad"`
	Rendimiento           float64 `json:"rendimiento"`
	RendimientoPorcentaje float64 `json:"rendimientoPorcentaje"`
	Riesgo                string  `json:"riesgo"`
	Duracion              string  `json:"duracion"`
}

// Parte de un total: dona de gastos o cartera de inversiones
type ResultadoParte struct {
	Etiqueta   string  `json:"etiqueta"`
	Monto      float64 `json:"monto"`
	Porcentaje int     `json:"porcentaje"`
}

// Punto de la gráfica de crecimiento (periodo = "2026-05")
type ResultadoPunto struct {
	Periodo string  `json:"periodo"`
	Valor   float64 `json:"valor"`
}

type ResultadoAlerta struct {
	Mostrar bool   `json:"mostrar"`
	Mensaje string `json:"mensaje"`
}

type ResultadoInversiones struct {
	Inversiones           []ResultadoInversion `json:"inversiones"`
	BalanceInversiones    float64              `json:"balanceInversiones"`
	RendimientoTotal      float64              `json:"rendimientoTotal"`
	RendimientoPorcentaje float64              `json:"rendimientoPorcentaje"`
	TotalPorRiesgo        map[string]float64   `json:"totalPorRiesgo"`
	AlertaConcentracion   ResultadoAlerta      `json:"alertaConcentracion"`
	Carteras              []ResultadoParte     `json:"carteras"`
	Crecimiento           []ResultadoPunto     `json:"crecimiento"`
}

type ResultadoSalud struct {
	Estado  string `json:"estado"`
	Mensaje string `json:"mensaje"`
}

type ResultadoFilaLibro struct {
	Id        int     `json:"id"`
	Fecha     string  `json:"fecha"`
	Concepto  string  `json:"concepto"`
	Categoria string  `json:"categoria"`
	Ingreso   float64 `json:"ingreso"`
	Egreso    float64 `json:"egreso"`
}

type ResultadoLibroUsuario struct {
	IdUsuario     int                  `json:"idUsuario"`
	NombreUsuario string               `json:"nombreUsuario"`
	Filas         []ResultadoFilaLibro `json:"filas"`
	TotalIngresos float64              `json:"totalIngresos"`
	TotalEgresos  float64              `json:"totalEgresos"`
	Resultado     float64              `json:"resultado"`
}

type ResultadoLibrosAdmin struct {
	Libros                  []ResultadoLibroUsuario `json:"libros"`
	TotalIngresosPlataforma float64                 `json:"totalIngresosPlataforma"`
	TotalEgresosPlataforma  float64                 `json:"totalEgresosPlataforma"`
	UsuariosActivos         int                     `json:"usuariosActivos"`
}

// ============================================================
// RESPUESTAS: JSON que el MID envía al frontend
// Todas llevan "OK" y "Mensaje", igual que las respuestas de error.
// ============================================================

type RespuestaResumenFinanciero struct {
	OK                 bool                  `json:"OK"`
	Mensaje            string                `json:"Mensaje"`
	IdUsuario          int                   `json:"idUsuario"`
	TotalIngresos      float64               `json:"totalIngresos"`
	TotalGastos        float64               `json:"totalGastos"`
	Disponible         float64               `json:"disponible"`
	PorcentajeAhorro   int                   `json:"porcentajeAhorro"`
	Salud              ResultadoSalud        `json:"salud"`
	Distribucion       []ResultadoParte      `json:"distribucion"`
	UltimosMovimientos []ResultadoMovimiento `json:"ultimosMovimientos"`
	Metas              []ResultadoMeta       `json:"metas"`
	TotalInvertido     float64               `json:"totalInvertido"`
	GananciaEstimada   float64               `json:"gananciaEstimada"`
	Inversiones        []ResultadoInversion  `json:"inversiones"`
	Crecimiento        []ResultadoPunto      `json:"crecimiento"`
}

type RespuestaCategorias struct {
	OK         bool                 `json:"OK"`
	Mensaje    string               `json:"Mensaje"`
	Categorias []ResultadoCategoria `json:"categorias"`
}

type RespuestaMovimientos struct {
	OK          bool                  `json:"OK"`
	Mensaje     string                `json:"Mensaje"`
	Movimientos []ResultadoMovimiento `json:"movimientos"`
	Totales     ResultadoTotales      `json:"totales"`
}

type RespuestaMovimiento struct {
	OK         bool                `json:"OK"`
	Mensaje    string              `json:"Mensaje"`
	Movimiento ResultadoMovimiento `json:"movimiento"`
}

type RespuestaImportacion struct {
	OK        bool                 `json:"OK"`
	Mensaje   string               `json:"Mensaje"`
	Modo      string               `json:"modo"`
	Validos   int                  `json:"validos"`
	Guardados int                  `json:"guardados"`
	Errores   []ResultadoErrorFila `json:"errores"`
}

type RespuestaMetas struct {
	OK              bool            `json:"OK"`
	Mensaje         string          `json:"Mensaje"`
	Metas           []ResultadoMeta `json:"metas"`
	TotalMetas      int             `json:"totalMetas"`
	MetasAlcanzadas int             `json:"metasAlcanzadas"`
	TotalAhorrado   float64         `json:"totalAhorrado"`
	TotalObjetivo   float64         `json:"totalObjetivo"`
	ProgresoGeneral int             `json:"progresoGeneral"`
}

type RespuestaMeta struct {
	OK      bool          `json:"OK"`
	Mensaje string        `json:"Mensaje"`
	Meta    ResultadoMeta `json:"meta"`
}

type RespuestaInversiones struct {
	OK                    bool                 `json:"OK"`
	Mensaje               string               `json:"Mensaje"`
	Inversiones           []ResultadoInversion `json:"inversiones"`
	BalanceInversiones    float64              `json:"balanceInversiones"`
	RendimientoTotal      float64              `json:"rendimientoTotal"`
	RendimientoPorcentaje float64              `json:"rendimientoPorcentaje"`
	TotalPorRiesgo        map[string]float64   `json:"totalPorRiesgo"`
	AlertaConcentracion   ResultadoAlerta      `json:"alertaConcentracion"`
	Carteras              []ResultadoParte     `json:"carteras"`
	Crecimiento           []ResultadoPunto     `json:"crecimiento"`
}

type RespuestaInversion struct {
	OK        bool               `json:"OK"`
	Mensaje   string             `json:"Mensaje"`
	Inversion ResultadoInversion `json:"inversion"`
}

type RespuestaLibrosAdmin struct {
	OK                      bool                    `json:"OK"`
	Mensaje                 string                  `json:"Mensaje"`
	Libros                  []ResultadoLibroUsuario `json:"libros"`
	TotalIngresosPlataforma float64                 `json:"totalIngresosPlataforma"`
	TotalEgresosPlataforma  float64                 `json:"totalEgresosPlataforma"`
	UsuariosActivos         int                     `json:"usuariosActivos"`
}
