package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type EstadoPqr struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

type Pqr struct {
	Id                int       `json:"Id"`
	IdUsuario         int       `json:"IdUsuario"`
	Radicado          string    `json:"Radicado"`
	Titulo            string    `json:"Titulo"`
	Tipo              string    `json:"Tipo"`
	Categoria         string    `json:"Categoria"`
	Prioridad         string    `json:"Prioridad"`
	Asesor            string    `json:"Asesor"`
	MensajeRespuesta  string    `json:"MensajeRespuesta"`
	Progreso          int       `json:"Progreso"`
	Descripcion       string    `json:"Descripcion"`
	IdEstado          *Ref      `json:"IdEstado"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type Adjunto struct {
	Id            int    `json:"Id"`
	IdPqr         *Ref   `json:"IdPqr"`
	NombreArchivo string `json:"NombreArchivo"`
	RutaArchivo   string `json:"RutaArchivo"`
	TipoMime      string `json:"TipoMime"`
	TamanoBytes   int    `json:"TamanoBytes"`
	Activo        bool   `json:"Activo"`
}

type RespuestaPqr struct {
	Id             int       `json:"Id"`
	IdPqr          *Ref      `json:"IdPqr"`
	IdUsuario      *int      `json:"IdUsuario"`
	Autor          string    `json:"Autor"`
	Mensaje        string    `json:"Mensaje"`
	FechaRespuesta time.Time `json:"FechaRespuesta"`
	Activo         bool      `json:"Activo"`
}

type RegistroActividad struct {
	IdUsuario       *int   `json:"IdUsuario"`
	TipoActividad   string `json:"TipoActividad"`
	Descripcion     string `json:"Descripcion"`
	EntidadAfectada string `json:"EntidadAfectada"`
}

type DatosPqr struct {
	IdUsuario   int    `json:"id_usuario"`
	Titulo      string `json:"titulo"`
	Tipo        string `json:"tipo"`
	Categoria   string `json:"categoria"`
	Prioridad   string `json:"prioridad"`
	Descripcion string `json:"descripcion"`
	Adjuntos    []struct {
		NombreArchivo string `json:"nombre_archivo"`
		RutaArchivo   string `json:"ruta_archivo"`
		TipoMime      string `json:"tipo_mime"`
		TamanoBytes   int    `json:"tamano_bytes"`
	} `json:"adjuntos"`
}

type DatosRespuestaPqr struct {
	IdUsuario int    `json:"id_usuario"`
	Autor     string `json:"autor"`
	Mensaje   string `json:"mensaje"`
	Estado    string `json:"estado"`
	Progreso  int    `json:"progreso"`
}

// SoporteController operations for Soporte
type SoporteController struct {
	beego.Controller
}

// ============================================================
// POST /v1/pqr
// Radica una PQR en estado "Abierta" con sus adjuntos
// (crud_auth + crud_soporte)
// ============================================================

func (c *SoporteController) PostPqr() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosPqr

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El cuerpo de la peticion no es un JSON valido"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Validar los datos
	// --------------------------------------------------------

	if datos.Titulo == "" || datos.Descripcion == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El titulo y la descripcion son obligatorios"}
		c.ServeJSON()
		return
	}

	if datos.Tipo != "peticion" && datos.Tipo != "queja" && datos.Tipo != "reclamo" && datos.Tipo != "sugerencia" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El tipo debe ser: peticion, queja, reclamo o sugerencia"}
		c.ServeJSON()
		return
	}

	if datos.Prioridad == "" {
		datos.Prioridad = "media"
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
	// 4. Buscar el estado "Abierta" (crud_soporte)
	// --------------------------------------------------------

	responseEstado, err := http.Get(urlCrudSoporte + "/estado_pqr?query=nombre:Abierta")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responseEstado.Body.Close()

	bodyEstado, _ := io.ReadAll(responseEstado.Body)

	var estados []EstadoPqr
	json.Unmarshal(bodyEstado, &estados)

	if len(estados) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"error": "No existe el estado Abierta en la base de datos"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Guardar la PQR con un numero de radicado
	//    (PQR- + fecha y hora, ej: PQR-20261009-153045)
	// --------------------------------------------------------

	pqr := Pqr{
		IdUsuario:   datos.IdUsuario,
		Radicado:    "PQR-" + time.Now().Format("20060102-150405"),
		Titulo:      datos.Titulo,
		Tipo:        datos.Tipo,
		Categoria:   datos.Categoria,
		Prioridad:   datos.Prioridad,
		Descripcion: datos.Descripcion,
		Progreso:    0,
		IdEstado:    &Ref{Id: estados[0].Id},
		Activo:      true,
	}

	jsonPqr, _ := json.Marshal(pqr)

	responseCrear, err := http.Post(urlCrudSoporte+"/pqr", "application/json", bytes.NewBuffer(jsonPqr))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responseCrear.Body.Close()

	bodyCrear, _ := io.ReadAll(responseCrear.Body)

	if responseCrear.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo crear la PQR: " + string(bodyCrear)}
		c.ServeJSON()
		return
	}

	json.Unmarshal(bodyCrear, &pqr)

	// --------------------------------------------------------
	// 6. Guardar los adjuntos (nombre y ruta de cada archivo)
	//    Nota: aqui se guarda la informacion del archivo.
	//    El archivo como tal lo sube la pagina a una carpeta
	//    o servicio de archivos y envia su ruta.
	// --------------------------------------------------------

	adjuntosGuardados := 0

	for _, archivo := range datos.Adjuntos {

		adjunto := Adjunto{
			IdPqr:         &Ref{Id: pqr.Id},
			NombreArchivo: archivo.NombreArchivo,
			RutaArchivo:   archivo.RutaArchivo,
			TipoMime:      archivo.TipoMime,
			TamanoBytes:   archivo.TamanoBytes,
			Activo:        true,
		}

		jsonAdjunto, _ := json.Marshal(adjunto)

		responseAdjunto, err := http.Post(urlCrudSoporte+"/adjunto", "application/json", bytes.NewBuffer(jsonAdjunto))

		if err == nil {
			if responseAdjunto.StatusCode == http.StatusCreated {
				adjuntosGuardados++
			}
			responseAdjunto.Body.Close()
		}
	}

	// --------------------------------------------------------
	// 7. Dejar registro de la actividad
	// --------------------------------------------------------

	actividad := RegistroActividad{
		IdUsuario:       &datos.IdUsuario,
		TipoActividad:   "CREAR_PQR",
		Descripcion:     "Se radico la PQR " + pqr.Radicado,
		EntidadAfectada: "soporte.pqr",
	}

	jsonActividad, _ := json.Marshal(actividad)

	if responseActividad, err := http.Post(urlCrudSoporte+"/registro_actividad", "application/json", bytes.NewBuffer(jsonActividad)); err == nil {
		responseActividad.Body.Close()
	}

	// --------------------------------------------------------
	// 8. Devolver la PQR creada
	// --------------------------------------------------------

	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"id_pqr":    pqr.Id,
		"radicado":  pqr.Radicado,
		"titulo":    pqr.Titulo,
		"tipo":      pqr.Tipo,
		"prioridad": pqr.Prioridad,
		"estado":    estados[0].Nombre,
		"adjuntos":  adjuntosGuardados,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/pqr/usuario/:id
// PQR de un usuario con el nombre de su estado (crud_soporte)
// ============================================================

func (c *SoporteController) GetPqrUsuario() {

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
	// 2. Consultar las PQR del usuario
	// --------------------------------------------------------

	responsePqrs, err := http.Get(fmt.Sprintf("%s/pqr?query=id_usuario:%d,activo:true&sortby=Id&order=desc&limit=1000", urlCrudSoporte, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responsePqrs.Body.Close()

	bodyPqrs, _ := io.ReadAll(responsePqrs.Body)

	var pqrs []Pqr
	json.Unmarshal(bodyPqrs, &pqrs)

	// --------------------------------------------------------
	// 3. Consultar los estados (para saber su nombre)
	// --------------------------------------------------------

	responseEstados, err := http.Get(urlCrudSoporte + "/estado_pqr?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los estados"}
		c.ServeJSON()
		return
	}

	defer responseEstados.Body.Close()

	bodyEstados, _ := io.ReadAll(responseEstados.Body)

	var estados []EstadoPqr
	json.Unmarshal(bodyEstados, &estados)

	nombreEstado := map[int]string{}
	for _, estado := range estados {
		nombreEstado[estado.Id] = estado.Nombre
	}

	// --------------------------------------------------------
	// 4. Construir la lista y devolverla
	// --------------------------------------------------------

	lista := []map[string]interface{}{}

	for _, pqr := range pqrs {

		estado := ""
		if pqr.IdEstado != nil {
			estado = nombreEstado[pqr.IdEstado.Id]
		}

		lista = append(lista, map[string]interface{}{
			"id_pqr":         pqr.Id,
			"radicado":       pqr.Radicado,
			"titulo":         pqr.Titulo,
			"tipo":           pqr.Tipo,
			"categoria":      pqr.Categoria,
			"prioridad":      pqr.Prioridad,
			"estado":         estado,
			"progreso":       pqr.Progreso,
			"descripcion":    pqr.Descripcion,
			"fecha_creacion": pqr.FechaCreacion,
		})
	}

	c.Data["json"] = lista
	c.ServeJSON()
}

// ============================================================
// GET /v1/pqr/radicado/:radicado
// Una PQR con su estado, respuestas y adjuntos (pagina ver-pqr)
// ============================================================

func (c *SoporteController) GetPqr() {

	// --------------------------------------------------------
	// 1. Obtener el radicado enviado en la URL
	// --------------------------------------------------------

	radicado := c.Ctx.Input.Param(":radicado")

	// --------------------------------------------------------
	// 2. Buscar la PQR por su radicado
	// --------------------------------------------------------

	responsePqr, err := http.Get(urlCrudSoporte + "/pqr?query=radicado:" + url.QueryEscape(radicado) + "&limit=1")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responsePqr.Body.Close()

	bodyPqr, _ := io.ReadAll(responsePqr.Body)

	var pqrs []Pqr
	json.Unmarshal(bodyPqr, &pqrs)

	if len(pqrs) == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "No existe una PQR con ese radicado"}
		c.ServeJSON()
		return
	}

	pqr := pqrs[0]

	// --------------------------------------------------------
	// 3. Consultar el nombre de su estado
	// --------------------------------------------------------

	estado := EstadoPqr{}

	responseEstado, err := http.Get(fmt.Sprintf("%s/estado_pqr/%d", urlCrudSoporte, pqr.IdEstado.Id))

	if err == nil {
		defer responseEstado.Body.Close()
		bodyEstado, _ := io.ReadAll(responseEstado.Body)
		json.Unmarshal(bodyEstado, &estado)
	}

	// --------------------------------------------------------
	// 4. Consultar las respuestas (de la mas vieja a la mas nueva)
	// --------------------------------------------------------

	responseRespuestas, err := http.Get(fmt.Sprintf("%s/respuesta_pqr?query=id_pqr:%d,activo:true&sortby=Id&order=asc&limit=1000", urlCrudSoporte, pqr.Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las respuestas"}
		c.ServeJSON()
		return
	}

	defer responseRespuestas.Body.Close()

	bodyRespuestas, _ := io.ReadAll(responseRespuestas.Body)

	var respuestas []RespuestaPqr
	json.Unmarshal(bodyRespuestas, &respuestas)

	listaRespuestas := []map[string]interface{}{}
	for _, respuesta := range respuestas {
		listaRespuestas = append(listaRespuestas, map[string]interface{}{
			"autor":   respuesta.Autor,
			"mensaje": respuesta.Mensaje,
			"fecha":   respuesta.FechaRespuesta,
		})
	}

	// --------------------------------------------------------
	// 5. Consultar los adjuntos
	// --------------------------------------------------------

	responseAdjuntos, err := http.Get(fmt.Sprintf("%s/adjunto?query=id_pqr:%d,activo:true&limit=1000", urlCrudSoporte, pqr.Id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los adjuntos"}
		c.ServeJSON()
		return
	}

	defer responseAdjuntos.Body.Close()

	bodyAdjuntos, _ := io.ReadAll(responseAdjuntos.Body)

	var adjuntos []Adjunto
	json.Unmarshal(bodyAdjuntos, &adjuntos)

	listaAdjuntos := []map[string]interface{}{}
	for _, adjunto := range adjuntos {
		listaAdjuntos = append(listaAdjuntos, map[string]interface{}{
			"nombre": adjunto.NombreArchivo,
			"ruta":   adjunto.RutaArchivo,
			"tipo":   adjunto.TipoMime,
			"tamano": adjunto.TamanoBytes,
		})
	}

	// --------------------------------------------------------
	// 6. Devolver la PQR completa
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_pqr":               pqr.Id,
		"radicado":             pqr.Radicado,
		"titulo":               pqr.Titulo,
		"tipo":                 pqr.Tipo,
		"categoria":            pqr.Categoria,
		"prioridad":            pqr.Prioridad,
		"descripcion":          pqr.Descripcion,
		"estado":               estado.Nombre,
		"progreso":             pqr.Progreso,
		"asesor":               pqr.Asesor,
		"mensaje_respuesta":    pqr.MensajeRespuesta,
		"fecha_creacion":       pqr.FechaCreacion,
		"ultima_actualizacion": pqr.FechaModificacion,
		"respuestas":           listaRespuestas,
		"adjuntos":             listaAdjuntos,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/pqr/:id/respuesta
// Agrega una respuesta a la PQR y (si llega) cambia su estado
// y su progreso. La usan el asesor (admin) y el usuario.
// ============================================================

func (c *SoporteController) PostRespuestaPqr() {

	// --------------------------------------------------------
	// 1. Obtener el ID de la PQR y los datos enviados
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	var datos DatosRespuestaPqr

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil || datos.Mensaje == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El mensaje es obligatorio"}
		c.ServeJSON()
		return
	}

	if datos.Autor == "" {
		datos.Autor = "Usuario"
	}

	// --------------------------------------------------------
	// 2. Consultar la PQR
	// --------------------------------------------------------

	responsePqr, err := http.Get(fmt.Sprintf("%s/pqr/%d", urlCrudSoporte, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responsePqr.Body.Close()

	bodyPqr, _ := io.ReadAll(responsePqr.Body)

	var pqr Pqr

	if err := json.Unmarshal(bodyPqr, &pqr); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "La PQR no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Guardar la respuesta
	// --------------------------------------------------------

	respuesta := RespuestaPqr{
		IdPqr:   &Ref{Id: id},
		Autor:   datos.Autor,
		Mensaje: datos.Mensaje,
		Activo:  true,
	}

	if datos.IdUsuario > 0 {
		respuesta.IdUsuario = &datos.IdUsuario
	}

	jsonRespuesta, _ := json.Marshal(respuesta)

	responseCrear, err := http.Post(urlCrudSoporte+"/respuesta_pqr", "application/json", bytes.NewBuffer(jsonRespuesta))

	if err != nil || responseCrear.StatusCode != http.StatusCreated {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo guardar la respuesta"}
		c.ServeJSON()
		return
	}

	responseCrear.Body.Close()

	// --------------------------------------------------------
	// 4. Si llego un estado nuevo, buscar su id
	// --------------------------------------------------------

	if datos.Estado != "" {

		responseEstado, err := http.Get(urlCrudSoporte + "/estado_pqr?query=nombre:" + url.QueryEscape(datos.Estado))

		if err == nil {
			defer responseEstado.Body.Close()
			bodyEstado, _ := io.ReadAll(responseEstado.Body)

			var estados []EstadoPqr
			json.Unmarshal(bodyEstado, &estados)

			if len(estados) > 0 {
				pqr.IdEstado = &Ref{Id: estados[0].Id}
			}
		}
	}

	// --------------------------------------------------------
	// 5. Actualizar la PQR (progreso, ultima respuesta, asesor)
	// --------------------------------------------------------

	if datos.Progreso > 0 && datos.Progreso <= 100 {
		pqr.Progreso = datos.Progreso
	}

	if datos.Autor != "Usuario" {
		pqr.Asesor = datos.Autor
		pqr.MensajeRespuesta = datos.Mensaje
	}

	jsonPqr, _ := json.Marshal(pqr)

	peticionPut, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/pqr/%d", urlCrudSoporte, id), bytes.NewBuffer(jsonPqr))
	peticionPut.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionPut); err == nil {
		responsePut.Body.Close()
	}

	// --------------------------------------------------------
	// 6. Responder
	// --------------------------------------------------------

	c.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)

	c.Data["json"] = map[string]interface{}{
		"mensaje":  "Respuesta guardada",
		"radicado": pqr.Radicado,
		"progreso": pqr.Progreso,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/admin/pqr
// Todas las PQR con su estado y el nombre del usuario
// (crud_auth + crud_soporte)
// ============================================================

func (c *SoporteController) GetPqrAdmin() {

	// --------------------------------------------------------
	// 1. Consultar todas las PQR activas
	// --------------------------------------------------------

	responsePqrs, err := http.Get(urlCrudSoporte + "/pqr?query=activo:true&sortby=Id&order=desc&limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de soporte"}
		c.ServeJSON()
		return
	}

	defer responsePqrs.Body.Close()

	bodyPqrs, _ := io.ReadAll(responsePqrs.Body)

	var pqrs []Pqr
	json.Unmarshal(bodyPqrs, &pqrs)

	// --------------------------------------------------------
	// 2. Consultar los estados
	// --------------------------------------------------------

	responseEstados, err := http.Get(urlCrudSoporte + "/estado_pqr?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los estados"}
		c.ServeJSON()
		return
	}

	defer responseEstados.Body.Close()

	bodyEstados, _ := io.ReadAll(responseEstados.Body)

	var estados []EstadoPqr
	json.Unmarshal(bodyEstados, &estados)

	nombreEstado := map[int]string{}
	for _, estado := range estados {
		nombreEstado[estado.Id] = estado.Nombre
	}

	// --------------------------------------------------------
	// 3. Consultar los usuarios (para poner su nombre)
	// --------------------------------------------------------

	responseUsuarios, err := http.Get(urlCrudAuth + "/usuario?limit=10000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de usuarios"}
		c.ServeJSON()
		return
	}

	defer responseUsuarios.Body.Close()

	bodyUsuarios, _ := io.ReadAll(responseUsuarios.Body)

	var usuarios []Usuario
	json.Unmarshal(bodyUsuarios, &usuarios)

	nombreUsuario := map[int]string{}
	for _, usuario := range usuarios {
		nombreUsuario[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	// --------------------------------------------------------
	// 4. Construir la lista y devolverla
	// --------------------------------------------------------

	lista := []map[string]interface{}{}

	for _, pqr := range pqrs {
		lista = append(lista, map[string]interface{}{
			"id_pqr":         pqr.Id,
			"radicado":       pqr.Radicado,
			"titulo":         pqr.Titulo,
			"tipo":           pqr.Tipo,
			"prioridad":      pqr.Prioridad,
			"estado":         nombreEstado[pqr.IdEstado.Id],
			"progreso":       pqr.Progreso,
			"usuario":        nombreUsuario[pqr.IdUsuario],
			"fecha_creacion": pqr.FechaCreacion,
		})
	}

	c.Data["json"] = lista
	c.ServeJSON()
}
