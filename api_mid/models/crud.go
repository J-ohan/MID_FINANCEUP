package models

import "time"

// Estructuras con la misma forma del JSON que devuelven los CRUD.
// Los CRUD generados por bee usan los nombres de Go como llaves ("Id", "IdUsuario"...).
// Solo se copian los campos que el MID necesita, salvo en las tablas que el MID
// actualiza con PUT (Usuario y Credencial), donde van todos los campos porque
// el PUT reemplaza el registro completo.

// Ref es una llave foranea tal como la manejan los CRUD: {"Id": 1}
type Ref struct {
	Id int
}

// ---------------- crud_auth ----------------

type Usuario struct {
	Id                int
	TipoDocumento     *Ref
	Nombre            string
	Apellido          string
	Email             string
	Telefono          string
	Cedula            string
	Ciudad            string
	Direccion         string
	FechaNacimiento   time.Time
	Estado            string
	FechaRegistro     time.Time
	FechaUltimaSesion time.Time
	Activo            bool
	FechaCreacion     time.Time
	FechaModificacion time.Time
}

type Credencial struct {
	Id                 int
	IdUsuario          *Ref
	ContrasenaHash     string
	Salt               string
	Algoritmo          string
	FechaActualizacion time.Time
	FechaUltimoCambio  time.Time
	IntentosFallidos   int
	BloqueadoHasta     time.Time
	RequiereCambio     bool
	Activo             bool
	FechaCreacion      time.Time
	FechaModificacion  time.Time
}

type Rol struct {
	Id        int
	NombreRol string
	Activo    bool
}

type UsuarioRol struct {
	Id        int
	IdUsuario *Ref
	IdRol     *Ref
	Activo    bool
}

type TipoDocumento struct {
	Id     int
	Nombre string
	Codigo string
}

type AuditoriaLogin struct {
	Id           int
	IdUsuario    *Ref
	TipoEvento   string
	IpAddress    string
	Navegador    string
	EstadoEvento string
}

// ---------------- crud_finanzas ----------------

type MovimientoIngresoEgreso struct {
	Id            int
	IdUsuario     int
	IdCategoria   *Ref
	Nombre        string
	Monto         float64
	EsIngreso     bool
	Fecha         time.Time
	MetodoPago    string
	Observaciones string
}

type Categoria struct {
	Id     int
	Nombre string
}

type Meta struct {
	Id            int
	IdUsuario     int
	Nombre        string
	Descripcion   string
	MontoObjetivo float64
	MontoActual   float64
	FechaLimite   time.Time
	Color         string
	Icono         string
}

type Inversion struct {
	Id              int
	IdUsuario       int
	IdTipoInversion *Ref
	IdNivelRiesgo   *Ref
	Nombre          string
	Monto           float64
	Rentabilidad    float64
	FechaInicio     time.Time
	FechaFin        time.Time
}

type NivelRiesgo struct {
	Id     int
	Nombre string
}

type TipoInversion struct {
	Id     int
	Nombre string
}

type SolicitudConsolidacion struct {
	Id             int
	IdUsuario      int
	SaldoTotal     float64
	CuotaActual    float64
	CuotaPropuesta float64
	Estado         string
	FechaCreacion  time.Time
}

// ---------------- crud_educacion ----------------

type ModuloEducativo struct {
	Id           int
	Titulo       string
	Descripcion  string
	Nivel        string
	UrlThumbnail string
}

type Leccion struct {
	Id       int
	IdModulo *Ref
}

type ProgresoEducativo struct {
	Id                   int
	IdUsuario            int
	IdModulo             *Ref
	PorcentajeCompletado int
	FechaInicio          time.Time
	FechaCompletado      time.Time
	Calificacion         int
}

type ProgresoLeccion struct {
	Id         int
	IdUsuario  int
	IdLeccion  *Ref
	Completado bool
}

// ---------------- crud_negocio ----------------

type Banco struct {
	Id          int
	NombreBanco string
	UrlLogo     string
	SitioWeb    string
	Telefono    string
	Email       string
}

type ProductoCrediticio struct {
	Id             int
	IdBanco        *Ref
	NombreProducto string
	Descripcion    string
	MontoMinimo    float64
	MontoMaximo    float64
	TasaMinima     float64
	TasaMaxima     float64
	PlazoMinimo    int
	PlazoMaximo    int
	Requisitos     string
	Activo         bool
}

type AsesorBancario struct {
	Id       int
	IdBanco  *Ref
	Nombre   string
	Apellido string
	Email    string
	Telefono string
	Activo   bool
}

type Lead struct {
	Id              int
	IdUsuario       int
	IdProducto      *Ref
	IdAsesor        *Ref
	TipoCredito     string
	MontoInteres    float64
	PlazoInteres    int
	EstadoLead      string
	FechaGeneracion time.Time
	Observaciones   string
	Activo          bool
}

// ---------------- crud_soporte ----------------

type Pqr struct {
	Id                int
	IdUsuario         int
	Radicado          string
	Titulo            string
	Tipo              string
	Categoria         string
	Prioridad         string
	Asesor            string
	MensajeRespuesta  string
	Descripcion       string
	IdEstado          *Ref
	Activo            bool
	FechaCreacion     time.Time
	FechaModificacion time.Time
}

type EstadoPqr struct {
	Id     int
	Nombre string
}

type Adjunto struct {
	Id            int
	IdPqr         *Ref
	NombreArchivo string
	RutaArchivo   string
	TipoMime      string
	TamanoBytes   int
}

type RegistroActividad struct {
	Id              int
	IdUsuario       *int
	TipoActividad   string
	Descripcion     string
	EntidadAfectada string
}
