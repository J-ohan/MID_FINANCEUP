package services

import (
	"fmt"
	"net/http"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// buscarCurso trae un curso (modulo_educativo) por su id.
func buscarCurso(id int) (models.ModuloEducativo, error) {
	var curso models.ModuloEducativo
	err := helpers.ConsultarUno(fmt.Sprintf("%s/modulo_educativo/%d", helpers.UrlCrudEducacion, id), &curso, "El curso no existe")
	return curso, err
}

// leccionesCompletadas devuelve un mapa con las lecciones que el usuario ya completo.
func leccionesCompletadas(idUsuario int) (map[int]bool, []models.ProgresoLeccion, error) {

	var avances []models.ProgresoLeccion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,completado:true&limit=10000", helpers.UrlCrudEducacion, idUsuario), &avances); err != nil {
		return nil, nil, err
	}

	completadas := map[int]bool{}
	for _, avance := range avances {
		completadas[avance.IdLeccion.Id] = true
	}

	return completadas, avances, nil
}

// progresoDelCurso trae el registro de progreso_educativo (si existe).
func progresoDelCurso(idUsuario int, idCurso int) (models.ProgresoEducativo, bool) {

	var progresos []models.ProgresoEducativo

	helpers.ConsultarLista(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d,id_modulo:%d&limit=1", helpers.UrlCrudEducacion, idUsuario, idCurso), &progresos)

	if len(progresos) == 0 {
		return models.ProgresoEducativo{IdUsuario: idUsuario, IdModulo: &models.Ref{Id: idCurso}, Activo: true}, false
	}

	return progresos[0], true
}

// guardar hace POST si el registro es nuevo (Id = 0) o PUT si ya existe.
func guardar(tabla string, id int, registro interface{}) error {

	if id == 0 {
		return helpers.Crear(helpers.UrlCrudEducacion+"/"+tabla, registro, nil)
	}

	return helpers.Actualizar(fmt.Sprintf("%s/%s/%d", helpers.UrlCrudEducacion, tabla, id), registro)
}

// datosCurso es como la pagina muestra un curso en el catalogo.
func datosCurso(curso models.ModuloEducativo) map[string]interface{} {
	return map[string]interface{}{
		"id":                curso.Id,
		"slug":              curso.Slug,
		"titulo":            curso.Titulo,
		"imagen":            curso.UrlThumbnail,
		"categoria":         curso.Categoria,
		"nivel":             curso.Nivel,
		"formato":           curso.Formato,
		"duracion":          curso.Duracion,
		"resumen_lecciones": curso.ResumenLecciones,
		"certificado":       curso.OtorgaCertificado,
		"descripcion":       curso.Descripcion,
		"estudiantes":       curso.Estudiantes,
		"calificacion":      curso.Calificacion,
	}
}

// ------------------------------------------------------------
// Funciones que usan los controladores
// ------------------------------------------------------------

// Catalogo: escuelas con sus cursos e instructor.
func Catalogo() ([]map[string]interface{}, error) {

	// 1. Escuelas, cursos e instructores
	var escuelas []models.Escuela
	var cursos []models.ModuloEducativo
	var instructores []models.Instructor

	if err := helpers.ConsultarLista(helpers.UrlCrudEducacion+"/escuela?query=activo:true&sortby=Orden&order=asc&limit=1000", &escuelas); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudEducacion+"/modulo_educativo?query=activo:true&limit=1000", &cursos); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudEducacion+"/instructor?limit=1000", &instructores); err != nil {
		return nil, err
	}

	instructorPorId := map[int]string{}
	for _, instructor := range instructores {
		instructorPorId[instructor.Id] = instructor.Nombre
	}

	// 2. Meter cada curso dentro de su escuela
	cursosPorEscuela := map[int][]map[string]interface{}{}

	for _, curso := range cursos {

		if curso.IdEscuela == nil {
			continue
		}

		datos := datosCurso(curso)
		datos["instructor"] = ""
		if curso.IdInstructor != nil {
			datos["instructor"] = instructorPorId[curso.IdInstructor.Id]
		}

		cursosPorEscuela[curso.IdEscuela.Id] = append(cursosPorEscuela[curso.IdEscuela.Id], datos)
	}

	// 3. Lista de escuelas
	lista := []map[string]interface{}{}

	for _, escuela := range escuelas {

		cursosEscuela := cursosPorEscuela[escuela.Id]
		if cursosEscuela == nil {
			cursosEscuela = []map[string]interface{}{}
		}

		lista = append(lista, map[string]interface{}{
			"id":          escuela.Id,
			"slug":        escuela.Slug,
			"nombre":      escuela.Nombre,
			"descripcion": escuela.Descripcion,
			"cursos":      cursosEscuela,
		})
	}

	return lista, nil
}

