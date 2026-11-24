package models

import "time"

// ------------------------------------------------------------
// Tablas de crud_negocio (como llegan del CRUD)
// ------------------------------------------------------------

type Banco struct {
	Id          int    `json:"Id"`
	NombreBanco string `json:"NombreBanco"`
	TipoAliado  string `json:"TipoAliado"`
	UrlLogo     string `json:"UrlLogo"`
	LogoTexto   string `json:"LogoTexto"`
	LogoFondo   string `json:"LogoFondo"`
}

type ProductoCrediticio struct {
	Id               int     `json:"Id"`
	IdBanco          *Ref    `json:"IdBanco"`
	Slug             string  `json:"Slug"`
	NombreProducto   string  `json:"NombreProducto"`
	Familia          string  `json:"Familia"`
	Beneficio        string  `json:"Beneficio"`
	Descripcion      string  `json:"Descripcion"`
	Promesa          string  `json:"Promesa"`
	Vigencia         string  `json:"Vigencia"`
	MontoMinimo      float64 `json:"MontoMinimo"`
	MontoMaximo      float64 `json:"MontoMaximo"`
	TasaMinima       float64 `json:"TasaMinima"`
	TasaMaxima       float64 `json:"TasaMaxima"`
	PlazoMinimo      int     `json:"PlazoMinimo"`
	PlazoMaximo      int     `json:"PlazoMaximo"`
	TiempoAprobacion string  `json:"TiempoAprobacion"`
	Calificacion     float64 `json:"Calificacion"`
	Usuarios         int     `json:"Usuarios"`
	Compatibilidad   int     `json:"Compatibilidad"`
	Destacada        bool    `json:"Destacada"`
	CtaPrimaria      string  `json:"CtaPrimaria"`
	CtaSecundaria    string  `json:"CtaSecundaria"`
	Activo           bool    `json:"Activo"`
}

type ProductoDetalle struct {
	Id         int    `json:"Id"`
	IdProducto *Ref   `json:"IdProducto"`
	Tipo       string `json:"Tipo"`
	Texto      string `json:"Texto"`
}

type ProductoTarifa struct {
	Concepto string `json:"Concepto"`
	Valor    string `json:"Valor"`
}

type ProductoPregunta struct {
	Pregunta  string `json:"Pregunta"`
	Respuesta string `json:"Respuesta"`
}

type AsesorBancario struct {
	Id       int    `json:"Id"`
	Nombre   string `json:"Nombre"`
	Apellido string `json:"Apellido"`
}

type Lead struct {
	Id                   int       `json:"Id"`
	IdUsuario            int       `json:"IdUsuario"`
	IdProducto           *Ref      `json:"IdProducto"`
	IdAsesor             *Ref      `json:"IdAsesor"`
	Radicado             string    `json:"Radicado"`
	TipoCredito          string    `json:"TipoCredito"`
	MontoInteres         float64   `json:"MontoInteres"`
	PlazoInteres         int       `json:"PlazoInteres"`
	NombreSolicitante    string    `json:"NombreSolicitante"`
	DocumentoSolicitante string    `json:"DocumentoSolicitante"`
	CorreoSolicitante    string    `json:"CorreoSolicitante"`
	CelularSolicitante   string    `json:"CelularSolicitante"`
	IngresosMensuales    float64   `json:"IngresosMensuales"`
	AceptaTerminos       bool      `json:"AceptaTerminos"`
	EstadoLead           string    `json:"EstadoLead"`
	FechaGeneracion      time.Time `json:"FechaGeneracion"`
	Activo               bool      `json:"Activo"`
}

// ------------------------------------------------------------
// Datos que manda la pagina
// ------------------------------------------------------------

type DatosSolicitudCredito struct {
	IdUsuario      int     `json:"id_usuario"`
	IdProducto     int     `json:"id_producto"`
	Monto          float64 `json:"monto"`
	PlazoMeses     int     `json:"plazo_meses"`
	Nombre         string  `json:"nombre"`
	Documento      string  `json:"documento"`
	Correo         string  `json:"correo"`
	Celular        string  `json:"celular"`
	Ingresos       float64 `json:"ingresos"`
	AceptaTerminos bool    `json:"acepta_terminos"`
}
