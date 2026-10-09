package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

// ------------------------------------------------------------
// Direcciones de los CRUD
// ------------------------------------------------------------

var urlCrud = map[string]string{
	"auth":      "http://localhost:8081/v1",
	"finanzas":  "http://localhost:8082/v1",
	"educacion": "http://localhost:8083/v1",
	"negocio":   "http://localhost:8084/v1",
	"soporte":   "http://localhost:8085/v1",
}

// ------------------------------------------------------------
// Estructuras que llegan de los CRUD
// (los CRUD de bee usan los nombres de Go: "Id", "IdUsuario"...)
// ------------------------------------------------------------

// Ref es una llave foranea tal como llega de los CRUD: {"Id": 1}
type Ref struct {
	Id int
}

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
}

type UsuarioRol struct {
	Id        int
	IdUsuario *Ref
	IdRol     *Ref
	Activo    bool
}

type TipoDocumento struct {
	Id     int
	Codigo string
}

type AuditoriaLogin struct {
	IdUsuario    *Ref
	TipoEvento   string
	IpAddress    string
	Navegador    string
	EstadoEvento string
}

type Movimiento struct {
	Id            int
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
	Nombre        string
	MontoObjetivo float64
	MontoActual   float64
	FechaLimite   time.Time
	Color         string
	Icono         string
}

type Inversion struct {
	Id              int
	IdTipoInversion *Ref
	IdNivelRiesgo   *Ref
	Nombre          string
	Monto           float64
	Rentabilidad    float64
	FechaInicio     time.Time
	FechaFin        time.Time
}

type Catalogo struct {
	Id     int
	Nombre string
}

type SolicitudConsolidacion struct {
	Id             int
	SaldoTotal     float64
	CuotaActual    float64
	CuotaPropuesta float64
	Estado         string
}

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
	IdModulo             *Ref
	PorcentajeCompletado int
	Calificacion         int
}

type ProgresoLeccion struct {
	IdLeccion *Ref
}

type Banco struct {
	Id          int
	NombreBanco string
	UrlLogo     string
	SitioWeb    string
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
	Nombre   string
	Apellido string
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

type Adjunto struct {
	NombreArchivo string
	RutaArchivo   string
	TipoMime      string
	TamanoBytes   int
}

type RegistroActividad struct {
	IdUsuario       *int
	TipoActividad   string
	Descripcion     string
	EntidadAfectada string
}

// ------------------------------------------------------------
// Estructuras que recibe y devuelve el MID
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
	FechaNacimiento string `json:"fecha_nacimiento"`
}

type DatosLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

type DatosSolicitudCredito struct {
	IdUsuario     int     `json:"id_usuario"`
	IdProducto    int     `json:"id_producto"`
	Monto         float64 `json:"monto"`
	PlazoMeses    int     `json:"plazo_meses"`
	Observaciones string  `json:"observaciones"`
}