// DetalleCurso: instructor, "lo que aprenderas", requisitos y lecciones.
// Si idUsuario > 0 marca las lecciones completadas y calcula el %.
func DetalleCurso(id int, idUsuario int) (map[string]interface{}, error) {

	// 1. El curso y su instructor
	curso, err := buscarCurso(id)

	if err != nil {
		return nil, err
	}

	instructor := models.Instructor{}
	if curso.IdInstructor != nil {
		helpers.ConsultarUno(fmt.Sprintf("%s/instructor/%d", helpers.UrlCrudEducacion, curso.IdInstructor.Id), &instructor, "")
	}

	// 2. "Lo que aprenderas" y requisitos
	var detalles []models.ModuloDetalle

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/modulo_detalle?query=id_modulo:%d,activo:true&sortby=Orden&order=asc&limit=1000", helpers.UrlCrudEducacion, id), &detalles); err != nil {
		return nil, err
	}

	aprenderas := []string{}
	requisitos := []string{}

	for _, detalle := range detalles {
		if detalle.Tipo == "aprendizaje" {
			aprenderas = append(aprenderas, detalle.Texto)
		} else {
			requisitos = append(requisitos, detalle.Texto)
		}
	}

	// 3. Lecciones en orden
	var lecciones []models.Leccion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true&sortby=NumeroLeccion&order=asc&limit=1000", helpers.UrlCrudEducacion, id), &lecciones); err != nil {
		return nil, err
	}

	// 4. Lecciones completadas por el usuario (si llego)
	completadas := map[int]bool{}

	if idUsuario > 0 {
		completadas, _, _ = leccionesCompletadas(idUsuario)
	}

	listaLecciones := []map[string]interface{}{}
	cuantas := 0

	for _, leccion := range lecciones {

		if completadas[leccion.Id] {
			cuantas++
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
		progreso = cuantas * 100 / len(lecciones)
	}

	// 5. Armar el curso completo
	datos := datosCurso(curso)
	datos["actualizado"] = curso.FechaModificacion.Format("2006-01-02")
	datos["instructor"] = map[string]interface{}{"nombre": instructor.Nombre, "cargo": instructor.Cargo, "iniciales": instructor.Iniciales}
	datos["aprenderas"] = aprenderas
	datos["requisitos"] = requisitos
	datos["lecciones"] = listaLecciones
	datos["progreso"] = progreso

	return datos, nil
}

// ProgresoEducativo: % de cada curso, cuantos completados y % general.
func ProgresoEducativo(idUsuario int) (map[string]interface{}, error) {

	// 1. Revisar que el usuario exista
	if _, err := BuscarUsuario(idUsuario); err != nil {
		return nil, err
	}

	// 2. Cursos y avance del usuario
	var cursos []models.ModuloEducativo
	var avances []models.ProgresoEducativo

	if err := helpers.ConsultarLista(helpers.UrlCrudEducacion+"/modulo_educativo?query=activo:true&limit=1000", &cursos); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/progreso_educativo?query=id_usuario:%d&limit=1000", helpers.UrlCrudEducacion, idUsuario), &avances); err != nil {
		return nil, err
	}

	avancePorCurso := map[int]models.ProgresoEducativo{}
	for _, avance := range avances {
		if avance.IdModulo != nil {
			avancePorCurso[avance.IdModulo.Id] = avance
		}
	}

	// 3. Unir cada curso con su avance
	lista := []map[string]interface{}{}
	completados := 0
	suma := 0

	for _, curso := range cursos {

		avance := avancePorCurso[curso.Id]
		porcentaje := avance.PorcentajeCompletado

		estado := "sin iniciar"
		if porcentaje >= 100 {
			estado = "completado"
			completados++
		} else if porcentaje > 0 {
			estado = "en curso"
		}

		suma += porcentaje

		lista = append(lista, map[string]interface{}{
			"id_modulo":    curso.Id,
			"titulo":       curso.Titulo,
			"nivel":        curso.Nivel,
			"porcentaje":   porcentaje,
			"estado":       estado,
			"calificacion": avance.Calificacion,
		})
	}

	general := 0
	if len(cursos) > 0 {
		general = suma / len(cursos)
	}

	return map[string]interface{}{
		"id_usuario":          idUsuario,
		"modulos_totales":     len(cursos),
		"modulos_completados": completados,
		"porcentaje_general":  general,
		"modulos":             lista,
	}, nil
}

