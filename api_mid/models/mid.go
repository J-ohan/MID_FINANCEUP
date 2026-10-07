package models

import "time"

// Estructuras que el MID recibe del frontend y las que le devuelve.
// Usan llaves en minuscula con guion bajo (snake_case), que es lo mas
// comodo para el frontend.

// ---------------- Usuarios ----------------

type DatosRegistro struct {
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	Email           string `json:"email"`
	Contrasena      string `json:"contrasena"`
	IdTipoDocumento int    `json:"id_tipo_documento"`
	Cedula          string `json:"cedula"`
	Telefono        string `json:"telefono"`
	Ciudad          string `json:"ciudad"`
	Direccion       string `json:"direccion"`
	FechaNacimiento string `json:"fecha_nacimiento"` // formato AAAA-MM-DD
}

type DatosLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

type Perfil struct {
	IdUsuario       int        `json:"id_usuario"`
	Nombre          string     `json:"nombre"`
	Apellido        string     `json:"apellido"`
	Email           string     `json:"email"`
	TipoDocumento   string     `json:"tipo_documento"`
	Cedula          string     `json:"cedula"`
	Telefono        string     `json:"telefono"`
	Ciudad          string     `json:"ciudad"`
	Direccion       string     `json:"direccion"`
	FechaNacimiento string     `json:"fecha_nacimiento"`
	Estado          string     `json:"estado"`
	FechaRegistro   time.Time  `json:"fecha_registro"`
	UltimaSesion    *time.Time `json:"ultima_sesion,omitempty"`
	Roles           []string   `json:"roles"`
}

// ---------------- Finanzas ----------------

type ResumenFinanciero struct {
	IdUsuario          int                `json:"id_usuario"`
	TotalIngresos      float64            `json:"total_ingresos"`
	TotalGastos        float64            `json:"total_gastos"`
	Balance            float64            `json:"balance"`
	GastosPorCategoria []GastoCategoria   `json:"gastos_por_categoria"`
	UltimosMovimientos []MovimientoVista  `json:"ultimos_movimientos"`
	Metas              []MetaVista        `json:"metas"`
	Inversiones        ResumenInversiones `json:"inversiones"`
	SolicitudesDeuda   []SolicitudDeuda   `json:"solicitudes_deuda"`
}

type GastoCategoria struct {
	Categoria string  `json:"categoria"`
	Total     float64 `json:"total"`
}

type MovimientoVista struct {
	Id            int     `json:"id"`
	Fecha         string  `json:"fecha"`
	Concepto      string  `json:"concepto"`
	Categoria     string  `json:"categoria"`
	Tipo          string  `json:"tipo"` // "ingreso" o "gasto"
	Valor         float64 `json:"valor"`
	MetodoPago    string  `json:"metodo_pago"`
	Observaciones string  `json:"observaciones"`
}

type MetaVista struct {
	Id          int     `json:"id"`
	Nombre      string  `json:"nombre"`
	Icono       string  `json:"icono"`
	Color       string  `json:"color"`
	Actual      float64 `json:"actual"`
	Objetivo    float64 `json:"objetivo"`
	Porcentaje  float64 `json:"porcentaje"`
	Cumplida    bool    `json:"cumplida"`
	FechaLimite string  `json:"fecha_limite"`
}

type ResumenInversiones struct {
	TotalInvertido   float64          `json:"total_invertido"`
	GananciaEstimada float64          `json:"ganancia_estimada"`
	Cantidad         int              `json:"cantidad"`
	Detalle          []InversionVista `json:"detalle"`
}

type InversionVista struct {
	Id           int     `json:"id"`
	Nombre       string  `json:"nombre"`
	Tipo         string  `json:"tipo"`
	Riesgo       string  `json:"riesgo"`
	Monto        float64 `json:"monto"`
	Rentabilidad float64 `json:"rentabilidad"`
	FechaInicio  string  `json:"fecha_inicio"`
	FechaFin     string  `json:"fecha_fin"`
}

type SolicitudDeuda struct {
	Id             int     `json:"id"`
	SaldoTotal     float64 `json:"saldo_total"`
	CuotaActual    float64 `json:"cuota_actual"`
	CuotaPropuesta float64 `json:"cuota_propuesta"`
	AhorroMensual  float64 `json:"ahorro_mensual"`
	Estado         string  `json:"estado"`
}

// ---------------- Educacion ----------------

