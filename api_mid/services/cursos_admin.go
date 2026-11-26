package services

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// ADMINISTRAR CURSOS (solo admin: el filtro ya lo revisa)
// Crear, editar y retirar cursos (tabla modulo_educativo).
// Retirar no borra: deja el curso con activo = false para no
// perder el avance ni los certificados de los usuarios.
// ============================================================

// Valores que acepta la tabla (ver los CHECK en sql/02_crear_tablas.sql)
var nivelesCurso = map[string]bool{"basico": true, "intermedio": true, "avanzado": true}
var formatosCurso = map[string]bool{"video": true, "articulo": true, "quiz": true}
var categoriasCurso = map[string]bool{"fundamentos": true, "credito": true, "ahorro": true, "deuda": true, "fraude": true}

// CrearCurso guarda un curso nuevo dentro de una escuela.
func CrearCurso(datos models.DatosCursoAdmin) (map[string]interface{}, error) {

	if err := validarCurso(&datos); err != nil {
		return nil, err
	}

	curso := models.ModuloEducativo{
		Slug:   slugDisponible(datos.Titulo, helpers.UrlCrudEducacion+"/modulo_educativo"),
		Activo: true,
	}
	pasarDatosCurso(&curso, datos)

	var creado models.ModuloEducativo
	if err := helpers.Crear(helpers.UrlCrudEducacion+"/modulo_educativo", curso, &creado); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Curso creado", "id": creado.Id}, nil
}

// EditarCurso cambia los datos de un curso que ya existe.
func EditarCurso(id int, datos models.DatosCursoAdmin) (map[string]interface{}, error) {

	curso, err := buscarCursoActivo(id)
	if err != nil {
		return nil, err
	}

	if err := validarCurso(&datos); err != nil {
		return nil, err
	}

	pasarDatosCurso(&curso, datos)

	// El CRUD reemplaza el registro: se manda el curso completo
	if err := helpers.Actualizar(fmt.Sprintf("%s/modulo_educativo/%d", helpers.UrlCrudEducacion, id), curso); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Curso actualizado", "id": id}, nil
}

// RetirarCurso quita el curso del catalogo (activo = false).
func RetirarCurso(id int) (map[string]interface{}, error) {

	curso, err := buscarCursoActivo(id)
	if err != nil {
		return nil, err
	}

	curso.Activo = false

	if err := helpers.Actualizar(fmt.Sprintf("%s/modulo_educativo/%d", helpers.UrlCrudEducacion, id), curso); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Curso retirado", "id": id}, nil
}

// buscarCursoActivo trae el curso o responde 404 si no existe o ya se retiro.
func buscarCursoActivo(id int) (models.ModuloEducativo, error) {

	curso, err := buscarCurso(id)
	if err != nil {
		return curso, err
	}

	if !curso.Activo {
		return curso, helpers.NuevoError(http.StatusNotFound, "El curso no existe")
	}

	return curso, nil
}

// validarCurso limpia los textos, revisa los valores fijos y que la escuela exista.
func validarCurso(datos *models.DatosCursoAdmin) error {

	datos.Titulo = strings.TrimSpace(datos.Titulo)
	datos.Descripcion = strings.TrimSpace(datos.Descripcion)
	datos.Duracion = strings.TrimSpace(datos.Duracion)
	datos.Categoria = strings.ToLower(strings.TrimSpace(datos.Categoria))
	datos.Nivel = strings.ToLower(strings.TrimSpace(datos.Nivel))
	datos.Formato = strings.ToLower(strings.TrimSpace(datos.Formato))

	if utf8.RuneCountInString(datos.Titulo) < 3 {
		return helpers.NuevoError(http.StatusBadRequest, "El titulo debe tener minimo 3 letras")
	}

	if !nivelesCurso[datos.Nivel] {
		return helpers.NuevoError(http.StatusBadRequest, "El nivel debe ser basico, intermedio o avanzado")
	}

	if !formatosCurso[datos.Formato] {
		return helpers.NuevoError(http.StatusBadRequest, "El formato debe ser video, articulo o quiz")
	}

	if !categoriasCurso[datos.Categoria] {
		return helpers.NuevoError(http.StatusBadRequest, "La categoria debe ser fundamentos, credito, ahorro, deuda o fraude")
	}

	var escuela models.Escuela
	if err := helpers.ConsultarUno(fmt.Sprintf("%s/escuela/%d", helpers.UrlCrudEducacion, datos.IdEscuela), &escuela, "La escuela no existe"); err != nil {
		return err
	}

	if escuela.Id == 0 {
		return helpers.NuevoError(http.StatusNotFound, "La escuela no existe")
	}

	return nil
}

// pasarDatosCurso copia lo que llena el admin en el curso.
func pasarDatosCurso(curso *models.ModuloEducativo, datos models.DatosCursoAdmin) {
	curso.IdEscuela = &models.Ref{Id: datos.IdEscuela}
	curso.Titulo = datos.Titulo
	curso.Descripcion = datos.Descripcion
	curso.Categoria = datos.Categoria
	curso.Nivel = datos.Nivel
	curso.Formato = datos.Formato
	curso.Duracion = datos.Duracion
}