// Clase: una leccion con su contenido y cual es la siguiente.
func Clase(id int) (map[string]interface{}, error) {

	// 1. La leccion
	var leccion models.Leccion

	if err := helpers.ConsultarUno(fmt.Sprintf("%s/leccion/%d", helpers.UrlCrudEducacion, id), &leccion, "La leccion no existe"); err != nil {
		return nil, err
	}

	// 2. Su contenido (si tiene)
	contenido := models.Contenido{}
	if leccion.IdContenido != nil {
		helpers.ConsultarUno(fmt.Sprintf("%s/contenido/%d", helpers.UrlCrudEducacion, leccion.IdContenido.Id), &contenido, "")
	}

	// 3. La siguiente leccion del mismo curso
	var siguientes []models.Leccion
	helpers.ConsultarLista(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true,numero_leccion__gt:%d&sortby=NumeroLeccion&order=asc&limit=1", helpers.UrlCrudEducacion, leccion.IdModulo.Id, leccion.NumeroLeccion), &siguientes)

	idSiguiente := 0
	if len(siguientes) > 0 {
		idSiguiente = siguientes[0].Id
	}

	urlVideo := leccion.UrlVideo
	if urlVideo == "" {
		urlVideo = contenido.UrlVideo
	}

	return map[string]interface{}{
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
	}, nil
}

// CompletarLeccion: marca la leccion y recalcula el % del curso.
func CompletarLeccion(idLeccion int, idUsuario int) (map[string]interface{}, error) {

	if idUsuario == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Falta el id_usuario")
	}

	// 1. La leccion (para saber de que curso es)
	var leccion models.Leccion

	if err := helpers.ConsultarUno(fmt.Sprintf("%s/leccion/%d", helpers.UrlCrudEducacion, idLeccion), &leccion, "La leccion no existe"); err != nil {
		return nil, err
	}

	idCurso := leccion.IdModulo.Id
	ahora := time.Now().UTC()

	// 2. Guardar la leccion como completada (se crea o se edita)
	var avances []models.ProgresoLeccion
	helpers.ConsultarLista(fmt.Sprintf("%s/progreso_leccion?query=id_usuario:%d,id_leccion:%d&limit=1", helpers.UrlCrudEducacion, idUsuario, idLeccion), &avances)

	avance := models.ProgresoLeccion{IdUsuario: idUsuario, IdLeccion: &models.Ref{Id: idLeccion}, Activo: true}
	if len(avances) > 0 {
		avance = avances[0]
	}

	avance.Completado = true
	avance.FechaCompletado = ahora

	if err := guardar("progreso_leccion", avance.Id, avance); err != nil {
		return nil, err
	}

	// 3. Contar las lecciones del curso y las completadas
	var lecciones []models.Leccion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/leccion?query=id_modulo:%d,activo:true&limit=1000", helpers.UrlCrudEducacion, idCurso), &lecciones); err != nil {
		return nil, err
	}

	completadas, _, err := leccionesCompletadas(idUsuario)

	if err != nil {
		return nil, err
	}

	cuantas := 0
	for _, l := range lecciones {
		if completadas[l.Id] {
			cuantas++
		}
	}

	porcentaje := 0
	if len(lecciones) > 0 {
		porcentaje = cuantas * 100 / len(lecciones)
	}

	// 4. Guardar el % del curso
	progreso, _ := progresoDelCurso(idUsuario, idCurso)
	progreso.PorcentajeCompletado = porcentaje

	if porcentaje == 100 {
		progreso.FechaCompletado = ahora
	}

	guardar("progreso_educativo", progreso.Id, progreso)

	return map[string]interface{}{
		"id_leccion":            idLeccion,
		"id_curso":              idCurso,
		"lecciones_completadas": cuantas,
		"lecciones_totales":     len(lecciones),
		"porcentaje":            porcentaje,
	}, nil
}