type ProgresoGeneral struct {
	IdUsuario          int           `json:"id_usuario"`
	ModulosTotales     int           `json:"modulos_totales"`
	ModulosCompletados int           `json:"modulos_completados"`
	PorcentajeGeneral  float64       `json:"porcentaje_general"`
	Modulos            []ModuloVista `json:"modulos"`
}

type ModuloVista struct {
	IdModulo             int    `json:"id_modulo"`
	Titulo               string `json:"titulo"`
	Descripcion          string `json:"descripcion"`
	Nivel                string `json:"nivel"`
	Imagen               string `json:"imagen"`
	Porcentaje           int    `json:"porcentaje"`
	Estado               string `json:"estado"` // "sin iniciar", "en curso" o "completado"
	LeccionesTotales     int    `json:"lecciones_totales"`
	LeccionesCompletadas int    `json:"lecciones_completadas"`
	Calificacion         *int   `json:"calificacion,omitempty"`
}

// ---------------- Negocio (creditos) ----------------

type Oferta struct {
	IdProducto  int      `json:"id_producto"`
	Producto    string   `json:"producto"`
	Descripcion string   `json:"descripcion"`
	Banco       string   `json:"banco"`
	LogoBanco   string   `json:"logo_banco"`
	SitioWeb    string   `json:"sitio_web"`
	MontoMinimo float64  `json:"monto_minimo"`
	MontoMaximo float64  `json:"monto_maximo"`
	TasaMinima  float64  `json:"tasa_minima"`
	TasaMaxima  float64  `json:"tasa_maxima"`
	PlazoMinimo int      `json:"plazo_minimo"`
	PlazoMaximo int      `json:"plazo_maximo"`
	Requisitos  []string `json:"requisitos"`
}

type DatosSolicitudCredito struct {
	IdUsuario     int     `json:"id_usuario"`
	IdProducto    int     `json:"id_producto"`
	Monto         float64 `json:"monto"`
	PlazoMeses    int     `json:"plazo_meses"`
	Observaciones string  `json:"observaciones"`
}

type SolicitudCredito struct {
	IdSolicitud    int       `json:"id_solicitud"`
	Producto       string    `json:"producto"`
	Banco          string    `json:"banco"`
	Monto          float64   `json:"monto"`
	PlazoMeses     int       `json:"plazo_meses"`
	CuotaEstimada  float64   `json:"cuota_estimada"`
	Estado         string    `json:"estado"`
	Asesor         string    `json:"asesor"`
	FechaSolicitud time.Time `json:"fecha_solicitud"`
}

// ---------------- Soporte (PQR) ----------------

type DatosPqr struct {
	IdUsuario   int    `json:"id_usuario"`
	Titulo      string `json:"titulo"`
	Tipo        string `json:"tipo"` // peticion, queja, reclamo o sugerencia
	Categoria   string `json:"categoria"`
	Prioridad   string `json:"prioridad"` // alta, media o baja
	Descripcion string `json:"descripcion"`
}

type PqrVista struct {
	IdPqr               int            `json:"id_pqr"`
	Radicado            string         `json:"radicado"`
	Titulo              string         `json:"titulo"`
	Tipo                string         `json:"tipo"`
	Categoria           string         `json:"categoria"`
	Prioridad           string         `json:"prioridad"`
	Estado              string         `json:"estado"`
	Descripcion         string         `json:"descripcion"`
	Asesor              string         `json:"asesor"`
	Respuesta           string         `json:"respuesta"`
	FechaCreacion       time.Time      `json:"fecha_creacion"`
	UltimaActualizacion time.Time      `json:"ultima_actualizacion"`
	Adjuntos            []AdjuntoVista `json:"adjuntos"`
}

type AdjuntoVista struct {
	Nombre      string `json:"nombre"`
	Ruta        string `json:"ruta"`
	Tipo        string `json:"tipo"`
	TamanoBytes int    `json:"tamano_bytes"`
}

// ---------------- Panel general ----------------

type Dashboard struct {
	Perfil             Perfil             `json:"perfil"`
	Finanzas           *ResumenFinanciero `json:"finanzas"`
	Educacion          *ProgresoGeneral   `json:"educacion"`
	SolicitudesCredito []SolicitudCredito `json:"solicitudes_credito"`
	Pqr                []PqrVista         `json:"pqr"`
	// Avisos lista las secciones que no se pudieron cargar (por ejemplo
	// porque un CRUD esta apagado). El resto del panel igual se entrega.
	Avisos []string `json:"avisos,omitempty"`
}
