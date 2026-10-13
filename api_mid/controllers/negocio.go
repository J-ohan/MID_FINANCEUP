package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

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

// NegocioController operations for Negocio
type NegocioController struct {
	beego.Controller
}

// ============================================================
// GET /v1/ofertas
// Ofertas de los aliados con su banco, etiquetas, perfiles y
// requisitos (crud_negocio)
// ============================================================

func (c *NegocioController) GetOfertas() {

	// --------------------------------------------------------
	// 1. Consultar los productos de credito activos
	// --------------------------------------------------------

	responseProductos, err := http.Get(urlCrudNegocio + "/producto_crediticio?query=activo:true&limit=1000")

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
	// 2. Consultar los bancos (aliados)
	// --------------------------------------------------------

	responseBancos, err := http.Get(urlCrudNegocio + "/banco?limit=1000")

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
	// 3. Consultar los detalles (etiquetas, perfiles, requisitos)
	// --------------------------------------------------------

	responseDetalles, err := http.Get(urlCrudNegocio + "/producto_detalle?query=activo:true&sortby=Orden&order=asc&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el detalle de las ofertas"}
		c.ServeJSON()
		return
	}

	defer responseDetalles.Body.Close()

	bodyDetalles, _ := io.ReadAll(responseDetalles.Body)

	var detalles []ProductoDetalle
	json.Unmarshal(bodyDetalles, &detalles)

	// textos[idProducto][tipo] = lista de textos
	textos := map[int]map[string][]string{}

	for _, detalle := range detalles {
		idProducto := detalle.IdProducto.Id
		if textos[idProducto] == nil {
			textos[idProducto] = map[string][]string{}
		}
		textos[idProducto][detalle.Tipo] = append(textos[idProducto][detalle.Tipo], detalle.Texto)
	}

	// --------------------------------------------------------
	// 4. Unir cada producto con su banco y sus detalles
	// --------------------------------------------------------

	ofertas := []map[string]interface{}{}

	for _, producto := range productos {

		banco := bancoPorId[producto.IdBanco.Id]

		ofertas = append(ofertas, map[string]interface{}{
			"id":             producto.Id,
			"slug":           producto.Slug,
			"aliado":         banco.NombreBanco,
			"tipo":           banco.TipoAliado,
			"logo_texto":     banco.LogoTexto,
			"logo_fondo":     banco.LogoFondo,
			"categoria":      producto.NombreProducto,
			"descripcion":    producto.Descripcion,
			"promesa":        producto.Promesa,
			"meta":           producto.Vigencia,
			"familia":        producto.Familia,
			"beneficio":      producto.Beneficio,
			"tasa_desde":     producto.TasaMinima,
			"monto_maximo":   producto.MontoMaximo,
			"plazo_maximo":   producto.PlazoMaximo,
			"aprobacion":     producto.TiempoAprobacion,
			"calificacion":   producto.Calificacion,
			"usuarios":       producto.Usuarios,
			"compatibilidad": producto.Compatibilidad,
			"destacada":      producto.Destacada,
			"cta_primaria":   producto.CtaPrimaria,
			"cta_secundaria": producto.CtaSecundaria,
			"etiquetas":      textos[producto.Id]["etiqueta"],
			"perfiles":       textos[producto.Id]["perfil"],
			"requisitos":     textos[producto.Id]["requisito"],
		})
	}

	// --------------------------------------------------------
	// 5. Devolver las ofertas
	// --------------------------------------------------------

	c.Data["json"] = ofertas
	c.ServeJSON()
}

// ============================================================
// GET /v1/ofertas/:id
// Ficha completa de una oferta: caracteristicas, documentos,
// tarifas, preguntas y otras ofertas de la misma familia
// ============================================================

