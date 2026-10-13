package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

type Escuela struct {
	Id          int    `json:"Id"`
	Slug        string `json:"Slug"`
	Nombre      string `json:"Nombre"`
	Descripcion string `json:"Descripcion"`
	Orden       int    `json:"Orden"`
}

type Instructor struct {
	Id        int    `json:"Id"`
	Nombre    string `json:"Nombre"`
	Cargo     string `json:"Cargo"`
	Iniciales string `json:"Iniciales"`
}

type ModuloEducativo struct {
	Id                int       `json:"Id"`
	IdEscuela         *Ref      `json:"IdEscuela"`
	IdInstructor      *Ref      `json:"IdInstructor"`
	Slug              string    `json:"Slug"`
	Titulo            string    `json:"Titulo"`
	Descripcion       string    `json:"Descripcion"`
	Nivel             string    `json:"Nivel"`
	Categoria         string    `json:"Categoria"`
	Formato           string    `json:"Formato"`
	Duracion          string    `json:"Duracion"`
	ResumenLecciones  string    `json:"ResumenLecciones"`
	OtorgaCertificado bool      `json:"OtorgaCertificado"`
	Estudiantes       int       `json:"Estudiantes"`
	Calificacion      float64   `json:"Calificacion"`
	UrlThumbnail      string    `json:"UrlThumbnail"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}

type ModuloDetalle struct {
	Id    int    `json:"Id"`
	Tipo  string `json:"Tipo"`
	Texto string `json:"Texto"`
}

type Leccion struct {
	Id              int    `json:"Id"`
	IdModulo        *Ref   `json:"IdModulo"`
	IdContenido     *Ref   `json:"IdContenido"`
	Titulo          string `json:"Titulo"`
	Tipo            string `json:"Tipo"`
	Descripcion     string `json:"Descripcion"`
	DuracionMinutos int    `json:"DuracionMinutos"`
	UrlVideo        string `json:"UrlVideo"`
	NumeroLeccion   int    `json:"NumeroLeccion"`
}

type Contenido struct {
	Id          int    `json:"Id"`
	Titulo      string `json:"Titulo"`
	Descripcion string `json:"Descripcion"`
	UrlVideo    string `json:"UrlVideo"`
}

type ProgresoEducativo struct {
	Id                   int       `json:"Id"`
	IdUsuario            int       `json:"IdUsuario"`
	IdModulo             *Ref      `json:"IdModulo"`
	PorcentajeCompletado int       `json:"PorcentajeCompletado"`
	FechaCompletado      time.Time `json:"FechaCompletado"`
	Calificacion         int       `json:"Calificacion"`
	Activo               bool      `json:"Activo"`
}

type ProgresoLeccion struct {
	Id              int       `json:"Id"`
	IdUsuario       int       `json:"IdUsuario"`
	IdLeccion       *Ref      `json:"IdLeccion"`
	Completado      bool      `json:"Completado"`
	FechaCompletado time.Time `json:"FechaCompletado"`
	Activo          bool      `json:"Activo"`
}

type Certificado struct {
	Id                 int       `json:"Id"`
	IdUsuario          int       `json:"IdUsuario"`
	IdModulo           *Ref      `json:"IdModulo"`
	CodigoVerificacion string    `json:"CodigoVerificacion"`
	FechaEmision       time.Time `json:"FechaEmision"`
	Activo             bool      `json:"Activo"`
}

type DatosUsuarioCurso struct {
	IdUsuario int `json:"id_usuario"`
	IdModulo  int `json:"id_modulo"`
}

// EducacionController operations for Educacion
type EducacionController struct {
	beego.Controller
}

// ============================================================
// GET /v1/cursos
// Catalogo: escuelas con sus cursos e instructor (crud_educacion)
// ============================================================

func (c *EducacionController) GetCursos() {

	// --------------------------------------------------------
	// 1. Consultar las escuelas (ordenadas)
	// --------------------------------------------------------

	responseEscuelas, err := http.Get(urlCrudEducacion + "/escuela?query=activo:true&sortby=Orden&order=asc&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseEscuelas.Body.Close()

	bodyEscuelas, _ := io.ReadAll(responseEscuelas.Body)

	var escuelas []Escuela
	json.Unmarshal(bodyEscuelas, &escuelas)

	// --------------------------------------------------------
	// 2. Consultar los cursos
	// --------------------------------------------------------

	responseModulos, err := http.Get(urlCrudEducacion + "/modulo_educativo?query=activo:true&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los cursos"}
		c.ServeJSON()
		return
	}

	defer responseModulos.Body.Close()

	bodyModulos, _ := io.ReadAll(responseModulos.Body)

	var modulos []ModuloEducativo
	json.Unmarshal(bodyModulos, &modulos)

	// --------------------------------------------------------
	// 3. Consultar los instructores
	// --------------------------------------------------------

	responseInstructores, err := http.Get(urlCrudEducacion + "/instructor?limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los instructores"}
		c.ServeJSON()
		return
	}

	defer responseInstructores.Body.Close()

	bodyInstructores, _ := io.ReadAll(responseInstructores.Body)

	var instructores []Instructor
	json.Unmarshal(bodyInstructores, &instructores)

	instructorPorId := map[int]Instructor{}
	for _, instructor := range instructores {
		instructorPorId[instructor.Id] = instructor
	}

	// --------------------------------------------------------
	// 4. Meter cada curso dentro de su escuela
	// --------------------------------------------------------

	cursosPorEscuela := map[int][]map[string]interface{}{}

	for _, modulo := range modulos {

		if modulo.IdEscuela == nil {
			continue
		}

		instructor := Instructor{}
		if modulo.IdInstructor != nil {
			instructor = instructorPorId[modulo.IdInstructor.Id]
		}

		cursosPorEscuela[modulo.IdEscuela.Id] = append(cursosPorEscuela[modulo.IdEscuela.Id], map[string]interface{}{
			"id":                modulo.Id,
			"slug":              modulo.Slug,
			"titulo":            modulo.Titulo,
			"imagen":            modulo.UrlThumbnail,
			"categoria":         modulo.Categoria,
			"nivel":             modulo.Nivel,
			"formato":           modulo.Formato,
			"duracion":          modulo.Duracion,
			"resumen_lecciones": modulo.ResumenLecciones,
			"certificado":       modulo.OtorgaCertificado,
			"descripcion":       modulo.Descripcion,
			"estudiantes":       modulo.Estudiantes,
			"calificacion":      modulo.Calificacion,
			"instructor":        instructor.Nombre,
		})
	}

	// --------------------------------------------------------
	// 5. Armar la lista de escuelas y devolverla
	// --------------------------------------------------------

	lista := []map[string]interface{}{}

	for _, escuela := range escuelas {

		cursos := cursosPorEscuela[escuela.Id]
		if cursos == nil {
			cursos = []map[string]interface{}{}
		}

		lista = append(lista, map[string]interface{}{
			"id":          escuela.Id,
			"slug":        escuela.Slug,
			"nombre":      escuela.Nombre,
			"descripcion": escuela.Descripcion,
			"cursos":      cursos,
		})
	}

	c.Data["json"] = lista
	c.ServeJSON()
}

// ============================================================
// GET /v1/cursos/:id?usuario=5
// Detalle de un curso: instructor, "lo que aprenderas",
// requisitos y lecciones. Si llega ?usuario=, se marca
// que lecciones ya completo y su porcentaje.
// ============================================================

func (c *EducacionController) GetCurso() {

	// --------------------------------------------------------
	// 1. Obtener el ID del curso y del usuario (opcional)
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	idUsuario, _ := c.GetInt("usuario")

	// --------------------------------------------------------
	// 2. Consultar el curso
	// --------------------------------------------------------

	responseModulo, err := http.Get(fmt.Sprintf("%s/modulo_educativo/%d", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseModulo.Body.Close()

	bodyModulo, _ := io.ReadAll(responseModulo.Body)

	var modulo ModuloEducativo

	if err := json.Unmarshal(bodyModulo, &modulo); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El curso no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar el instructor
	// --------------------------------------------------------

	instructor := Instructor{}

	if modulo.IdInstructor != nil {

		responseInstructor, err := http.Get(fmt.Sprintf("%s/instructor/%d", urlCrudEducacion, modulo.IdInstructor.Id))

		if err == nil {
			defer responseInstructor.Body.Close()
			bodyInstructor, _ := io.ReadAll(responseInstructor.Body)
			json.Unmarshal(bodyInstructor, &instructor)
		}
	}

	// --------------------------------------------------------
	// 4. Consultar "lo que aprenderas" y los requisitos
	// --------------------------------------------------------

	responseDetalles, err := http.Get(fmt.Sprintf("%s/modulo_detalle?query=id_modulo:%d,activo:true&sortby=Orden&order=asc&limit=1000", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el detalle del curso"}
		c.ServeJSON()
		return
	}

	defer responseDetalles.Body.Close()

	bodyDetalles, _ := io.ReadAll(responseDetalles.Body)

	var detalles []ModuloDetalle
	json.Unmarshal(bodyDetalles, &detalles)

	aprenderas := []string{}
	requisitos := []string{}

	for _, detalle := range detalles {
		if detalle.Tipo == "aprendizaje" {
			aprenderas = append(aprenderas, detalle.Texto)
		} else {
			requisitos = append(requisitos, detalle.Texto)
		}
	}

	// --------------------------------------------------------
	// 5. Consultar las lecciones (en orden)
	// --------------------------------------------------------

	responseLecciones, err := http.Get(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true&sortby=NumeroLeccion&order=asc&limit=1000", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las lecciones"}
		c.ServeJSON()
		return
	}

	defer responseLecciones.Body.Close()

	bodyLecciones, _ := io.ReadAll(responseLecciones.Body)

	var lecciones []Leccion
	json.Unmarshal(bodyLecciones, &lecciones)

	// --------------------------------------------------------
	// 6. Si llego un usuario, ver que lecciones ya completo
	// --------------------------------------------------------

	completadas := map[int]bool{}

	if idUsuario > 0 {

		responseAvance, err := http.Get(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,completado:true&limit=10000", urlCrudEducacion, idUsuario))

		if err == nil {
			defer responseAvance.Body.Close()
			bodyAvance, _ := io.ReadAll(responseAvance.Body)

			var avances []ProgresoLeccion
			json.Unmarshal(bodyAvance, &avances)

			for _, avance := range avances {
				completadas[avance.IdLeccion.Id] = true
			}
		}
	}

	// --------------------------------------------------------
	// 7. Armar la lista de lecciones y calcular el progreso
	// --------------------------------------------------------

	listaLecciones := []map[string]interface{}{}
	cuantasCompletadas := 0

	for _, leccion := range lecciones {

		if completadas[leccion.Id] {
			cuantasCompletadas++
		}

		listaLecciones = append(listaLecciones, map[string]interface{}{
			"id":           leccion.Id,
			"titulo":       leccion.Titulo,
			"tipo":         leccion.Tipo,
			"duracion_min": leccion.DuracionMinutos,
			"resumen":      leccion.Descripcion,
			"completada":   completadas[leccion.Id],
		})
	}

	progreso := 0
	if len(lecciones) > 0 {
		progreso = cuantasCompletadas * 100 / len(lecciones)
	}

	// --------------------------------------------------------
	// 8. Devolver el curso completo
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id":                modulo.Id,
		"slug":              modulo.Slug,
		"titulo":            modulo.Titulo,
		"imagen":            modulo.UrlThumbnail,
		"categoria":         modulo.Categoria,
		"nivel":             modulo.Nivel,
		"formato":           modulo.Formato,
		"duracion":          modulo.Duracion,
		"resumen_lecciones": modulo.ResumenLecciones,
		"certificado":       modulo.OtorgaCertificado,
		"descripcion":       modulo.Descripcion,
		"estudiantes":       modulo.Estudiantes,
		"calificacion":      modulo.Calificacion,
		"actualizado":       modulo.FechaModificacion.Format("2006-01-02"),
		"instructor": map[string]interface{}{
			"nombre":    instructor.Nombre,
			"cargo":     instructor.Cargo,
			"iniciales": instructor.Iniciales,
		},
		"aprenderas": aprenderas,
		"requisitos": requisitos,
		"lecciones":  listaLecciones,
		"progreso":   progreso,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/progreso-educativo/:id
// Avance del usuario en cada curso (crud_auth + crud_educacion)
// ============================================================

func (c *EducacionController) GetProgresoEducativo() {

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
	// 2. Revisar que el usuario exista (crud_auth)
	// --------------------------------------------------------

	responseUsuario, err := http.Get(fmt.Sprintf("%s/usuario/%d", urlCrudAuth, id))

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
	// 3. Consultar todos los cursos (modulos)
	// --------------------------------------------------------

	responseModulos, err := http.Get(urlCrudEducacion + "/modulo_educativo?query=activo:true&limit=1000")

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseModulos.Body.Close()

	bodyModulos, _ := io.ReadAll(responseModulos.Body)

	var modulos []ModuloEducativo
	json.Unmarshal(bodyModulos, &modulos)

	// --------------------------------------------------------
	// 4. Consultar el avance del usuario en cada curso
	// --------------------------------------------------------

	responseProgreso, err := http.Get(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d&limit=1000", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseProgreso.Body.Close()

	bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

	var avances []ProgresoEducativo
	json.Unmarshal(bodyProgreso, &avances)

	// Guardamos el avance de cada curso por su id
	avancePorModulo := map[int]ProgresoEducativo{}
	for _, avance := range avances {
		if avance.IdModulo != nil {
			avancePorModulo[avance.IdModulo.Id] = avance
		}
	}

	// --------------------------------------------------------
	// 5. Unir cada curso con su avance
	// --------------------------------------------------------

	listaModulos := []map[string]interface{}{}
	completados := 0
	sumaPorcentajes := 0

	for _, modulo := range modulos {

		avance := avancePorModulo[modulo.Id]
		porcentaje := avance.PorcentajeCompletado

		estado := "sin iniciar"
		if porcentaje >= 100 {
			estado = "completado"
			completados++
		} else if porcentaje > 0 {
			estado = "en curso"
		}

		sumaPorcentajes += porcentaje

		listaModulos = append(listaModulos, map[string]interface{}{
			"id_modulo":    modulo.Id,
			"titulo":       modulo.Titulo,
			"nivel":        modulo.Nivel,
			"porcentaje":   porcentaje,
			"estado":       estado,
			"calificacion": avance.Calificacion,
		})
	}

	porcentajeGeneral := 0
	if len(modulos) > 0 {
		porcentajeGeneral = sumaPorcentajes / len(modulos)
	}

	// --------------------------------------------------------
	// 6. Devolver el progreso
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_usuario":          usuario.Id,
		"modulos_totales":     len(modulos),
		"modulos_completados": completados,
		"porcentaje_general":  porcentajeGeneral,
		"modulos":             listaModulos,
	}

	c.ServeJSON()
}

// ============================================================
// GET /v1/clase/:id
// Una leccion con su contenido y cual es la siguiente
// ============================================================

func (c *EducacionController) GetClase() {

	// --------------------------------------------------------
	// 1. Obtener el ID de la leccion
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar la leccion
	// --------------------------------------------------------

	responseLeccion, err := http.Get(fmt.Sprintf("%s/leccion/%d", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseLeccion.Body.Close()

	bodyLeccion, _ := io.ReadAll(responseLeccion.Body)

	var leccion Leccion

	if err := json.Unmarshal(bodyLeccion, &leccion); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "La leccion no existe"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 3. Consultar su contenido (si tiene)
	// --------------------------------------------------------

	contenido := Contenido{}

	if leccion.IdContenido != nil {

		responseContenido, err := http.Get(fmt.Sprintf("%s/contenido/%d", urlCrudEducacion, leccion.IdContenido.Id))

		if err == nil {
			defer responseContenido.Body.Close()
			bodyContenido, _ := io.ReadAll(responseContenido.Body)
			json.Unmarshal(bodyContenido, &contenido)
		}
	}

	// --------------------------------------------------------
	// 4. Buscar la siguiente leccion del mismo curso
	// --------------------------------------------------------

	responseSiguiente, err := http.Get(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true,numero_leccion__gt:%d&sortby=NumeroLeccion&order=asc&limit=1", urlCrudEducacion, leccion.IdModulo.Id, leccion.NumeroLeccion))

	idSiguiente := 0

	if err == nil {
		defer responseSiguiente.Body.Close()
		bodySiguiente, _ := io.ReadAll(responseSiguiente.Body)

		var siguientes []Leccion
		json.Unmarshal(bodySiguiente, &siguientes)

		if len(siguientes) > 0 {
			idSiguiente = siguientes[0].Id
		}
	}

	// --------------------------------------------------------
	// 5. Devolver la clase
	// --------------------------------------------------------

	urlVideo := leccion.UrlVideo
	if urlVideo == "" {
		urlVideo = contenido.UrlVideo
	}

	c.Data["json"] = map[string]interface{}{
		"id":                 leccion.Id,
		"id_curso":           leccion.IdModulo.Id,
		"numero":             leccion.NumeroLeccion,
		"titulo":             leccion.Titulo,
		"tipo":               leccion.Tipo,
		"duracion_min":       leccion.DuracionMinutos,
		"resumen":            leccion.Descripcion,
		"url_video":          urlVideo,
		"contenido":          contenido.Descripcion,
		"id_siguiente_clase": idSiguiente,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/clase/:id/completar
// Marca la leccion como completada y recalcula el % del curso
// ============================================================

func (c *EducacionController) PostCompletarLeccion() {

	// --------------------------------------------------------
	// 1. Obtener el ID de la leccion y el usuario
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	var datos DatosUsuarioCurso

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil || datos.IdUsuario == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Falta el id_usuario"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar la leccion (para saber de que curso es)
	// --------------------------------------------------------

	responseLeccion, err := http.Get(fmt.Sprintf("%s/leccion/%d", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseLeccion.Body.Close()

	bodyLeccion, _ := io.ReadAll(responseLeccion.Body)

	var leccion Leccion

	if err := json.Unmarshal(bodyLeccion, &leccion); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "La leccion no existe"}
		c.ServeJSON()
		return
	}

	idModulo := leccion.IdModulo.Id
	ahora := time.Now().UTC()

	// --------------------------------------------------------
	// 3. Guardar la leccion como completada
	//    Si ya habia un registro se edita (PUT), si no se crea (POST)
	// --------------------------------------------------------

	responseAvance, err := http.Get(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,id_leccion:%d&limit=1", urlCrudEducacion, datos.IdUsuario, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseAvance.Body.Close()

	bodyAvance, _ := io.ReadAll(responseAvance.Body)

	var avances []ProgresoLeccion
	json.Unmarshal(bodyAvance, &avances)

	avance := ProgresoLeccion{IdUsuario: datos.IdUsuario, IdLeccion: &Ref{Id: id}, Activo: true}

	if len(avances) > 0 {
		avance = avances[0]
	}

	avance.Completado = true
	avance.FechaCompletado = ahora

	jsonAvance, _ := json.Marshal(avance)

	metodo := http.MethodPost
	direccion := urlCrudEducacion + "/progreso_leccion"

	if avance.Id > 0 {
		metodo = http.MethodPut
		direccion = fmt.Sprintf("%s/progreso_leccion/%d", urlCrudEducacion, avance.Id)
	}

	peticionAvance, _ := http.NewRequest(metodo, direccion, bytes.NewBuffer(jsonAvance))
	peticionAvance.Header.Set("Content-Type", "application/json")

	responseGuardar, err := http.DefaultClient.Do(peticionAvance)

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No se pudo guardar el avance"}
		c.ServeJSON()
		return
	}

	responseGuardar.Body.Close()

	// --------------------------------------------------------
	// 4. Contar las lecciones del curso y las completadas
	// --------------------------------------------------------

	responseLecciones, err := http.Get(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true&limit=1000", urlCrudEducacion, idModulo))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar las lecciones"}
		c.ServeJSON()
		return
	}

	defer responseLecciones.Body.Close()

	bodyLecciones, _ := io.ReadAll(responseLecciones.Body)

	var lecciones []Leccion
	json.Unmarshal(bodyLecciones, &lecciones)

	responseCompletadas, err := http.Get(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,completado:true&limit=10000", urlCrudEducacion, datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseCompletadas.Body.Close()

	bodyCompletadas, _ := io.ReadAll(responseCompletadas.Body)

	var completadas []ProgresoLeccion
	json.Unmarshal(bodyCompletadas, &completadas)

	leccionCompletada := map[int]bool{}
	for _, completada := range completadas {
		leccionCompletada[completada.IdLeccion.Id] = true
	}

	cuantas := 0
	for _, l := range lecciones {
		if leccionCompletada[l.Id] {
			cuantas++
		}
	}

	porcentaje := 0
	if len(lecciones) > 0 {
		porcentaje = cuantas * 100 / len(lecciones)
	}

	// --------------------------------------------------------
	// 5. Guardar el porcentaje del curso (progreso_educativo)
	// --------------------------------------------------------

	responseProgreso, err := http.Get(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d,id_modulo:%d&limit=1", urlCrudEducacion, datos.IdUsuario, idModulo))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso del curso"}
		c.ServeJSON()
		return
	}

	defer responseProgreso.Body.Close()

	bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

	var progresos []ProgresoEducativo
	json.Unmarshal(bodyProgreso, &progresos)

	progreso := ProgresoEducativo{IdUsuario: datos.IdUsuario, IdModulo: &Ref{Id: idModulo}, Activo: true}

	if len(progresos) > 0 {
		progreso = progresos[0]
	}

	progreso.PorcentajeCompletado = porcentaje

	if porcentaje == 100 {
		progreso.FechaCompletado = ahora
	}

	jsonProgreso, _ := json.Marshal(progreso)

	metodo = http.MethodPost
	direccion = urlCrudEducacion + "/progreso_educativo"

	if progreso.Id > 0 {
		metodo = http.MethodPut
		direccion = fmt.Sprintf("%s/progreso_educativo/%d", urlCrudEducacion, progreso.Id)
	}

	peticionProgreso, _ := http.NewRequest(metodo, direccion, bytes.NewBuffer(jsonProgreso))
	peticionProgreso.Header.Set("Content-Type", "application/json")

	if responsePut, err := http.DefaultClient.Do(peticionProgreso); err == nil {
		responsePut.Body.Close()
	}

	// --------------------------------------------------------
	// 6. Devolver el nuevo avance
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"id_leccion":            id,
		"id_curso":              idModulo,
		"lecciones_completadas": cuantas,
		"lecciones_totales":     len(lecciones),
		"porcentaje":            porcentaje,
	}

	c.ServeJSON()
}

// ============================================================
// POST /v1/cursos/:id/reiniciar
// Deja el curso en 0%: todas sus lecciones sin completar
// ============================================================

func (c *EducacionController) PostReiniciarCurso() {

	// --------------------------------------------------------
	// 1. Obtener el ID del curso y el usuario
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	var datos DatosUsuarioCurso

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil || datos.IdUsuario == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Falta el id_usuario"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar las lecciones del curso
	// --------------------------------------------------------

	responseLecciones, err := http.Get(fmt.Sprintf("%s/leccion?query=id_modulo:%d&limit=1000", urlCrudEducacion, id))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseLecciones.Body.Close()

	bodyLecciones, _ := io.ReadAll(responseLecciones.Body)

	var lecciones []Leccion
	json.Unmarshal(bodyLecciones, &lecciones)

	esDelCurso := map[int]bool{}
	for _, leccion := range lecciones {
		esDelCurso[leccion.Id] = true
	}

	// --------------------------------------------------------
	// 3. Marcar como NO completadas las lecciones de ese curso
	// --------------------------------------------------------

	responseAvances, err := http.Get(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,completado:true&limit=10000", urlCrudEducacion, datos.IdUsuario))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseAvances.Body.Close()

	bodyAvances, _ := io.ReadAll(responseAvances.Body)

	var avances []ProgresoLeccion
	json.Unmarshal(bodyAvances, &avances)

	for _, avance := range avances {

		if !esDelCurso[avance.IdLeccion.Id] {
			continue
		}

		avance.Completado = false
		avance.FechaCompletado = time.Time{}

		jsonAvance, _ := json.Marshal(avance)

		peticion, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/progreso_leccion/%d", urlCrudEducacion, avance.Id), bytes.NewBuffer(jsonAvance))
		peticion.Header.Set("Content-Type", "application/json")

		if responsePut, err := http.DefaultClient.Do(peticion); err == nil {
			responsePut.Body.Close()
		}
	}

	// --------------------------------------------------------
	// 4. Dejar el porcentaje del curso en 0
	// --------------------------------------------------------

	responseProgreso, err := http.Get(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d,id_modulo:%d&limit=1", urlCrudEducacion, datos.IdUsuario, id))

	if err == nil {
		defer responseProgreso.Body.Close()
		bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

		var progresos []ProgresoEducativo
		json.Unmarshal(bodyProgreso, &progresos)

		if len(progresos) > 0 {
			progreso := progresos[0]
			progreso.PorcentajeCompletado = 0
			progreso.FechaCompletado = time.Time{}

			jsonProgreso, _ := json.Marshal(progreso)

			peticion, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/progreso_educativo/%d", urlCrudEducacion, progreso.Id), bytes.NewBuffer(jsonProgreso))
			peticion.Header.Set("Content-Type", "application/json")

			if responsePut, err := http.DefaultClient.Do(peticion); err == nil {
				responsePut.Body.Close()
			}
		}
	}

	// --------------------------------------------------------
	// 5. Responder
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{"id_curso": id, "porcentaje": 0, "mensaje": "Curso reiniciado"}
	c.ServeJSON()
}

// ============================================================
// POST /v1/certificado
// Entrega el certificado solo si el curso esta al 100%
// (crud_auth + crud_educacion)
// ============================================================

func (c *EducacionController) PostCertificado() {

	// --------------------------------------------------------
	// 1. Leer los datos enviados
	// --------------------------------------------------------

	var datos DatosUsuarioCurso

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &datos); err != nil || datos.IdUsuario == 0 || datos.IdModulo == 0 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Faltan el id_usuario o el id_modulo"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Consultar el usuario (su nombre va en el certificado)
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
	// 3. Consultar el curso y revisar que de certificado
	// --------------------------------------------------------

	responseModulo, err := http.Get(fmt.Sprintf("%s/modulo_educativo/%d", urlCrudEducacion, datos.IdModulo))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible comunicarse con la API de educacion"}
		c.ServeJSON()
		return
	}

	defer responseModulo.Body.Close()

	bodyModulo, _ := io.ReadAll(responseModulo.Body)

	var modulo ModuloEducativo

	if err := json.Unmarshal(bodyModulo, &modulo); err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El curso no existe"}
		c.ServeJSON()
		return
	}

	if !modulo.OtorgaCertificado {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Este curso no entrega certificado"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 4. Revisar que el curso este al 100%
	// --------------------------------------------------------

	responseProgreso, err := http.Get(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d,id_modulo:%d&limit=1", urlCrudEducacion, datos.IdUsuario, datos.IdModulo))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar el progreso"}
		c.ServeJSON()
		return
	}

	defer responseProgreso.Body.Close()

	bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

	var progresos []ProgresoEducativo
	json.Unmarshal(bodyProgreso, &progresos)

	if len(progresos) == 0 || progresos[0].PorcentajeCompletado < 100 {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "Debes completar el 100% del curso para recibir el certificado"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 5. Si ya tenia certificado se devuelve el mismo;
	//    si no, se crea uno nuevo con su codigo
	// --------------------------------------------------------

	responseCertificado, err := http.Get(fmt.Sprintf("%s/certificado?query=id_usuario:%d,id_modulo:%d&limit=1", urlCrudEducacion, datos.IdUsuario, datos.IdModulo))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		c.Data["json"] = map[string]interface{}{"error": "No fue posible consultar los certificados"}
		c.ServeJSON()
		return
	}

	defer responseCertificado.Body.Close()

	bodyCertificado, _ := io.ReadAll(responseCertificado.Body)

	var certificados []Certificado
	json.Unmarshal(bodyCertificado, &certificados)

	var certificado Certificado

	if len(certificados) > 0 {

		certificado = certificados[0]

	} else {

		// Codigo: FUP-curso-usuario-fecha, ej: FUP-3-5-20261013
		certificado = Certificado{
			IdUsuario:          datos.IdUsuario,
			IdModulo:           &Ref{Id: datos.IdModulo},
			CodigoVerificacion: fmt.Sprintf("FUP-%d-%d-%s", datos.IdModulo, datos.IdUsuario, time.Now().Format("20060102")),
			Activo:             true,
		}

		jsonCertificado, _ := json.Marshal(certificado)

		responseCrear, err := http.Post(urlCrudEducacion+"/certificado", "application/json", bytes.NewBuffer(jsonCertificado))

		if err != nil || responseCrear.StatusCode != http.StatusCreated {
			c.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
			c.Data["json"] = map[string]interface{}{"error": "No se pudo crear el certificado"}
			c.ServeJSON()
			return
		}

		defer responseCrear.Body.Close()

		bodyCrear, _ := io.ReadAll(responseCrear.Body)
		json.Unmarshal(bodyCrear, &certificado)
	}

	// --------------------------------------------------------
	// 6. Devolver los datos del certificado
	// --------------------------------------------------------

	c.Data["json"] = map[string]interface{}{
		"codigo":        certificado.CodigoVerificacion,
		"estudiante":    usuario.Nombre + " " + usuario.Apellido,
		"curso":         modulo.Titulo,
		"duracion":      modulo.Duracion,
		"fecha_emision": certificado.FechaEmision.Format("2006-01-02"),
	}

	c.ServeJSON()
}
