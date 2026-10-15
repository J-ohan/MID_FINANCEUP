package models

import "time"

// ------------------------------------------------------------
// Tablas de crud_auth (como llegan del CRUD)
// ------------------------------------------------------------

type Usuario struct {
	Id                      int       `json:"Id"`
	TipoDocumento           *Ref      `json:"TipoDocumento"`
	Nombre                  string    `json:"Nombre"`
	Apellido                string    `json:"Apellido"`
	Email                   string    `json:"Email"`
	Telefono                string    `json:"Telefono"`
	Cedula                  *string   `json:"Cedula"` // nil = sin cedula (se guarda NULL)
	Ciudad                  string    `json:"Ciudad"`
	Direccion               string    `json:"Direccion"`
	FechaNacimiento         time.Time `json:"FechaNacimiento"`
	AceptaTerminos          bool      `json:"AceptaTerminos"`
	FechaAceptacionTerminos time.Time `json:"FechaAceptacionTerminos"`
	Estado                  string    `json:"Estado"`
	FechaRegistro           time.Time `json:"FechaRegistro"`
	FechaUltimaSesion       time.Time `json:"FechaUltimaSesion"`
	Activo                  bool      `json:"Activo"`
	FechaCreacion           time.Time `json:"FechaCreacion"`
	FechaModificacion       time.Time `json:"FechaModificacion"`
}

type Credencial struct {
	Id                 int       `json:"Id"`
	IdUsuario          *Ref      `json:"IdUsuario"`
	ContrasenaHash     string    `json:"ContrasenaHash"`
	Salt               string    `json:"Salt"`
	Algoritmo          string    `json:"Algoritmo"`
	FechaActualizacion time.Time `json:"FechaActualizacion"`
	FechaUltimoCambio  time.Time `json:"FechaUltimoCambio"`
	IntentosFallidos   int       `json:"IntentosFallidos"`
	BloqueadoHasta     time.Time `json:"BloqueadoHasta"`
	RequiereCambio     bool      `json:"RequiereCambio"`
	Activo             bool      `json:"Activo"`
	FechaCreacion      time.Time `json:"FechaCreacion"`
	FechaModificacion  time.Time `json:"FechaModificacion"`
}

type CodigoVerificacion struct {
	Id                int       `json:"Id"`
	IdUsuario         *Ref      `json:"IdUsuario"`
	Proposito         string    `json:"Proposito"`
	Canal             string    `json:"Canal"`
	CodigoHash        string    `json:"CodigoHash"`
	FechaExpiracion   time.Time `json:"FechaExpiracion"`
	FechaUso          time.Time `json:"FechaUso"`
	Intentos          int       `json:"Intentos"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type CuentaSocial struct {
	Id             int    `json:"Id"`
	IdUsuario      *Ref   `json:"IdUsuario"`
	Proveedor      string `json:"Proveedor"`
	IdExterno      string `json:"IdExterno"`
	EmailProveedor string `json:"EmailProveedor"`
	Activo         bool   `json:"Activo"`
}

type Rol struct {
	Id        int    `json:"Id"`
	NombreRol string `json:"NombreRol"`
}

type UsuarioRol struct {
	Id        int  `json:"Id"`
	IdUsuario *Ref `json:"IdUsuario"`
	IdRol     *Ref `json:"IdRol"`
	Activo    bool `json:"Activo"`
}

type TipoDocumento struct {
	Id     int    `json:"Id"`
	Codigo string `json:"Codigo"`
}

type AuditoriaLogin struct {
	IdUsuario    *Ref   `json:"IdUsuario"`
	TipoEvento   string `json:"TipoEvento"`
	IpAddress    string `json:"IpAddress"`
	Navegador    string `json:"Navegador"`
	EstadoEvento string `json:"EstadoEvento"`
}

// ------------------------------------------------------------
// Datos que manda la pagina
// ------------------------------------------------------------

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
	AceptaTerminos  bool   `json:"acepta_terminos"`
}

type DatosLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
	Ip         string `json:"-"`
	Navegador  string `json:"-"`
}

type DatosCodigo struct {
	IdUsuario int    `json:"id_usuario"`
	Codigo    string `json:"codigo"`
}

type DatosLoginSocial struct {
	Proveedor string `json:"proveedor"`
	IdExterno string `json:"id_externo"`
	Email     string `json:"email"`
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
}

type DatosContrasena struct {
	Email            string `json:"email"`
	Codigo           string `json:"codigo"`
	ContrasenaActual string `json:"contrasena_actual"`
	ContrasenaNueva  string `json:"contrasena_nueva"`
}