func (c *NegocioController) GetOferta() {

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
	// 2. Consultar el producto
	// --------------------------------------------------------

	responseProducto, err := http.Get(fmt.Sprintf("%s/producto_crediticio/%d", urlCrudNegocio, id))

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
		c.Data["json"] = map[string]interface{}{"error": "La oferta no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar su banco
	// --------------------------------------------------------

	banco := Banco{}

	responseBanco, err := http.Get(fmt.Sprintf("%s/banco/%d", urlCrudNegocio, producto.IdBanco.Id))

	if err == nil {
		defer responseBanco.Body.Close()
		bodyBanco, _ := io.ReadAll(responseBanco.Body)
		json.Unmarshal(bodyBanco, &banco)
	}

	// --------------------------------------------------------
	// 4. Consultar sus detalles y separarlos por tipo
	// --------------------------------------------------------

	responseDetalles, err := http.Get(fmt.Sprintf("%s/producto_detalle?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", urlCrudNegocio, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el detalle de la oferta"}
		c.ServeJSON()
		return
	}

	defer responseDetalles.Body.Close()

	bodyDetalles, _ := io.ReadAll(responseDetalles.Body)

	var detalles []ProductoDetalle
	json.Unmarshal(bodyDetalles, &detalles)

	textos := map[string][]string{}
	for _, detalle := range detalles {
		textos[detalle.Tipo] = append(textos[detalle.Tipo], detalle.Texto)
	}

	// --------------------------------------------------------
	// 5. Consultar sus tarifas
	// --------------------------------------------------------

	responseTarifas, err := http.Get(fmt.Sprintf("%s/producto_tarifa?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", urlCrudNegocio, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las tarifas"}
		c.ServeJSON()
		return
	}

	defer responseTarifas.Body.Close()

	bodyTarifas, _ := io.ReadAll(responseTarifas.Body)

	var tarifas []ProductoTarifa
	json.Unmarshal(bodyTarifas, &tarifas)

	listaTarifas := []map[string]interface{}{}
	for _, tarifa := range tarifas {
		listaTarifas = append(listaTarifas, map[string]interface{}{"concepto": tarifa.Concepto, "valor": tarifa.Valor})
	}

	// --------------------------------------------------------
	// 6. Consultar sus preguntas frecuentes
	// --------------------------------------------------------

	responsePreguntas, err := http.Get(fmt.Sprintf("%s/producto_pregunta?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", urlCrudNegocio, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las preguntas"}
		c.ServeJSON()
		return
	}

	defer responsePreguntas.Body.Close()

	bodyPreguntas, _ := io.ReadAll(responsePreguntas.Body)

	var preguntas []ProductoPregunta
	json.Unmarshal(bodyPreguntas, &preguntas)

	listaPreguntas := []map[string]interface{}{}
	for _, pregunta := range preguntas {
		listaPreguntas = append(listaPreguntas, map[string]interface{}{"pregunta": pregunta.Pregunta, "respuesta": pregunta.Respuesta})
	}

	// --------------------------------------------------------
	// 7. Buscar otras ofertas de la misma familia
	// --------------------------------------------------------

	relacionadas := []map[string]interface{}{}

	responseFamilia, err := http.Get(urlCrudNegocio + "/producto_crediticio?query=activo:true,familia:" + producto.Familia + "&limit=1000")

	if err == nil {
		defer responseFamilia.Body.Close()
		bodyFamilia, _ := io.ReadAll(responseFamilia.Body)

		var familia []ProductoCrediticio
		json.Unmarshal(bodyFamilia, &familia)

		for _, otro := range familia {
			if otro.Id != producto.Id {
				relacionadas = append(relacionadas, map[string]interface{}{
					"id":         otro.Id,
					"categoria":  otro.NombreProducto,
					"tasa_desde": otro.TasaMinima,
				})
			}
		}
	}

	// --------------------------------------------------------
	// 8. Devolver la ficha completa
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id":              producto.Id,
		"slug":            producto.Slug,
		"aliado":          banco.NombreBanco,
		"tipo":            banco.TipoAliado,
		"logo_texto":      banco.LogoTexto,
		"logo_fondo":      banco.LogoFondo,
		"categoria":       producto.NombreProducto,
		"descripcion":     producto.Descripcion,
		"promesa":         producto.Promesa,
		"meta":            producto.Vigencia,
		"familia":         producto.Familia,
		"beneficio":       producto.Beneficio,
		"tasa_desde":      producto.TasaMinima,
		"monto_minimo":    producto.MontoMinimo,
		"monto_maximo":    producto.MontoMaximo,
		"plazo_minimo":    producto.PlazoMinimo,
		"plazo_maximo":    producto.PlazoMaximo,
		"aprobacion":      producto.TiempoAprobacion,
		"calificacion":    producto.Calificacion,
		"usuarios":        producto.Usuarios,
		"compatibilidad":  producto.Compatibilidad,
		"cta_primaria":    producto.CtaPrimaria,
		"cta_secundaria":  producto.CtaSecundaria,
		"etiquetas":       textos["etiqueta"],
		"perfiles":        textos["perfil"],
		"requisitos":      textos["requisito"],
		"caracteristicas": textos["caracteristica"],
		"documentos":      textos["documento"],
		"tarifas":         listaTarifas,
		"preguntas":       listaPreguntas,
		"relacionadas":    relacionadas,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/solicitud-credito
// Valida usuario, producto, monto y plazo y radica la solicitud
// con un numero SOL-xxxxxx (crud_auth + crud_negocio + crud_soporte)
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
	// 2. Validar los datos del solicitante
	// --------------------------------------------------------

	if datos.Nombre == "" || datos.Documento == "" || datos.Celular == "" || !strings.Contains(datos.Correo, "@") {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Faltan datos: nombre, documento, correo y celular son obligatorios"}
		c.ServeJSON()
		return
	}

	if datos.Ingresos <= 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Los ingresos mensuales deben ser mayores a 0"}
		c.ServeJSON()
		return
	}

	if !datos.AceptaTerminos {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Debes aceptar los terminos y la autorizacion de datos"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("%s/usuario/%d", urlCrudAuth, datos.IdUsuario))

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
	// 4. Consultar el producto de credito (crud_negocio)
	// --------------------------------------------------------

	responseProducto, err := http.Get(fmt.Sprintf("%s/producto_crediticio/%d", urlCrudNegocio, datos.IdProducto))

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
	// 5. Revisar que el monto y el plazo esten permitidos
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
	// 6. Buscar un asesor del banco del producto
	// --------------------------------------------------------

	responseAsesores, err := http.Get(fmt.Sprintf("%s/asesor_bancario?query=id_banco:%d,activo:true", urlCrudNegocio, producto.IdBanco.Id))

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
	// 7. Guardar la solicitud (lead) con su numero de radicado
	//    SOL- + 6 numeros al azar, ej: SOL-482913
	// --------------------------------------------------------

	numero, _ := rand.Int(rand.Reader, big.NewInt(1000000))

	solicitud := Lead{
		IdUsuario:            datos.IdUsuario,
		IdProducto:           &Ref{Id: producto.Id},
		Radicado:             fmt.Sprintf("SOL-%06d", numero.Int64()),
		TipoCredito:          producto.NombreProducto,
		MontoInteres:         datos.Monto,
		PlazoInteres:         datos.PlazoMeses,
		NombreSolicitante:    datos.Nombre,
		DocumentoSolicitante: datos.Documento,
		CorreoSolicitante:    datos.Correo,
		CelularSolicitante:   datos.Celular,
		IngresosMensuales:    datos.Ingresos,
		AceptaTerminos:       true,
		EstadoLead:           "nuevo",
		Activo:               true,
	}

	nombreAsesor := ""
	if len(asesores) > 0 {
		solicitud.IdAsesor = &Ref{Id: asesores[0].Id}
		nombreAsesor = asesores[0].Nombre + " " + asesores[0].Apellido
	}

	jsonSolicitud, _ := json.Marshal(solicitud)

	responseCrear, err := http.Post(urlCrudNegocio+"/lead", "application/json", bytes.NewBuffer(jsonSolicitud))

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
	// 8. Dejar registro de la actividad en crud_soporte
	// --------------------------------------------------------

	actividad := RegistroActividad{
		IdUsuario:       &datos.IdUsuario,
		TipoActividad:   "CREAR_SOLICITUD",
		Descripcion:     fmt.Sprintf("Solicitud %s de %s por %.0f", solicitud.Radicado, producto.NombreProducto, datos.Monto),
		EntidadAfectada: "negocio.lead",
	}

	jsonActividad, _ := json.Marshal(actividad)

	if responseActividad, err := http.Post(urlCrudSoporte+"/registro_actividad", "application/json", bytes.NewBuffer(jsonActividad)); err == nil {
		responseActividad.Body.Close()
	}

	// --------------------------------------------------------
	// 9. Calcular la cuota mensual estimada (cuota fija)
	//    cuota = monto * i / (1 - (1 + i)^-meses)
	//    donde i = tasa anual / 12 / 100
	// --------------------------------------------------------

	cuota := datos.Monto / float64(datos.PlazoMeses)

	i := producto.TasaMinima / 12 / 100
	if i > 0 {
		cuota = datos.Monto * i / (1 - math.Pow(1+i, float64(-datos.PlazoMeses)))
	}

	// --------------------------------------------------------
	// 10. Devolver la solicitud radicada
	// --------------------------------------------------------

	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"id_solicitud":   solicitud.Id,
		"radicado":       solicitud.Radicado,
		"producto":       producto.NombreProducto,
		"monto":          datos.Monto,
		"plazo_meses":    datos.PlazoMeses,
		"cuota_estimada": math.Round(cuota*100) / 100,
		"estado":         "En estudio",
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

	responseSolicitudes, err := http.Get(fmt.Sprintf("%s/lead?query=id_usuario:%d,activo:true&limit=1000", urlCrudNegocio, id))

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
			"radicado":        solicitud.Radicado,
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
