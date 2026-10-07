package services

import (
	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ProgresoEducativo devuelve cada modulo con el avance del usuario.
func ProgresoEducativo(idUsuario int) (models.ProgresoGeneral, error) {
	progreso := models.ProgresoGeneral{IdUsuario: idUsuario, Modulos: []models.ModuloVista{}}
	if err := ExisteUsuario(idUsuario); err != nil {
		return progreso, err
	}

	var modulos []models.ModuloEducativo
	if err := helpers.Consultar("educacion", "modulo_educativo?"+helpers.Filtro("activo", true), &modulos); err != nil {
		return progreso, err
	}
	var lecciones []models.Leccion
	if err := helpers.Consultar("educacion", "leccion?"+helpers.Filtro("activo", true), &lecciones); err != nil {
		return progreso, err
	}
	var avances []models.ProgresoEducativo
	if err := helpers.Consultar("educacion", "progreso_educativo?"+helpers.Filtro("id_usuario", idUsuario), &avances); err != nil {
		return progreso, err
	}
	var leccionesVistas []models.ProgresoLeccion
	if err := helpers.Consultar("educacion", "progreso_leccion?"+helpers.Filtro("id_usuario", idUsuario, "completado", true), &leccionesVistas); err != nil {
		return progreso, err
	}

	// A que modulo pertenece cada leccion y cuantas lecciones tiene cada modulo.
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
	avancePorModulo := map[int]models.ProgresoEducativo{}
	for _, a := range avances {
		if a.IdModulo != nil {
			avancePorModulo[a.IdModulo.Id] = a
		}
	}

	suma := 0
	for _, m := range modulos {
		vista := models.ModuloVista{
			IdModulo:             m.Id,
			Titulo:               m.Titulo,
			Descripcion:          m.Descripcion,
			Nivel:                m.Nivel,
			Imagen:               m.UrlThumbnail,
			LeccionesTotales:     totalLecciones[m.Id],
			LeccionesCompletadas: completadas[m.Id],
			Estado:               "sin iniciar",
		}

		// Si hay un registro de progreso se usa ese porcentaje;
		// si no, se calcula con las lecciones completadas.
		if avance, ok := avancePorModulo[m.Id]; ok {
			vista.Porcentaje = avance.PorcentajeCompletado
			if avance.Calificacion > 0 {
				calificacion := avance.Calificacion
				vista.Calificacion = &calificacion
			}
		} else if vista.LeccionesTotales > 0 {
			vista.Porcentaje = vista.LeccionesCompletadas * 100 / vista.LeccionesTotales
		}

		switch {
		case vista.Porcentaje >= 100:
			vista.Estado = "completado"
			progreso.ModulosCompletados++
		case vista.Porcentaje > 0:
			vista.Estado = "en curso"
		}
		suma += vista.Porcentaje
		progreso.Modulos = append(progreso.Modulos, vista)
	}

	progreso.ModulosTotales = len(modulos)
	if len(modulos) > 0 {
		progreso.PorcentajeGeneral = redondear(float64(suma) / float64(len(modulos)))
	}
	return progreso, nil
}
