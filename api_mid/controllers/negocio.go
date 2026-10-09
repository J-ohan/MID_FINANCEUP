package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type Banco struct {
	Id          int    `json:"Id"`
	NombreBanco string `json:"NombreBanco"`
	UrlLogo     string `json:"UrlLogo"`
}

type ProductoCrediticio struct {
	Id             int     `json:"Id"`
	IdBanco        *Ref    `json:"IdBanco"`
	NombreProducto string  `json:"NombreProducto"`
	Descripcion    string  `json:"Descripcion"`
	MontoMinimo    float64 `json:"MontoMinimo"`
	MontoMaximo    float64 `json:"MontoMaximo"`
	TasaMinima     float64 `json:"TasaMinima"`
	TasaMaxima     float64 `json:"TasaMaxima"`
	PlazoMinimo    int     `json:"PlazoMinimo"`
	PlazoMaximo    int     `json:"PlazoMaximo"`
	Requisitos     string  `json:"Requisitos"`
	Activo         bool    `json:"Activo"`
}

type AsesorBancario struct {
	Id       int    `json:"Id"`
	Nombre   string `json:"Nombre"`
	Apellido string `json:"Apellido"`
}

type Lead struct {
	Id              int       `json:"Id"`
	IdUsuario       int       `json:"IdUsuario"`
	IdProducto      *Ref      `json:"IdProducto"`
	IdAsesor        *Ref      `json:"IdAsesor"`
	TipoCredito     string    `json:"TipoCredito"`
	MontoInteres    float64   `json:"MontoInteres"`
	PlazoInteres    int       `json:"PlazoInteres"`
	EstadoLead      string    `json:"EstadoLead"`
	FechaGeneracion time.Time `json:"FechaGeneracion"`
	Activo          bool      `json:"Activo"`
}

type DatosSolicitudCredito struct {
	IdUsuario  int     `json:"id_usuario"`
	IdProducto int     `json:"id_producto"`
	Monto      float64 `json:"monto"`
	PlazoMeses int     `json:"plazo_meses"`
}

// NegocioController operations for Negocio
type NegocioController struct {
	beego.Controller
}

// ============================================================
// GET /v1/ofertas
// Productos de credito con el nombre de su banco (crud_negocio)
// ============================================================

func (c *NegocioController) GetOfertas() {

	// --------------------------------------------------------
	// 1. Consultar los productos de credito activos
	// --------------------------------------------------------

	responseProductos, err := http.Get("http://localhost:8084/v1/producto_crediticio?query=activo:true&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de negocio"}
		c.ServeJSON()
		return
	}

	defer responseProductos.Body.Close()

	bodyProductos, _ := io.ReadAll(responseProductos.Body)

	var productos []ProductoCrediticio
	json.Unmarshal(bodyProductos, &productos)

	// --------------------------------------------------------
	// 2. Consultar los bancos
	// --------------------------------------------------------

	responseBancos, err := http.Get("http://localhost:8084/v1/banco?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los bancos"}
		c.ServeJSON()
		return
	}

	defer responseBancos.Body.Close()

	bodyBancos, _ := io.ReadAll(responseBancos.Body)

	var bancos []Banco
	json.Unmarshal(bodyBancos, &bancos)

	// Guardamos cada banco por su id para encontrarlo rapido
	bancoPorId := map[int]Banco{}
	for _, banco := range bancos {
		bancoPorId[banco.Id] = banco
	}

	// --------------------------------------------------------
	// 3. Unir cada producto con su banco
	// --------------------------------------------------------

	ofertas := []map[string]interface{}{}

	for _, producto := range productos {

		banco := bancoPorId[producto.IdBanco.Id]

		ofertas = append(ofertas, map[string]interface{}{
			"id_producto":  producto.Id,
			"producto":     producto.NombreProducto,
			"descripcion":  producto.Descripcion,
			"banco":        banco.NombreBanco,
			"logo_banco":   banco.UrlLogo,
			"monto_minimo": producto.MontoMinimo,
			"monto_maximo": producto.MontoMaximo,
			"tasa_minima":  producto.TasaMinima,
			"tasa_maxima":  producto.TasaMaxima,
			"plazo_minimo": producto.PlazoMinimo,
			"plazo_maximo": producto.PlazoMaximo,
			"requisitos":   producto.Requisitos,
		})
	}

	// --------------------------------------------------------
	// 4. Devolver las ofertas
	// --------------------------------------------------------

	c.Data["json"] = ofertas
	c.ServeJSON()
}

// ============================================================
// POST /v1/solicitud-credito
// Valida usuario, producto, monto y plazo y crea la solicitud
// (crud_auth + crud_negocio + crud_soporte)
// ============================================================