type DatosPqr struct {
	IdUsuario   int    `json:"id_usuario"`
	Titulo      string `json:"titulo"`
	Tipo        string `json:"tipo"`
	Categoria   string `json:"categoria"`
	Prioridad   string `json:"prioridad"`
	Descripcion string `json:"descripcion"`
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

type ResumenFinanciero struct {
	IdUsuario          int                      `json:"id_usuario"`
	TotalIngresos      float64                  `json:"total_ingresos"`
	TotalGastos        float64                  `json:"total_gastos"`
	Balance            float64                  `json:"balance"`
	GastosPorCategoria []map[string]interface{} `json:"gastos_por_categoria"`
	UltimosMovimientos []map[string]interface{} `json:"ultimos_movimientos"`
	Metas              []map[string]interface{} `json:"metas"`
	TotalInvertido     float64                  `json:"total_invertido"`
	GananciaEstimada   float64                  `json:"ganancia_estimada"`
	Inversiones        []map[string]interface{} `json:"inversiones"`
	SolicitudesDeuda   []map[string]interface{} `json:"solicitudes_deuda"`
}

type ProgresoGeneral struct {
	IdUsuario          int                      `json:"id_usuario"`
	ModulosTotales     int                      `json:"modulos_totales"`
	ModulosCompletados int                      `json:"modulos_completados"`
	PorcentajeGeneral  float64                  `json:"porcentaje_general"`
	Modulos            []map[string]interface{} `json:"modulos"`
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

type PqrCompleta struct {
	IdPqr               int                      `json:"id_pqr"`
	Radicado            string                   `json:"radicado"`
	Titulo              string                   `json:"titulo"`
	Tipo                string                   `json:"tipo"`
	Categoria           string                   `json:"categoria"`
	Prioridad           string                   `json:"prioridad"`
	Estado              string                   `json:"estado"`
	Descripcion         string                   `json:"descripcion"`
	Asesor              string                   `json:"asesor"`
	Respuesta           string                   `json:"respuesta"`
	FechaCreacion       time.Time                `json:"fecha_creacion"`
	UltimaActualizacion time.Time                `json:"ultima_actualizacion"`
	Adjuntos            []map[string]interface{} `json:"adjuntos"`
}

// MidController operations for Mid
type MidController struct {
	beego.Controller
}

// ============================================================
// GET /v1/estado
// Dice que CRUD estan encendidos.
// ============================================================

func (c *MidController) Estado() {

	estado := map[string]string{"mid": "encendido"}

	for conjunto, base := range urlCrud {

		respuesta, err := http.Get(base)

		if err != nil {
			estado["crud_"+conjunto] = "apagado"
			continue
		}

		respuesta.Body.Close()
		estado["crud_"+conjunto] = "encendido"
	}

	c.Data["json"] = estado
	c.ServeJSON()
}

// ============================================================
// GET /v1/perfil/:id
// Usuario + tipo de documento + roles (crud_auth).
// ============================================================

func (c *MidController) GetPerfil() {

	// 1. Obtener el ID enviado en la URL
	id, ok := c.leerId()
	if !ok {
		return
	}

	// 2. Armar el perfil consultando crud_auth
	perfil, err := obtenerPerfil(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Devolver el JSON
	c.Data["json"] = perfil
	c.ServeJSON()
}

// ============================================================
// POST /v1/registro
// Crea usuario, clave cifrada y rol "Usuario" (crud_auth).
// ============================================================

func (c *MidController) PostRegistro() {

	// --------------------------------------------------------
	// 1. Leer y validar los datos enviados
	// --------------------------------------------------------

	var datos DatosRegistro

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.responderError(nuevoError(http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido"))
		return
	}

	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))

	switch {
	case strings.TrimSpace(datos.Nombre) == "":
		c.responderError(nuevoError(http.StatusBadRequest, "El nombre es obligatorio"))
		return
	case strings.TrimSpace(datos.Apellido) == "":
		c.responderError(nuevoError(http.StatusBadRequest, "El apellido es obligatorio"))
		return
	case !strings.Contains(datos.Email, "@") || !strings.Contains(datos.Email, "."):
		c.responderError(nuevoError(http.StatusBadRequest, "El correo no es valido"))
		return
	case len(datos.Contrasena) < 8:
		c.responderError(nuevoError(http.StatusBadRequest, "La contrasena debe tener minimo 8 caracteres"))
		return
	}

	var fechaNacimiento time.Time
	if datos.FechaNacimiento != "" {
		fecha, err := time.Parse("2006-01-02", datos.FechaNacimiento)
		if err != nil {
			c.responderError(nuevoError(http.StatusBadRequest, "La fecha de nacimiento debe tener el formato AAAA-MM-DD"))
			return
		}
		fechaNacimiento = fecha
	}

	// --------------------------------------------------------
	// 2. Revisar que el correo y la cedula no esten repetidos
	// --------------------------------------------------------

	var repetidos []Usuario

	if err := consultarCrud(urlCrud["auth"]+"/usuario?"+filtro("email", datos.Email), &repetidos); err != nil {
		c.responderError(err)
		return
	}

	if len(repetidos) > 0 {
		c.responderError(nuevoError(http.StatusConflict, "Ya existe un usuario con ese correo"))
		return
	}

	if datos.Cedula != "" {

		if err := consultarCrud(urlCrud["auth"]+"/usuario?"+filtro("cedula", datos.Cedula), &repetidos); err != nil {
			c.responderError(err)
			return
		}

		if len(repetidos) > 0 {
			c.responderError(nuevoError(http.StatusConflict, "Ya existe un usuario con ese numero de documento"))
			return
		}
	}

	// --------------------------------------------------------
	// 3. Buscar el rol "Usuario"
	// --------------------------------------------------------

	var roles []Rol

	if err := consultarCrud(urlCrud["auth"]+"/rol?"+filtro("nombre_rol", "Usuario"), &roles); err != nil {
		c.responderError(err)
		return
	}

	if len(roles) == 0 {
		c.responderError(nuevoError(http.StatusInternalServerError, "No existe el rol 'Usuario' en la base de datos"))
		return
	}

	// --------------------------------------------------------
	// 4. Cifrar la contrasena (nunca se guarda en texto plano)
	// --------------------------------------------------------

	hash, err := bcrypt.GenerateFromPassword([]byte(datos.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		c.responderError(nuevoError(http.StatusInternalServerError, "No fue posible proteger la contrasena"))
		return
	}

	// --------------------------------------------------------
	// 5. Crear el usuario
	// --------------------------------------------------------

	usuario := Usuario{
		Nombre:          strings.TrimSpace(datos.Nombre),
		Apellido:        strings.TrimSpace(datos.Apellido),
		Email:           datos.Email,
		Cedula:          datos.Cedula,
		Telefono:        datos.Telefono,
		Ciudad:          datos.Ciudad,
		Direccion:       datos.Direccion,
		FechaNacimiento: fechaNacimiento,
		Estado:          "activo",
		Activo:          true,
	}

	if datos.IdTipoDocumento > 0 {
		usuario.TipoDocumento = &Ref{Id: datos.IdTipoDocumento}
	}

	if err := enviarCrud("POST", urlCrud["auth"]+"/usuario", usuario, &usuario); err != nil {
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 6. Guardar la clave cifrada
	//    Si falla, se borra el usuario para no dejarlo a medias.
	// --------------------------------------------------------

	credencial := Credencial{
		IdUsuario:      &Ref{Id: usuario.Id},
		ContrasenaHash: string(hash),
		Salt:           string(hash[7:29]),
		Algoritmo:      "bcrypt",
		Activo:         true,
	}

	if err := enviarCrud("POST", urlCrud["auth"]+"/credencial", credencial, &credencial); err != nil {
		enviarCrud("DELETE", fmt.Sprintf("%s/usuario/%d", urlCrud["auth"], usuario.Id), nil, nil)
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 7. Asignar el rol
	// --------------------------------------------------------

	asignacion := UsuarioRol{
		IdUsuario: &Ref{Id: usuario.Id},
		IdRol:     &Ref{Id: roles[0].Id},
		Activo:    true,
	}

	if err := enviarCrud("POST", urlCrud["auth"]+"/usuario_rol", asignacion, nil); err != nil {
		enviarCrud("DELETE", fmt.Sprintf("%s/credencial/%d", urlCrud["auth"], credencial.Id), nil, nil)
		enviarCrud("DELETE", fmt.Sprintf("%s/usuario/%d", urlCrud["auth"], usuario.Id), nil, nil)
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 8. Devolver el perfil del usuario creado
	// --------------------------------------------------------

	perfil, err := obtenerPerfil(usuario.Id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = perfil
	c.ServeJSON()
}

// ============================================================
// POST /v1/login
// Valida correo y clave. Bloquea 15 minutos tras 5 intentos.
// ============================================================

func (c *MidController) PostLogin() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosLogin

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.responderError(nuevoError(http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido"))
		return
	}

	email := strings.ToLower(strings.TrimSpace(datos.Email))

	if email == "" || datos.Contrasena == "" {
		c.responderError(nuevoError(http.StatusBadRequest, "El correo y la contrasena son obligatorios"))
		return
	}

	errorCredenciales := nuevoError(http.StatusUnauthorized, "Correo o contrasena incorrectos")

	// --------------------------------------------------------
	// 2. Buscar el usuario por correo
	// --------------------------------------------------------

	var usuarios []Usuario

	if err := consultarCrud(urlCrud["auth"]+"/usuario?"+filtro("email", email), &usuarios); err != nil {
		c.responderError(err)
		return
	}

	if len(usuarios) == 0 {
		c.responderError(errorCredenciales)
		return
	}

	usuario := usuarios[0]

	if !usuario.Activo || usuario.Estado != "activo" {
		c.responderError(nuevoError(http.StatusForbidden, "La cuenta esta "+usuario.Estado+". Comunicate con soporte."))
		return
	}

	// --------------------------------------------------------
	// 3. Buscar su credencial y revisar si esta bloqueada
	// --------------------------------------------------------

	var credenciales []Credencial

	if err := consultarCrud(urlCrud["auth"]+"/credencial?"+filtro("id_usuario", usuario.Id), &credenciales); err != nil {
		c.responderError(err)
		return
	}

	if len(credenciales) == 0 {
		c.responderError(errorCredenciales)
		return
	}

	credencial := credenciales[0]
	ahora := time.Now().UTC()

	if credencial.BloqueadoHasta.After(ahora) {
		minutos := int(credencial.BloqueadoHasta.Sub(ahora).Minutes()) + 1
		c.responderError(nuevoError(http.StatusLocked, fmt.Sprintf("Cuenta bloqueada por intentos fallidos. Intenta de nuevo en %d minutos.", minutos)))
		return
	}

	auditoria := AuditoriaLogin{
		IdUsuario:  &Ref{Id: usuario.Id},
		TipoEvento: "login",
		IpAddress:  c.Ctx.Input.IP(),
		Navegador:  recortar(c.Ctx.Input.UserAgent(), 200),
	}

	urlCredencial := fmt.Sprintf("%s/credencial/%d", urlCrud["auth"], credencial.Id)

	// --------------------------------------------------------
	// 4. Comparar la clave. Si esta mal, sumar un intento.
	// --------------------------------------------------------

	if bcrypt.CompareHashAndPassword([]byte(credencial.ContrasenaHash), []byte(datos.Contrasena)) != nil {

		credencial.IntentosFallidos++

		if credencial.IntentosFallidos >= 5 {
			credencial.IntentosFallidos = 0
			credencial.BloqueadoHasta = ahora.Add(15 * time.Minute)
		}

		enviarCrud("PUT", urlCredencial, credencial, nil)

		auditoria.EstadoEvento = "fallido"
		enviarCrud("POST", urlCrud["auth"]+"/auditoria_login", auditoria, nil)

		c.responderError(errorCredenciales)
		return
	}

	// --------------------------------------------------------
	// 5. Clave correcta: reiniciar intentos y guardar la sesion
	// --------------------------------------------------------

	if credencial.IntentosFallidos > 0 {
		credencial.IntentosFallidos = 0
		enviarCrud("PUT", urlCredencial, credencial, nil)
	}

	usuario.FechaUltimaSesion = ahora
	enviarCrud("PUT", fmt.Sprintf("%s/usuario/%d", urlCrud["auth"], usuario.Id), usuario, nil)

	auditoria.EstadoEvento = "exitoso"
	enviarCrud("POST", urlCrud["auth"]+"/auditoria_login", auditoria, nil)

	// --------------------------------------------------------
	// 6. Devolver el perfil
	// --------------------------------------------------------

	perfil, err := obtenerPerfil(usuario.Id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Data["json"] = perfil
	c.ServeJSON()
}

// ============================================================
// GET /v1/resumen-financiero/:id
// Ingresos, gastos, balance, metas, inversiones y deudas.
// ============================================================

func (c *MidController) GetResumenFinanciero() {

	id, ok := c.leerId()
	if !ok {
		return
	}

	resumen, err := obtenerResumenFinanciero(id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Data["json"] = resumen
	c.ServeJSON()
}

// ============================================================
// GET /v1/progreso-educativo/:id
// Avance del usuario en cada curso.
// ============================================================

func (c *MidController) GetProgresoEducativo() {

	id, ok := c.leerId()
	if !ok {
		return
	}

	progreso, err := obtenerProgresoEducativo(id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Data["json"] = progreso
	c.ServeJSON()
}

// ============================================================
// GET /v1/ofertas
// Productos de credito activos con los datos del banco.
// ============================================================

func (c *MidController) GetOfertas() {

	// --------------------------------------------------------
	// 1. Consultar productos y bancos en crud_negocio
	// --------------------------------------------------------

	var productos []ProductoCrediticio

	if err := consultarCrud(urlCrud["negocio"]+"/producto_crediticio?"+filtro("activo", true), &productos); err != nil {
		c.responderError(err)
		return
	}

	bancos, err := obtenerBancos()
	if err != nil {
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 2. Unir cada producto con su banco
	// --------------------------------------------------------

	ofertas := []map[string]interface{}{}

	for _, p := range productos {

		banco := bancos[p.IdBanco.Id]

		requisitos := []string{}
		for _, r := range strings.Split(p.Requisitos, ",") {
			if r = strings.TrimSpace(r); r != "" {
				requisitos = append(requisitos, r)
			}
		}

		ofertas = append(ofertas, map[string]interface{}{
			"id_producto":  p.Id,
			"producto":     p.NombreProducto,
			"descripcion":  p.Descripcion,
			"banco":        banco.NombreBanco,
			"logo_banco":   banco.UrlLogo,
			"sitio_web":    banco.SitioWeb,
			"monto_minimo": p.MontoMinimo,
			"monto_maximo": p.MontoMaximo,
			"tasa_minima":  p.TasaMinima,
			"tasa_maxima":  p.TasaMaxima,
			"plazo_minimo": p.PlazoMinimo,
			"plazo_maximo": p.PlazoMaximo,
			"requisitos":   requisitos,
		})
	}

	// --------------------------------------------------------
	// 3. Devolver el JSON
	// --------------------------------------------------------

	c.Data["json"] = ofertas
	c.ServeJSON()
}

// ============================================================
// POST /v1/solicitud-credito
// Valida usuario, producto, monto y plazo y crea la solicitud.
// ============================================================

func (c *MidController) PostSolicitudCredito() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosSolicitudCredito

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.responderError(nuevoError(http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido"))
		return
	}

	if datos.IdUsuario <= 0 || datos.IdProducto <= 0 {
		c.responderError(nuevoError(http.StatusBadRequest, "id_usuario e id_producto son obligatorios"))
		return
	}

	// --------------------------------------------------------
	// 2. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	if err := existeUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 3. Revisar el producto y los rangos (crud_negocio)
	// --------------------------------------------------------

	var producto ProductoCrediticio

	if err := consultarCrud(fmt.Sprintf("%s/producto_crediticio/%d", urlCrud["negocio"], datos.IdProducto), &producto); err != nil {
		c.responderError(cambiarMensaje404(err, "El producto de credito no existe"))
		return
	}

	if !producto.Activo {
		c.responderError(nuevoError(http.StatusBadRequest, "El producto de credito ya no esta disponible"))
		return
	}

	if datos.Monto < producto.MontoMinimo || datos.Monto > producto.MontoMaximo {
		c.responderError(nuevoError(http.StatusBadRequest, fmt.Sprintf("El monto debe estar entre %.0f y %.0f", producto.MontoMinimo, producto.MontoMaximo)))
		return
	}

	if datos.PlazoMeses < producto.PlazoMinimo || datos.PlazoMeses > producto.PlazoMaximo {
		c.responderError(nuevoError(http.StatusBadRequest, fmt.Sprintf("El plazo debe estar entre %d y %d meses", producto.PlazoMinimo, producto.PlazoMaximo)))
		return
	}

	// --------------------------------------------------------
	// 4. Asignar el primer asesor activo del banco
	// --------------------------------------------------------

	var asesores []AsesorBancario
	consultarCrud(urlCrud["negocio"]+"/asesor_bancario?"+filtro("id_banco", producto.IdBanco.Id, "activo", true), &asesores)

	// --------------------------------------------------------
	// 5. Guardar la solicitud (lead)
	// --------------------------------------------------------

	lead := Lead{
		IdUsuario:     datos.IdUsuario,
		IdProducto:    &Ref{Id: producto.Id},
		TipoCredito:   producto.NombreProducto,
		MontoInteres:  datos.Monto,
		PlazoInteres:  datos.PlazoMeses,
		EstadoLead:    "nuevo",
		Observaciones: datos.Observaciones,
		Activo:        true,
	}

	if len(asesores) > 0 {
		lead.IdAsesor = &Ref{Id: asesores[0].Id}
	}

	if err := enviarCrud("POST", urlCrud["negocio"]+"/lead", lead, &lead); err != nil {
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 6. Dejar registro de la actividad (crud_soporte)
	// --------------------------------------------------------

	registrarActividad(datos.IdUsuario, "CREAR_SOLICITUD",
		fmt.Sprintf("Solicitud de %s por %.0f a %d meses", producto.NombreProducto, datos.Monto, datos.PlazoMeses),
		"negocio.lead")

	// --------------------------------------------------------
	// 7. Devolver la solicitud completa
	// --------------------------------------------------------

	solicitudes, err := completarSolicitudes([]Lead{lead})
	if err != nil || len(solicitudes) == 0 {
		c.responderError(nuevoError(http.StatusBadGateway, "La solicitud se creo pero no se pudo consultar"))
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = solicitudes[0]
	c.ServeJSON()
}

// ============================================================
// GET /v1/solicitud-credito/usuario/:id
// Solicitudes de credito de un usuario.
// ============================================================

func (c *MidController) GetSolicitudesUsuario() {

	id, ok := c.leerId()
	if !ok {
		return
	}

	solicitudes, err := obtenerSolicitudesUsuario(id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Data["json"] = solicitudes
	c.ServeJSON()
}

// ============================================================
// POST /v1/pqr
// Radica una PQR en estado "Abierta".
// ============================================================

func (c *MidController) PostPqr() {

	// --------------------------------------------------------
	// 1. Leer y validar los datos enviados
	// --------------------------------------------------------

	var datos DatosPqr

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.responderError(nuevoError(http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido"))
		return
	}

	datos.Tipo = strings.ToLower(strings.TrimSpace(datos.Tipo))
	datos.Prioridad = strings.ToLower(strings.TrimSpace(datos.Prioridad))

	if datos.Prioridad == "" {
		datos.Prioridad = "media"
	}

	tiposValidos := map[string]bool{"peticion": true, "queja": true, "reclamo": true, "sugerencia": true}
	prioridadesValidas := map[string]bool{"alta": true, "media": true, "baja": true}

	switch {
	case datos.IdUsuario <= 0:
		c.responderError(nuevoError(http.StatusBadRequest, "id_usuario es obligatorio"))
		return
	case strings.TrimSpace(datos.Titulo) == "":
		c.responderError(nuevoError(http.StatusBadRequest, "El titulo es obligatorio"))
		return
	case strings.TrimSpace(datos.Descripcion) == "":
		c.responderError(nuevoError(http.StatusBadRequest, "La descripcion es obligatoria"))
		return
	case !tiposValidos[datos.Tipo]:
		c.responderError(nuevoError(http.StatusBadRequest, "El tipo debe ser: peticion, queja, reclamo o sugerencia"))
		return
	case !prioridadesValidas[datos.Prioridad]:
		c.responderError(nuevoError(http.StatusBadRequest, "La prioridad debe ser: alta, media o baja"))
		return
	}

	// --------------------------------------------------------
	// 2. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	if err := existeUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	// --------------------------------------------------------
	// 3. Buscar el estado "Abierta" (crud_soporte)
	// --------------------------------------------------------

	var estados []Catalogo

	if err := consultarCrud(urlCrud["soporte"]+"/estado_pqr?"+filtro("nombre", "Abierta"), &estados); err != nil {
		c.responderError(err)
		return
	}

	if len(estados) == 0 {
		c.responderError(nuevoError(http.StatusInternalServerError, "No existe el estado 'Abierta' en la base de datos"))
		return
	}

	// --------------------------------------------------------
	// 4. Guardar la PQR con un numero de radicado unico
	// --------------------------------------------------------

	ahora := time.Now()

	pqr := Pqr{
		IdUsuario:   datos.IdUsuario,
		Radicado:    fmt.Sprintf("PQR-%s-%03d", ahora.Format("20060102-150405"), ahora.Nanosecond()/1e6),
		Titulo:      strings.TrimSpace(datos.Titulo),
		Tipo:        datos.Tipo,
		Categoria:   datos.Categoria,
		Prioridad:   datos.Prioridad,
		Descripcion: strings.TrimSpace(datos.Descripcion),
		IdEstado:    &Ref{Id: estados[0].Id},
		Activo:      true,
	}

	if err := enviarCrud("POST", urlCrud["soporte"]+"/pqr", pqr, &pqr); err != nil {
		c.responderError(err)
		return
	}

	registrarActividad(datos.IdUsuario, "CREAR_PQR", "Se radico la PQR "+pqr.Radicado, "soporte.pqr")

	// --------------------------------------------------------
	// 5. Devolver la PQR creada
	// --------------------------------------------------------

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = completarPqr(pqr, map[int]string{estados[0].Id: estados[0].Nombre})
	c.ServeJSON()
}

// ============================================================
// GET /v1/pqr/usuario/:id
// PQR de un usuario con su estado y sus adjuntos.
// ============================================================

func (c *MidController) GetPqrUsuario() {

	id, ok := c.leerId()
	if !ok {
		return
	}

	pqrs, err := obtenerPqrUsuario(id)
	if err != nil {
		c.responderError(err)
		return
	}

	c.Data["json"] = pqrs
	c.ServeJSON()
}

// ============================================================
// GET /v1/dashboard/:id
// Todo junto. Si un CRUD esta apagado se avisa y se
// entrega el resto.
// ============================================================

func (c *MidController) GetDashboard() {

	id, ok := c.leerId()
	if !ok {
		return
	}

	// 1. El perfil es obligatorio
	perfil, err := obtenerPerfil(id)
	if err != nil {
		c.responderError(err)
		return
	}

	panel := map[string]interface{}{"perfil": perfil}
	avisos := []string{}

	// 2. Cada seccion se agrega si su CRUD responde
	if finanzas, err := obtenerResumenFinanciero(id); err == nil {
		panel["finanzas"] = finanzas
	} else {
		avisos = append(avisos, "finanzas: "+err.Error())
	}

	if educacion, err := obtenerProgresoEducativo(id); err == nil {
		panel["educacion"] = educacion
	} else {
		avisos = append(avisos, "educacion: "+err.Error())
	}

	if solicitudes, err := obtenerSolicitudesUsuario(id); err == nil {
		panel["solicitudes_credito"] = solicitudes
	} else {
		avisos = append(avisos, "negocio: "+err.Error())
	}

	if pqrs, err := obtenerPqrUsuario(id); err == nil {
		panel["pqr"] = pqrs
	} else {
		avisos = append(avisos, "soporte: "+err.Error())
	}

	if len(avisos) > 0 {
		panel["avisos"] = avisos
	}

	// 3. Devolver el JSON
	c.Data["json"] = panel
	c.ServeJSON()
}

// ============================================================
// /v1/crud/:conjunto/*  (cualquier metodo)
// Pasarela: reenvia la peticion tal cual al CRUD indicado.
// Ej: GET /v1/crud/finanzas/categoria -> GET :8082/v1/categoria
// ============================================================

func (c *MidController) Reenviar() {

	// --------------------------------------------------------
	// 1. Revisar que el conjunto exista
	// --------------------------------------------------------

	conjunto := c.Ctx.Input.Param(":conjunto")
	base, existe := urlCrud[conjunto]

	if !existe {
		c.responderError(nuevoError(http.StatusNotFound, "El conjunto '"+conjunto+"' no existe. Usa: auth, finanzas, educacion, negocio o soporte"))
		return
	}

	// --------------------------------------------------------
	// 2. Armar la URL del CRUD
	// --------------------------------------------------------

	direccion := base + "/" + c.Ctx.Input.Param(":splat")
	if c.Ctx.Request.URL.RawQuery != "" {
		direccion += "?" + c.Ctx.Request.URL.RawQuery
	}

	// --------------------------------------------------------
	// 3. Enviar la peticion con el mismo metodo y cuerpo
	// --------------------------------------------------------

	peticion, err := http.NewRequest(c.Ctx.Input.Method(), direccion, bytes.NewReader(c.Ctx.Input.RequestBody))
	if err != nil {
		c.responderError(nuevoError(http.StatusBadRequest, "Direccion invalida: "+direccion))
		return
	}
	peticion.Header.Set("Content-Type", "application/json")

	respuesta, err := http.DefaultClient.Do(peticion)
	if err != nil {
		c.responderError(nuevoError(http.StatusBadGateway, "No fue posible comunicarse con el CRUD de "+conjunto))
		return
	}
	defer respuesta.Body.Close()

	cuerpo, _ := io.ReadAll(respuesta.Body)

	// --------------------------------------------------------
	// 4. Devolver la respuesta del CRUD sin cambios
	// --------------------------------------------------------

	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Ctx.Output.SetStatus(respuesta.StatusCode)
	c.Ctx.Output.Body(cuerpo)
}

// ############################################################
// Funciones que arman la informacion
// (las usan varias rutas y el dashboard)
// ############################################################

// obtenerPerfil junta usuario, tipo de documento y roles.
func obtenerPerfil(idUsuario int) (Perfil, error) {

	// 1. Consultar el usuario
	var usuario Usuario

	if err := consultarCrud(fmt.Sprintf("%s/usuario/%d", urlCrud["auth"], idUsuario), &usuario); err != nil {
		return Perfil{}, cambiarMensaje404(err, "El usuario no existe")
	}

	perfil := Perfil{
		IdUsuario:     usuario.Id,
		Nombre:        usuario.Nombre,
		Apellido:      usuario.Apellido,
		Email:         usuario.Email,
		Cedula:        usuario.Cedula,
		Telefono:      usuario.Telefono,
		Ciudad:        usuario.Ciudad,
		Direccion:     usuario.Direccion,
		Estado:        usuario.Estado,
		FechaRegistro: usuario.FechaRegistro,
		Roles:         []string{},
	}

	if !usuario.FechaNacimiento.IsZero() {
		perfil.FechaNacimiento = usuario.FechaNacimiento.Format("2006-01-02")
	}

	if !usuario.FechaUltimaSesion.IsZero() {
		perfil.UltimaSesion = &usuario.FechaUltimaSesion
	}

	// 2. Consultar el tipo de documento
	if usuario.TipoDocumento != nil {
		var tipo TipoDocumento
		if consultarCrud(fmt.Sprintf("%s/tipo_documento/%d", urlCrud["auth"], usuario.TipoDocumento.Id), &tipo) == nil {
			perfil.TipoDocumento = tipo.Codigo
		}
	}

	// 3. Consultar los roles
	var asignaciones []UsuarioRol

	if err := consultarCrud(urlCrud["auth"]+"/usuario_rol?"+filtro("id_usuario", idUsuario, "activo", true), &asignaciones); err != nil {
		return Perfil{}, err
	}

	for _, a := range asignaciones {
		var rol Rol
		if consultarCrud(fmt.Sprintf("%s/rol/%d", urlCrud["auth"], a.IdRol.Id), &rol) == nil {
			perfil.Roles = append(perfil.Roles, rol.NombreRol)
		}
	}

	return perfil, nil
}

// existeUsuario confirma que el usuario existe y esta activo.
func existeUsuario(idUsuario int) error {

	var usuario Usuario

	if err := consultarCrud(fmt.Sprintf("%s/usuario/%d", urlCrud["auth"], idUsuario), &usuario); err != nil {
		return cambiarMensaje404(err, "El usuario no existe")
	}

	if !usuario.Activo {
		return nuevoError(http.StatusForbidden, "El usuario esta inactivo")
	}

	return nil
}

// obtenerResumenFinanciero calcula todo el resumen de finanzas.
func obtenerResumenFinanciero(idUsuario int) (ResumenFinanciero, error) {

	resumen := ResumenFinanciero{
		IdUsuario:          idUsuario,
		GastosPorCategoria: []map[string]interface{}{},
		UltimosMovimientos: []map[string]interface{}{},
		Metas:              []map[string]interface{}{},
		Inversiones:        []map[string]interface{}{},
		SolicitudesDeuda:   []map[string]interface{}{},
	}

	if err := existeUsuario(idUsuario); err != nil {
		return resumen, err
	}

	delUsuario := filtro("id_usuario", idUsuario, "activo", true)

	// --------------------------------------------------------
	// 1. Movimientos: ingresos, gastos y gasto por categoria
	// --------------------------------------------------------

	var categorias []Categoria
	if err := consultarCrud(urlCrud["finanzas"]+"/categoria?limit=1000", &categorias); err != nil {
		return resumen, err
	}

	nombreCategoria := map[int]string{}
	for _, cat := range categorias {
		nombreCategoria[cat.Id] = cat.Nombre
	}

	var movimientos []Movimiento
	if err := consultarCrud(urlCrud["finanzas"]+"/movimiento_ingreso_egreso?"+delUsuario, &movimientos); err != nil {
		return resumen, err
	}

	gastoCategoria := map[string]float64{}

	for _, m := range movimientos {
		if m.EsIngreso {
			resumen.TotalIngresos += m.Monto
		} else {
			resumen.TotalGastos += m.Monto
			gastoCategoria[nombreDeCategoria(m, nombreCategoria)] += m.Monto
		}
	}

	resumen.Balance = resumen.TotalIngresos - resumen.TotalGastos

	for categoria, total := range gastoCategoria {
		resumen.GastosPorCategoria = append(resumen.GastosPorCategoria, map[string]interface{}{
			"categoria": categoria,
			"total":     total,
		})
	}

	sort.Slice(resumen.GastosPorCategoria, func(i, j int) bool {
		return resumen.GastosPorCategoria[i]["total"].(float64) > resumen.GastosPorCategoria[j]["total"].(float64)
	})

	// Los 5 movimientos mas recientes
	sort.Slice(movimientos, func(i, j int) bool { return movimientos[i].Fecha.After(movimientos[j].Fecha) })

	for i, m := range movimientos {

		if i == 5 {
			break
		}

		tipo := "gasto"
		if m.EsIngreso {
			tipo = "ingreso"
		}

		resumen.UltimosMovimientos = append(resumen.UltimosMovimientos, map[string]interface{}{
			"id":            m.Id,
			"fecha":         fecha(m.Fecha),
			"concepto":      m.Nombre,
			"categoria":     nombreDeCategoria(m, nombreCategoria),
			"tipo":          tipo,
			"valor":         m.Monto,
			"metodo_pago":   m.MetodoPago,
			"observaciones": m.Observaciones,
		})
	}

	// --------------------------------------------------------
	// 2. Metas: porcentaje de avance
	// --------------------------------------------------------

	var metas []Meta
	if err := consultarCrud(urlCrud["finanzas"]+"/meta?"+delUsuario, &metas); err != nil {
		return resumen, err
	}

	for _, m := range metas {

		porcentaje := 0.0
		if m.MontoObjetivo > 0 {
			porcentaje = math.Min(m.MontoActual/m.MontoObjetivo*100, 100)
		}

		resumen.Metas = append(resumen.Metas, map[string]interface{}{
			"id":           m.Id,
			"nombre":       m.Nombre,
			"icono":        m.Icono,
			"color":        m.Color,
			"actual":       m.MontoActual,
			"objetivo":     m.MontoObjetivo,
			"porcentaje":   redondear(porcentaje),
			"cumplida":     m.MontoActual >= m.MontoObjetivo,
			"fecha_limite": fecha(m.FechaLimite),
		})
	}

	// --------------------------------------------------------
	// 3. Inversiones: total invertido y ganancia estimada
	// --------------------------------------------------------

	var inversiones []Inversion
	if err := consultarCrud(urlCrud["finanzas"]+"/inversion?"+delUsuario, &inversiones); err != nil {
		return resumen, err
	}

	var riesgos, tipos []Catalogo
	consultarCrud(urlCrud["finanzas"]+"/nivel_riesgo?limit=1000", &riesgos)
	consultarCrud(urlCrud["finanzas"]+"/tipo_inversion?limit=1000", &tipos)

	nombreRiesgo := map[int]string{}
	for _, r := range riesgos {
		nombreRiesgo[r.Id] = r.Nombre
	}

	nombreTipo := map[int]string{}
	for _, t := range tipos {
		nombreTipo[t.Id] = t.Nombre
	}

	for _, inv := range inversiones {

		resumen.TotalInvertido += inv.Monto
		resumen.GananciaEstimada += inv.Monto * inv.Rentabilidad / 100

		detalle := map[string]interface{}{
			"id":           inv.Id,
			"nombre":       inv.Nombre,
			"monto":        inv.Monto,
			"rentabilidad": inv.Rentabilidad,
			"fecha_inicio": fecha(inv.FechaInicio),
			"fecha_fin":    fecha(inv.FechaFin),
		}

		if inv.IdTipoInversion != nil {
			detalle["tipo"] = nombreTipo[inv.IdTipoInversion.Id]
		}

		if inv.IdNivelRiesgo != nil {
			detalle["riesgo"] = nombreRiesgo[inv.IdNivelRiesgo.Id]
		}

		resumen.Inversiones = append(resumen.Inversiones, detalle)
	}

	resumen.GananciaEstimada = redondear(resumen.GananciaEstimada)

	// --------------------------------------------------------
	// 4. Solicitudes de "Resuelve tu deuda"
	// --------------------------------------------------------

	var solicitudes []SolicitudConsolidacion
	if err := consultarCrud(urlCrud["finanzas"]+"/solicitud_consolidacion?"+delUsuario, &solicitudes); err != nil {
		return resumen, err
	}

	for _, s := range solicitudes {
		resumen.SolicitudesDeuda = append(resumen.SolicitudesDeuda, map[string]interface{}{
			"id":              s.Id,
			"saldo_total":     s.SaldoTotal,
			"cuota_actual":    s.CuotaActual,
			"cuota_propuesta": s.CuotaPropuesta,
			"ahorro_mensual":  s.CuotaActual - s.CuotaPropuesta,
			"estado":          s.Estado,
		})
	}

	return resumen, nil
}

// obtenerProgresoEducativo calcula el avance de cada curso.
func obtenerProgresoEducativo(idUsuario int) (ProgresoGeneral, error) {

	progreso := ProgresoGeneral{IdUsuario: idUsuario, Modulos: []map[string]interface{}{}}

	if err := existeUsuario(idUsuario); err != nil {
		return progreso, err
	}

	// --------------------------------------------------------
	// 1. Consultar modulos, lecciones y avances (crud_educacion)
	// --------------------------------------------------------

	var modulos []ModuloEducativo
	if err := consultarCrud(urlCrud["educacion"]+"/modulo_educativo?"+filtro("activo", true), &modulos); err != nil {
		return progreso, err
	}

	var lecciones []Leccion
	if err := consultarCrud(urlCrud["educacion"]+"/leccion?"+filtro("activo", true), &lecciones); err != nil {
		return progreso, err
	}

	var avances []ProgresoEducativo
	if err := consultarCrud(urlCrud["educacion"]+"/progreso_educativo?"+filtro("id_usuario", idUsuario), &avances); err != nil {
		return progreso, err
	}

	var leccionesVistas []ProgresoLeccion
	if err := consultarCrud(urlCrud["educacion"]+"/progreso_leccion?"+filtro("id_usuario", idUsuario, "completado", true), &leccionesVistas); err != nil {
		return progreso, err
	}

	// --------------------------------------------------------
	// 2. Contar lecciones por modulo y lecciones completadas
	// --------------------------------------------------------

	moduloDeLeccion := map[int]int{}
	totalLecciones := map[int]int{}

	for _, l := range lecciones {
		if l.IdModulo != nil {
			moduloDeLeccion[l.Id] = l.IdModulo.Id
			totalLecciones[l.IdModulo.Id]++
		}
	}

	completadas := map[int]int{}
	for _, pl := range leccionesVistas {
		if pl.IdLeccion != nil {
			completadas[moduloDeLeccion[pl.IdLeccion.Id]]++
		}
	}

	avancePorModulo := map[int]ProgresoEducativo{}
	for _, a := range avances {
		if a.IdModulo != nil {
			avancePorModulo[a.IdModulo.Id] = a
		}
	}

	// --------------------------------------------------------
	// 3. Armar el avance de cada modulo
	// --------------------------------------------------------

	suma := 0

	for _, m := range modulos {

		porcentaje := 0
		var calificacion interface{}

		// Si hay registro de progreso se usa; si no, se calcula con las lecciones
		if avance, ok := avancePorModulo[m.Id]; ok {
			porcentaje = avance.PorcentajeCompletado
			if avance.Calificacion > 0 {
				calificacion = avance.Calificacion
			}
		} else if totalLecciones[m.Id] > 0 {
			porcentaje = completadas[m.Id] * 100 / totalLecciones[m.Id]
		}

		estado := "sin iniciar"
		if porcentaje >= 100 {
			estado = "completado"
			progreso.ModulosCompletados++
		} else if porcentaje > 0 {
			estado = "en curso"
		}

		suma += porcentaje

		progreso.Modulos = append(progreso.Modulos, map[string]interface{}{
			"id_modulo":             m.Id,
			"titulo":                m.Titulo,
			"descripcion":           m.Descripcion,
			"nivel":                 m.Nivel,
			"imagen":                m.UrlThumbnail,
			"porcentaje":            porcentaje,
			"estado":                estado,
			"lecciones_totales":     totalLecciones[m.Id],
			"lecciones_completadas": completadas[m.Id],
			"calificacion":          calificacion,
		})
	}

	progreso.ModulosTotales = len(modulos)
	if len(modulos) > 0 {
		progreso.PorcentajeGeneral = redondear(float64(suma) / float64(len(modulos)))
	}

	return progreso, nil
}

// obtenerSolicitudesUsuario lista las solicitudes de credito de un usuario.
func obtenerSolicitudesUsuario(idUsuario int) ([]SolicitudCredito, error) {

	if err := existeUsuario(idUsuario); err != nil {
		return []SolicitudCredito{}, err
	}

	var leads []Lead
	if err := consultarCrud(urlCrud["negocio"]+"/lead?"+filtro("id_usuario", idUsuario, "activo", true)+"&sortby=fecha_generacion&order=desc", &leads); err != nil {
		return []SolicitudCredito{}, err
	}

	return completarSolicitudes(leads)
}

// completarSolicitudes agrega producto, banco, asesor y cuota estimada a cada lead.
func completarSolicitudes(leads []Lead) ([]SolicitudCredito, error) {

	solicitudes := []SolicitudCredito{}

	bancos, err := obtenerBancos()
	if err != nil {
		return solicitudes, err
	}

	for _, l := range leads {

		solicitud := SolicitudCredito{
			IdSolicitud:    l.Id,
			Producto:       l.TipoCredito,
			Monto:          l.MontoInteres,
			PlazoMeses:     l.PlazoInteres,
			Estado:         l.EstadoLead,
			FechaSolicitud: l.FechaGeneracion,
		}

		var producto ProductoCrediticio
		if l.IdProducto != nil && consultarCrud(fmt.Sprintf("%s/producto_crediticio/%d", urlCrud["negocio"], l.IdProducto.Id), &producto) == nil {
			solicitud.Producto = producto.NombreProducto
			solicitud.Banco = bancos[producto.IdBanco.Id].NombreBanco
			solicitud.CuotaEstimada = cuotaMensual(l.MontoInteres, producto.TasaMinima, l.PlazoInteres)
		}

		var asesor AsesorBancario
		if l.IdAsesor != nil && consultarCrud(fmt.Sprintf("%s/asesor_bancario/%d", urlCrud["negocio"], l.IdAsesor.Id), &asesor) == nil {
			solicitud.Asesor = asesor.Nombre + " " + asesor.Apellido
		}

		solicitudes = append(solicitudes, solicitud)
	}

	return solicitudes, nil
}

// obtenerPqrUsuario lista las PQR de un usuario.
func obtenerPqrUsuario(idUsuario int) ([]PqrCompleta, error) {

	lista := []PqrCompleta{}

	if err := existeUsuario(idUsuario); err != nil {
		return lista, err
	}

	var pqrs []Pqr
	if err := consultarCrud(urlCrud["soporte"]+"/pqr?"+filtro("id_usuario", idUsuario, "activo", true)+"&sortby=fecha_creacion&order=desc", &pqrs); err != nil {
		return lista, err
	}

	var estados []Catalogo
	if err := consultarCrud(urlCrud["soporte"]+"/estado_pqr?limit=1000", &estados); err != nil {
		return lista, err
	}

	nombreEstado := map[int]string{}
	for _, e := range estados {
		nombreEstado[e.Id] = e.Nombre
	}

	for _, p := range pqrs {
		lista = append(lista, completarPqr(p, nombreEstado))
	}

	return lista, nil
}

// completarPqr agrega el nombre del estado y los adjuntos a una PQR.
func completarPqr(p Pqr, nombreEstado map[int]string) PqrCompleta {

	completa := PqrCompleta{
		IdPqr:               p.Id,
		Radicado:            p.Radicado,
		Titulo:              p.Titulo,
		Tipo:                p.Tipo,
		Categoria:           p.Categoria,
		Prioridad:           p.Prioridad,
		Descripcion:         p.Descripcion,
		Asesor:              p.Asesor,
		Respuesta:           p.MensajeRespuesta,
		FechaCreacion:       p.FechaCreacion,
		UltimaActualizacion: p.FechaModificacion,
		Adjuntos:            []map[string]interface{}{},
	}

	if p.IdEstado != nil {
		completa.Estado = nombreEstado[p.IdEstado.Id]
	}

	var adjuntos []Adjunto
	consultarCrud(urlCrud["soporte"]+"/adjunto?"+filtro("id_pqr", p.Id, "activo", true), &adjuntos)

	for _, a := range adjuntos {
		completa.Adjuntos = append(completa.Adjuntos, map[string]interface{}{
			"nombre":       a.NombreArchivo,
			"ruta":         a.RutaArchivo,
			"tipo":         a.TipoMime,
			"tamano_bytes": a.TamanoBytes,
		})
	}

	return completa
}

// registrarActividad deja una huella en soporte.registro_actividad.
// Si falla no detiene la operacion principal.
func registrarActividad(idUsuario int, tipo, descripcion, entidad string) {

	actividad := RegistroActividad{
		IdUsuario:       &idUsuario,
		TipoActividad:   tipo,
		Descripcion:     descripcion,
		EntidadAfectada: entidad,
	}

	enviarCrud("POST", urlCrud["soporte"]+"/registro_actividad", actividad, nil)
}

// obtenerBancos devuelve los bancos por su id.
func obtenerBancos() (map[int]Banco, error) {

	var bancos []Banco
	if err := consultarCrud(urlCrud["negocio"]+"/banco?limit=1000", &bancos); err != nil {
		return nil, err
	}

	resultado := map[int]Banco{}
	for _, b := range bancos {
		resultado[b.Id] = b
	}

	return resultado, nil
}

// cuotaMensual calcula la cuota fija de un credito.
// tasaAnual viene en porcentaje: 12 = 12% anual.
func cuotaMensual(monto, tasaAnual float64, meses int) float64 {

	if meses <= 0 {
		return 0
	}

	tasaMensual := tasaAnual / 100 / 12
	if tasaMensual == 0 {
		return redondear(monto / float64(meses))
	}

	return redondear(monto * tasaMensual / (1 - math.Pow(1+tasaMensual, float64(-meses))))
}

// ############################################################
// Funciones para hablar con los CRUD
// ############################################################

// errorMid es un error con el codigo HTTP que debe responder el MID.
type errorMid struct {
	status  int
	mensaje string
}

func (e *errorMid) Error() string {
	return e.mensaje
}

func nuevoError(status int, mensaje string) *errorMid {
	return &errorMid{status: status, mensaje: mensaje}
}

// consultarCrud hace un GET a un CRUD y guarda el JSON en "destino".
func consultarCrud(direccion string, destino interface{}) error {
	return enviarCrud("GET", direccion, nil, destino)
}

// enviarCrud hace GET, POST, PUT o DELETE a un CRUD.
func enviarCrud(metodo, direccion string, cuerpo interface{}, destino interface{}) error {

	// --------------------------------------------------------
	// 1. Convertir el cuerpo a JSON (si hay)
	// --------------------------------------------------------

	var lector io.Reader

	if cuerpo != nil {
		datos, err := json.Marshal(cuerpo)
		if err != nil {
			return nuevoError(http.StatusInternalServerError, "No fue posible preparar los datos")
		}
		lector = bytes.NewReader(datos)
	}

	// --------------------------------------------------------
	// 2. Hacer la peticion HTTP
	// --------------------------------------------------------

	peticion, err := http.NewRequest(metodo, direccion, lector)
	if err != nil {
		return nuevoError(http.StatusInternalServerError, "Direccion invalida: "+direccion)
	}
	peticion.Header.Set("Content-Type", "application/json")

	respuesta, err := http.DefaultClient.Do(peticion)
	if err != nil {
		return nuevoError(http.StatusBadGateway, "No fue posible comunicarse con "+direccion+". Verifica que el CRUD este encendido.")
	}

	// Cerramos la respuesta HTTP.
	defer respuesta.Body.Close()

	// --------------------------------------------------------
	// 3. Verificar el codigo HTTP y leer el JSON recibido
	// --------------------------------------------------------

	if respuesta.StatusCode == http.StatusNotFound {
		return nuevoError(http.StatusNotFound, "La ruta no existe: "+direccion)
	}

	if respuesta.StatusCode >= 300 {
		return nuevoError(http.StatusBadGateway, fmt.Sprintf("El CRUD respondio con error %d", respuesta.StatusCode))
	}

	body, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return nuevoError(http.StatusBadGateway, "Error leyendo la respuesta del CRUD")
	}

	// --------------------------------------------------------
	// 4. Los CRUD de bee responden los errores como un texto
	//    entre comillas con codigo 200. Ej: "<QuerySeter> no row found"
	//    (una lista vacia llega como null y no es error)
	// --------------------------------------------------------

	var texto string

	if bytes.HasPrefix(bytes.TrimSpace(body), []byte(`"`)) && json.Unmarshal(body, &texto) == nil {

		if texto == "OK" {
			return nil
		}

		if strings.Contains(texto, "no row found") {
			return nuevoError(http.StatusNotFound, "No se encontro el registro")
		}

		return nuevoError(http.StatusBadRequest, "El CRUD rechazo la operacion: "+texto)
	}

	// --------------------------------------------------------
	// 5. Convertir JSON a la estructura de destino
	// --------------------------------------------------------

	if destino != nil {
		if err := json.Unmarshal(body, destino); err != nil {
			return nuevoError(http.StatusBadGateway, "Respuesta invalida del CRUD")
		}
	}

	return nil
}

// filtro arma el "query" de bee.
// Ej: filtro("id_usuario", 1, "activo", true) -> "query=id_usuario:1,activo:true&limit=1000"
func filtro(pares ...interface{}) string {

	condiciones := []string{}

	for i := 0; i+1 < len(pares); i += 2 {
		condiciones = append(condiciones, fmt.Sprintf("%v:%v", pares[i], pares[i+1]))
	}

	return "query=" + url.QueryEscape(strings.Join(condiciones, ",")) + "&limit=1000"
}

// ############################################################
// Funciones pequenas de apoyo
// ############################################################

// leerId lee el :id de la URL. Si no es valido responde 400.
func (c *MidController) leerId() (int, bool) {

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil || id <= 0 {
		c.responderError(nuevoError(http.StatusBadRequest, "El ID debe ser numerico"))
		return 0, false
	}

	return id, true
}

// responderError responde {"error": "..."} con el codigo que corresponda.
func (c *MidController) responderError(err error) {

	status := http.StatusInternalServerError

	if e, ok := err.(*errorMid); ok {
		status = e.status
	}

	c.Ctx.ResponseWriter.WriteHeader(status)

	c.Data["json"] = map[string]interface{}{
		"error": err.Error(),
	}

	c.ServeJSON()
}

// cambiarMensaje404 cambia el mensaje de "no encontrado" por uno mas claro.
func cambiarMensaje404(err error, mensaje string) error {

	if e, ok := err.(*errorMid); ok && e.status == http.StatusNotFound {
		return nuevoError(http.StatusNotFound, mensaje)
	}

	return err
}

func nombreDeCategoria(m Movimiento, nombres map[int]string) string {
	if m.IdCategoria == nil {
		return "Sin categoria"
	}
	return nombres[m.IdCategoria.Id]
}

func fecha(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func redondear(valor float64) float64 {
	return math.Round(valor*100) / 100
}

func recortar(texto string, largo int) string {
	if len(texto) > largo {
		return texto[:largo]
	}
	return texto
}