// ReiniciarCurso: deja todas las lecciones del curso sin completar y el % en 0.
func ReiniciarCurso(idCurso int, idUsuario int) (map[string]interface{}, error) {

	if idUsuario == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Falta el id_usuario")
	}

	// 1. Lecciones del curso
	var lecciones []models.Leccion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/leccion?query=id_modulo:%d&limit=1000", helpers.UrlCrudEducacion, idCurso), &lecciones); err != nil {
		return nil, err
	}

	esDelCurso := map[int]bool{}
	for _, leccion := range lecciones {
		esDelCurso[leccion.Id] = true
	}

	// 2. Marcar como NO completadas las de este curso
	_, avances, err := leccionesCompletadas(idUsuario)

	if err != nil {
		return nil, err
	}

	for _, avance := range avances {
		if esDelCurso[avance.IdLeccion.Id] {
			avance.Completado = false
			avance.FechaCompletado = time.Time{}
			guardar("progreso_leccion", avance.Id, avance)
		}
	}

	// 3. Dejar el % del curso en 0
	if progreso, existe := progresoDelCurso(idUsuario, idCurso); existe {
		progreso.PorcentajeCompletado = 0
		progreso.FechaCompletado = time.Time{}
		guardar("progreso_educativo", progreso.Id, progreso)
	}

	return map[string]interface{}{"id_curso": idCurso, "porcentaje": 0, "mensaje": "Curso reiniciado"}, nil
}

// EmitirCertificado: solo si el curso da certificado y esta al 100%.
// Si ya tenia uno, devuelve el mismo.
func EmitirCertificado(datos models.DatosUsuarioCurso) (map[string]interface{}, error) {

	if datos.IdUsuario == 0 || datos.IdModulo == 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Faltan el id_usuario o el id_modulo")
	}

	// 1. Usuario y curso
	usuario, err := BuscarUsuario(datos.IdUsuario)

	if err != nil {
		return nil, err
	}

	curso, err := buscarCurso(datos.IdModulo)

	if err != nil {
		return nil, err
	}

	if !curso.OtorgaCertificado {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Este curso no entrega certificado")
	}

	// 2. Revisar que este al 100%
	progreso, existe := progresoDelCurso(datos.IdUsuario, datos.IdModulo)

	if !existe || progreso.PorcentajeCompletado < 100 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Debes completar el 100% del curso para recibir el certificado")
	}

	// 3. Buscar si ya tiene certificado; si no, crearlo
	var certificados []models.Certificado

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/certificado?query=id_usuario:%d,id_modulo:%d&limit=1", helpers.UrlCrudEducacion, datos.IdUsuario, datos.IdModulo), &certificados); err != nil {
		return nil, err
	}

	var certificado models.Certificado

	if len(certificados) > 0 {

		certificado = certificados[0]

	} else {

		// Codigo: FUP-curso-usuario-fecha, ej: FUP-3-5-20261013
		certificado = models.Certificado{
			IdUsuario:          datos.IdUsuario,
			IdModulo:           &models.Ref{Id: datos.IdModulo},
			CodigoVerificacion: fmt.Sprintf("FUP-%d-%d-%s", datos.IdModulo, datos.IdUsuario, time.Now().Format("20060102")),
			Activo:             true,
		}

		if err := helpers.Crear(helpers.UrlCrudEducacion+"/certificado", certificado, &certificado); err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"codigo":        certificado.CodigoVerificacion,
		"estudiante":    usuario.Nombre + " " + usuario.Apellido,
		"curso":         curso.Titulo,
		"duracion":      curso.Duracion,
		"fecha_emision": certificado.FechaEmision.Format("2006-01-02"),
	}, nil
}
