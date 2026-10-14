package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE EDUCACION (crud_educacion)
// Pendiente: cambiar las llamadas por los helpers cuando esten.
// ============================================================

// ProgresoEducativo: avance del usuario en cada curso.
// Devuelve el resultado, el codigo HTTP y el error (si hubo).
func ProgresoEducativo(idUsuario int) (map[string]interface{}, int, error) {

	// 1. Revisar que el usuario exista (crud_auth)
	responseUsuario, err := http.Get(fmt.Sprintf("http://localhost:8081/v1/usuario/%d", idUsuario))

	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("No fue posible comunicarse con la API de usuarios")
	}

	defer responseUsuario.Body.Close()

	bodyUsuario, _ := io.ReadAll(responseUsuario.Body)

	var usuario struct {
		Id int `json:"Id"`
	}

	if err := json.Unmarshal(bodyUsuario, &usuario); err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("El usuario no existe")
	}

	// 2. Consultar todos los cursos (modulos)
	responseModulos, err := http.Get("http://localhost:8083/v1/modulo_educativo?query=activo:true&limit=1000")

	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("No fue posible comunicarse con la API de educacion")
	}

	defer responseModulos.Body.Close()

	bodyModulos, _ := io.ReadAll(responseModulos.Body)

	var modulos []models.ModuloEducativo
	json.Unmarshal(bodyModulos, &modulos)

	// 3. Consultar el avance del usuario en cada curso
	responseProgreso, err := http.Get(fmt.Sprintf("http://localhost:8083/v1/progreso_educativo?query=id_usuario:%d&limit=1000", idUsuario))

	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("No fue posible consultar el progreso")
	}

	defer responseProgreso.Body.Close()

	bodyProgreso, _ := io.ReadAll(responseProgreso.Body)

	var avances []models.ProgresoEducativo
	json.Unmarshal(bodyProgreso, &avances)

	// Guardamos el avance de cada curso por su id
	avancePorModulo := map[int]models.ProgresoEducativo{}
	for _, avance := range avances {
		if avance.IdModulo != nil {
			avancePorModulo[avance.IdModulo.Id] = avance
		}
	}

	// 4. Unir cada curso con su porcentaje y estado
	listaModulos := []map[string]interface{}{}

	for _, modulo := range modulos {

		porcentaje := avancePorModulo[modulo.Id].PorcentajeCompletado

		estado := "sin iniciar"
		if porcentaje >= 100 {
			estado = "completado"
		} else if porcentaje > 0 {
			estado = "en curso"
		}

		listaModulos = append(listaModulos, map[string]interface{}{
			"id_modulo":  modulo.Id,
			"titulo":     modulo.Titulo,
			"nivel":      modulo.Nivel,
			"porcentaje": porcentaje,
			"estado":     estado,
		})
	}

	// 5. Devolver el progreso
	//    Pendiente: porcentaje general y cursos completados
	return map[string]interface{}{
		"id_usuario": usuario.Id,
		"modulos":    listaModulos,
	}, http.StatusOK, nil
}

// Pendiente: Catalogo, DetalleCurso, Clase, CompletarLeccion,
// ReiniciarCurso y EmitirCertificado