func (c *NegocioController) PostSolicitudCredito() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosSolicitudCredito

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario Usuario

	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar el producto de credito (crud_negocio)
	// --------------------------------------------------------

	responseProducto, err := http.Get(fmt.Sprintf("http://localhost:8084/v1/producto_crediticio/%d", datos.IdProducto))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de negocio"}
		c.ServeJSON()
		return
	}

	defer responseProducto.Body.Close()

	bodyProducto, _ := io.ReadAll(responseProducto.Body)

	var producto ProductoCrediticio

	if err := json.Unmarshal(bodyProducto, &producto); err != nil || !producto.Activo {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El producto de credito no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Revisar que el monto y el plazo esten permitidos
	// --------------------------------------------------------

	if datos.Monto < producto.MontoMinimo || datos.Monto > producto.MontoMaximo {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{
			"error": fmt.Sprintf("El monto debe estar entre %.0f y %.0f", producto.MontoMinimo, producto.MontoMaximo),
		}
		c.ServeJSON()
		return
	}

	if datos.PlazoMeses < producto.PlazoMinimo || datos.PlazoMeses > producto.PlazoMaximo {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{
			"error": fmt.Sprintf("El plazo debe estar entre %d y %d meses", producto.PlazoMinimo, producto.PlazoMaximo),
		}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Buscar un asesor del banco del producto
	// --------------------------------------------------------

	responseAsesores, err := http.Get(fmt.Sprintf("http://localhost:8084/v1/asesor_bancario?query=id_banco:%d,activo:true", producto.IdBanco.Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los asesores"}
		c.ServeJSON()
		return
	}

	defer responseAsesores.Body.Close()

	bodyAsesores, _ := io.ReadAll(responseAsesores.Body)

	var asesores []AsesorBancario
	json.Unmarshal(bodyAsesores, &asesores)

	// --------------------------------------------------------
	// 6. Guardar la solicitud (lead) en crud_negocio
	// --------------------------------------------------------

	solicitud := Lead{
		IdUsuario:    datos.IdUsuario,
		IdProducto:   &Ref{Id: producto.Id},
		TipoCredito:  producto.NombreProducto,
		MontoInteres: datos.Monto,
		PlazoInteres: datos.PlazoMeses,
		EstadoLead:   "nuevo",
		Activo:       true,
	}

	nombreAsesor := ""
	if len(asesores) > 0 {
		solicitud.IdAsesor = &Ref{Id: asesores[0].Id}
		nombreAsesor = asesores[0].Nombre + " " + asesores[0].Apellido
	}

	jsonSolicitud, _ := json.Marshal(solicitud)

	responseCrear, err := http.Post("http://localhost:8084/v1/lead", "application/json", bytes.NewBuffer(jsonSolicitud))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de negocio"}
		c.ServeJSON()
		return
	}

	defer responseCrear.Body.Close()

	bodyCrear, _ := io.ReadAll(responseCrear.Body)

	if responseCrear.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo crear la solicitud: " + string(bodyCrear)}
		c.ServeJSON()
		return
	}

	json.Unmarshal(bodyCrear, &solicitud)

	// --------------------------------------------------------
	// 7. Dejar registro de la actividad en crud_soporte
	// --------------------------------------------------------

	actividad := RegistroActividad{
		IdUsuario:       &datos.IdUsuario,
		TipoActividad:   "CREAR_SOLICITUD",
		Descripcion:     fmt.Sprintf("Solicitud de %s por %.0f", producto.NombreProducto, datos.Monto),
		EntidadAfectada: "negocio.lead",
	}

	jsonActividad, _ := json.Marshal(actividad)

	if responseActividad, err := http.Post("http://localhost:8085/v1/registro_actividad", "application/json", bytes.NewBuffer(jsonActividad)); err == nil {
		responseActividad.Body.Close()
	}

	// --------------------------------------------------------
	// 8. Calcular la cuota mensual estimada (cuota fija)
	//    cuota = monto * i / (1 - (1 + i)^-meses)
	//    donde i = tasa anual / 12 / 100
	// --------------------------------------------------------

	i := producto.TasaMinima / 12 / 100
	cuota := datos.Monto * i / (1 - math.Pow(1+i, float64(-datos.PlazoMeses)))

	// --------------------------------------------------------
	// 9. Devolver la solicitud creada
	// --------------------------------------------------------

	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"id_solicitud":   solicitud.Id,
		"producto":       producto.NombreProducto,
		"monto":          datos.Monto,
		"plazo_meses":    datos.PlazoMeses,
		"cuota_estimada": math.Round(cuota*100) / 100,
		"estado":         solicitud.EstadoLead,
		"asesor":         nombreAsesor,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/solicitud-credito/usuario/:id
// Solicitudes de credito de un usuario (crud_negocio)
// ============================================================

func (c *NegocioController) GetSolicitudesUsuario() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar las solicitudes del usuario
	// --------------------------------------------------------

	responseSolicitudes, err := http.Get(fmt.Sprintf("http://localhost:8084/v1/lead?query=id_usuario:%d,activo:true&limit=1000", id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de negocio"}
		c.ServeJSON()
		return
	}

	defer responseSolicitudes.Body.Close()

	bodySolicitudes, _ := io.ReadAll(responseSolicitudes.Body)

	var solicitudes []Lead
	json.Unmarshal(bodySolicitudes, &solicitudes)

	// --------------------------------------------------------
	// 3. Construir la lista y devolverla
	// --------------------------------------------------------

	lista := []map[string]interface{}{}

	for _, solicitud := range solicitudes {
		lista = append(lista, map[string]interface{}{
			"id_solicitud":    solicitud.Id,
			"producto":        solicitud.TipoCredito,
			"monto":           solicitud.MontoInteres,
			"plazo_meses":     solicitud.PlazoInteres,
			"estado":          solicitud.EstadoLead,
			"fecha_solicitud": solicitud.FechaGeneracion,
		})
	}

	c.Data["json"] = lista
	c.ServeJSON()
}
