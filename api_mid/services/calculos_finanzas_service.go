package services

import (
	"api_mid_financeup/models"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
)

// ============================================================
// FUNCIONES GENERALES
// ============================================================

// porcentaje calcula qué % es "parte" de "total", redondeado a entero.
func porcentaje(parte, total float64) int {
	if total == 0 {
		return 0
	}

	return int(math.Round(parte / total * 100))
}

// porcentajeDecimal es igual que porcentaje, pero con un decimal.
func porcentajeDecimal(parte, total float64) float64 {
	if total == 0 {
		return 0
	}

	return math.Round(parte/total*1000) / 10
}

var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// Normalizar deja un texto listo para comparar: sin espacios, en minúsculas y sin tildes.
// Ejemplo: " Alimentación " -> "alimentacion"
func Normalizar(texto string) string {
	return sinTildes.Replace(strings.ToLower(strings.TrimSpace(texto)))
}

// Capitalizar pone en mayúscula la primera letra de cada palabra.
// Ejemplo: "harold arciniegas" -> "Harold Arciniegas"
func Capitalizar(texto string) string {
	palabras := strings.Fields(strings.ToLower(texto))

	for i, palabra := range palabras {
		letras := []rune(palabra)
		letras[0] = unicode.ToUpper(letras[0])
		palabras[i] = string(letras)
	}

	return strings.Join(palabras, " ")
}

func NombreCompleto(usuario models.Usuario) string {
	return Capitalizar(usuario.Nombre + " " + usuario.Apellido)
}

// ============================================================
// CATEGORÍAS
// ============================================================

func ArmarCategorias(categorias []models.Categoria) []models.ResultadoCategoria {
	lista := []models.ResultadoCategoria{}

	for _, categoria := range categorias {
		lista = append(lista, models.ResultadoCategoria{
			Id:     categoria.Id,
			Nombre: categoria.Nombre,
			Icono:  categoria.Icono,
		})
	}

	return lista
}

// BuscarCategoria busca una categoría por su nombre, sin importar mayúsculas ni tildes.
func BuscarCategoria(categorias []models.Categoria, nombre string) (models.Categoria, error) {
	for _, categoria := range categorias {
		if Normalizar(categoria.Nombre) == Normalizar(nombre) {
			return categoria, nil
		}
	}

	return models.Categoria{}, fmt.Errorf("La categoría \"%s\" no existe.", nombre)
}

func mapaCategorias(categorias []models.Categoria) map[int]models.Categoria {
	mapa := map[int]models.Categoria{}

	for _, categoria := range categorias {
		mapa[categoria.Id] = categoria
	}

	return mapa
}

// ============================================================
// LIBRO MAYOR (movimientos)
// ============================================================

// NuevoMovimiento pasa la solicitud del frontend al formato de la tabla.
func NuevoMovimiento(idUsuario int, solicitud models.SolicitudMovimiento, categoria models.Categoria) models.MovimientoIngresoEgreso {
	fecha, _ := time.Parse("2006-01-02", solicitud.Fecha)

	movimiento := models.MovimientoIngresoEgreso{
		IdUsuario:   idUsuario,
		IdCategoria: &models.Ref{Id: categoria.Id},
		Nombre:      solicitud.Concepto,
		Monto:       solicitud.Valor,
		EsIngreso:   solicitud.Tipo == "ingreso",
		Fecha:       fecha,
		Activo:      true,
	}

	if solicitud.MetodoPago != nil {
		movimiento.MetodoPago = strings.TrimSpace(*solicitud.MetodoPago)
	}

	if solicitud.Observaciones != nil {
		movimiento.Observaciones = strings.TrimSpace(*solicitud.Observaciones)
	}

	return movimiento
}

// ArmarMovimiento pasa una fila de la tabla al formato del frontend.
func ArmarMovimiento(movimiento models.MovimientoIngresoEgreso, categorias []models.Categoria) models.ResultadoMovimiento {
	return armarMovimiento(movimiento, mapaCategorias(categorias))
}

func armarMovimiento(movimiento models.MovimientoIngresoEgreso, categorias map[int]models.Categoria) models.ResultadoMovimiento {
	resultado := models.ResultadoMovimiento{
		Id:            movimiento.Id,
		IdUsuario:     movimiento.IdUsuario,
		Fecha:         movimiento.Fecha.Format("2006-01-02"),
		Concepto:      movimiento.Nombre,
		Categoria:     "Sin categoria",
		Icono:         "💰",
		Tipo:          "gasto",
		Valor:         movimiento.Monto,
		MetodoPago:    movimiento.MetodoPago,
		Observaciones: movimiento.Observaciones,
	}

	if movimiento.EsIngreso {
		resultado.Tipo = "ingreso"
	}

	if movimiento.IdCategoria != nil {
		categoria, existe := categorias[movimiento.IdCategoria.Id]

		if existe {
			resultado.Categoria = categoria.Nombre

			if categoria.Icono != "" {
				resultado.Icono = categoria.Icono
			}
		}
	}

	return resultado
}

// ArmarMovimientos pasa todos los movimientos al formato del frontend,
// del más reciente al más antiguo.
func ArmarMovimientos(movimientos []models.MovimientoIngresoEgreso, categorias []models.Categoria) []models.ResultadoMovimiento {
	mapa := mapaCategorias(categorias)

	lista := []models.ResultadoMovimiento{}

	for _, movimiento := range movimientos {
		lista = append(lista, armarMovimiento(movimiento, mapa))
	}

	sort.SliceStable(lista, func(i, j int) bool {
		if lista[i].Fecha != lista[j].Fecha {
			return lista[i].Fecha > lista[j].Fecha
		}

		return lista[i].Id > lista[j].Id
	})

	return lista
}

// FiltrarMovimientos deja solo los movimientos que cumplen todos los filtros.
func FiltrarMovimientos(lista []models.ResultadoMovimiento, filtros models.FiltrosMovimientos) []models.ResultadoMovimiento {
	texto := Normalizar(filtros.Texto)
	categoria := Normalizar(filtros.Categoria)

	resultado := []models.ResultadoMovimiento{}

	for _, movimiento := range lista {
		if filtros.Tipo != "" && filtros.Tipo != "todos" && movimiento.Tipo != filtros.Tipo {
			continue
		}

		if categoria != "" && categoria != "todas" && Normalizar(movimiento.Categoria) != categoria {
			continue
		}

		// Las fechas "AAAA-MM-DD" se pueden comparar como texto
		if filtros.Desde != "" && movimiento.Fecha < filtros.Desde {
			continue
		}

		if filtros.Hasta != "" && movimiento.Fecha > filtros.Hasta {
			continue
		}

		if texto != "" && !strings.Contains(Normalizar(movimiento.Concepto), texto) {
			continue
		}

		resultado = append(resultado, movimiento)
	}

	return resultado
}

// CalcularTotales suma los movimientos (tarjetas del Libro mayor).
func CalcularTotales(lista []models.ResultadoMovimiento) models.ResultadoTotales {
	totales := models.ResultadoTotales{}

	for _, movimiento := range lista {
		if movimiento.Tipo == "ingreso" {
			totales.TotalIngresos += movimiento.Valor
		} else {
			totales.TotalGastos += movimiento.Valor

			if Normalizar(movimiento.Categoria) == "gastos hormiga" {
				totales.GastosHormiga += movimiento.Valor
			}
		}

		if Normalizar(movimiento.Categoria) == "ahorro" {
			totales.TotalAhorros += movimiento.Valor
		}
	}

	totales.Ahorro = totales.TotalIngresos - totales.TotalGastos
	totales.PorcentajeAhorro = porcentaje(totales.Ahorro, totales.TotalIngresos)
	totales.PorcentajeGastado = 100 - totales.PorcentajeAhorro
	// Lo que quedaría si se controlan los gastos hormiga
	totales.CapacidadAhorro = totales.TotalIngresos - (totales.TotalGastos - totales.GastosHormiga)
	// Si el ahorro de este periodo se repitiera 12 veces
	totales.ProyeccionAhorroAnual = totales.Ahorro * 12

	return totales
}

// CalcularSalud: "Alto riesgo" si gasta más de lo que gana,
// "Atención" si ahorra menos del 10 %, si no "En orden".
func CalcularSalud(totales models.ResultadoTotales) models.ResultadoSalud {
	if totales.Ahorro < 0 {
		return models.ResultadoSalud{Estado: "Alto riesgo", Mensaje: "Estás gastando más de lo que ganas."}
	}

	ahorro := 0.0

	if totales.TotalIngresos > 0 {
		ahorro = totales.Ahorro / totales.TotalIngresos * 100
	}

	if ahorro < 10 {
		return models.ResultadoSalud{Estado: "Atención", Mensaje: "Tu margen de ahorro es bajo, revisa tus gastos."}
	}

	return models.ResultadoSalud{Estado: "En orden", Mensaje: "Tus finanzas van bien, sigue así."}
}

// CalcularDistribucion: gastos por categoría (dona), de mayor a menor.
func CalcularDistribucion(lista []models.ResultadoMovimiento) []models.ResultadoParte {
	montos := map[string]float64{}
	total := 0.0

	for _, movimiento := range lista {
		if movimiento.Tipo == "gasto" {
			montos[movimiento.Categoria] += movimiento.Valor
			total += movimiento.Valor
		}
	}

	partes := []models.ResultadoParte{}

	for etiqueta, monto := range montos {
		partes = append(partes, models.ResultadoParte{
			Etiqueta:   etiqueta,
			Monto:      monto,
			Porcentaje: porcentaje(monto, total),
		})
	}

	sort.Slice(partes, func(i, j int) bool {
		if partes[i].Monto != partes[j].Monto {
			return partes[i].Monto > partes[j].Monto
		}

		return partes[i].Etiqueta < partes[j].Etiqueta
	})

	return partes
}

func PrimerosMovimientos(lista []models.ResultadoMovimiento, cantidad int) []models.ResultadoMovimiento {
	if len(lista) > cantidad {
		return lista[:cantidad]
	}

	return lista
}

// CalcularLibrosAdmin agrupa los movimientos por usuario
// (filas de la más antigua a la más reciente).
func CalcularLibrosAdmin(lista []models.ResultadoMovimiento, nombres map[int]string) models.ResultadoLibrosAdmin {
	libros := map[int]*models.ResultadoLibroUsuario{}
	resultado := models.ResultadoLibrosAdmin{Libros: []models.ResultadoLibroUsuario{}}

	for i := len(lista) - 1; i >= 0; i-- {
		movimiento := lista[i]

		libro, existe := libros[movimiento.IdUsuario]

		if !existe {
			libro = &models.ResultadoLibroUsuario{
				IdUsuario:     movimiento.IdUsuario,
				NombreUsuario: nombres[movimiento.IdUsuario],
				Filas:         []models.ResultadoFilaLibro{},
			}
			libros[movimiento.IdUsuario] = libro
		}

		fila := models.ResultadoFilaLibro{
			Id:        movimiento.Id,
			Fecha:     movimiento.Fecha,
			Concepto:  movimiento.Concepto,
			Categoria: movimiento.Categoria,
		}

		if movimiento.Tipo == "ingreso" {
			fila.Ingreso = movimiento.Valor
			libro.TotalIngresos += movimiento.Valor
		} else {
			fila.Egreso = movimiento.Valor
			libro.TotalEgresos += movimiento.Valor
		}

		libro.Filas = append(libro.Filas, fila)
	}

	for _, libro := range libros {
		libro.Resultado = libro.TotalIngresos - libro.TotalEgresos
		resultado.TotalIngresosPlataforma += libro.TotalIngresos
		resultado.TotalEgresosPlataforma += libro.TotalEgresos
		resultado.Libros = append(resultado.Libros, *libro)
	}

	sort.Slice(resultado.Libros, func(i, j int) bool {
		return resultado.Libros[i].IdUsuario < resultado.Libros[j].IdUsuario
	})

	resultado.UsuariosActivos = len(resultado.Libros)

	return resultado
}

// ============================================================
// METAS
// ============================================================

// Las metas viejas guardan el ícono como palabra; el frontend muestra emojis.
var iconosMeta = map[string]string{"shield": "🛟", "plane": "✈️", "home": "🏠", "car": "🚙"}

func iconoMeta(icono string) string {
	emoji, existe := iconosMeta[icono]

	if existe {
		return emoji
	}

	if icono == "" {
		return "🎯"
	}

	return icono
}

// ArmarMeta calcula el porcentaje (máximo 100), lo que falta y si está cumplida.
func ArmarMeta(meta models.Meta) models.ResultadoMeta {
	avance := 0

	if meta.MontoObjetivo > 0 {
		avance = int(math.Min(100, math.Round(meta.MontoActual/meta.MontoObjetivo*100)))
	}

	return models.ResultadoMeta{
		Id:         meta.Id,
		IdUsuario:  meta.IdUsuario,
		Nombre:     meta.Nombre,
		Icono:      iconoMeta(meta.Icono),
		Actual:     meta.MontoActual,
		Objetivo:   meta.MontoObjetivo,
		Porcentaje: avance,
		Faltante:   math.Max(0, meta.MontoObjetivo-meta.MontoActual),
		Cumplida:   meta.MontoActual >= meta.MontoObjetivo,
	}
}

// CalcularMetas arma todas las metas y los totales de la pantalla de Metas.
func CalcularMetas(metas []models.Meta) models.ResultadoMetas {
	resultado := models.ResultadoMetas{Metas: []models.ResultadoMeta{}}

	for _, meta := range metas {
		item := ArmarMeta(meta)

		resultado.Metas = append(resultado.Metas, item)
		resultado.TotalAhorrado += item.Actual
		resultado.TotalObjetivo += item.Objetivo

		if item.Cumplida {
			resultado.MetasAlcanzadas++
		}
	}

	sort.Slice(resultado.Metas, func(i, j int) bool {
		return resultado.Metas[i].Id < resultado.Metas[j].Id
	})

	resultado.TotalMetas = len(resultado.Metas)
	resultado.ProgresoGeneral = porcentaje(resultado.TotalAhorrado, resultado.TotalObjetivo)

	return resultado
}

// ============================================================
// INVERSIONES
// ============================================================

// BuscarNivelRiesgo busca un nivel de riesgo por su nombre ("bajo" encuentra "Bajo").
func BuscarNivelRiesgo(niveles []models.NivelRiesgo, nombre string) (models.NivelRiesgo, error) {
	for _, nivel := range niveles {
		if Normalizar(nivel.Nombre) == Normalizar(nombre) {
			return nivel, nil
		}
	}

	return models.NivelRiesgo{}, fmt.Errorf("El nivel de riesgo \"%s\" no existe.", nombre)
}

func mapaNiveles(niveles []models.NivelRiesgo) map[int]string {
	mapa := map[int]string{}

	for _, nivel := range niveles {
		mapa[nivel.Id] = nivel.Nombre
	}

	return mapa
}

// CalcularRentabilidad pasa el rendimiento en pesos a % del monto (dos decimales).
func CalcularRentabilidad(rendimiento, monto float64) float64 {
	if monto == 0 {
		return 0
	}

	return math.Round(rendimiento/monto*10000) / 100
}

// rendimientoDe: si la inversión no tiene rendimiento en pesos, se calcula con su rentabilidad.
func rendimientoDe(inversion models.Inversion) float64 {
	if inversion.Rendimiento == 0 && inversion.Rentabilidad != 0 {
		return math.Round(inversion.Monto * inversion.Rentabilidad / 100)
	}

	return inversion.Rendimiento
}

// ArmarInversion pasa una inversión al formato del frontend.
func ArmarInversion(inversion models.Inversion, niveles []models.NivelRiesgo) models.ResultadoInversion {
	return armarInversion(inversion, mapaNiveles(niveles))
}

func armarInversion(inversion models.Inversion, niveles map[int]string) models.ResultadoInversion {
	rendimiento := rendimientoDe(inversion)

	riesgo := ""

	if inversion.IdNivelRiesgo != nil {
		riesgo = niveles[inversion.IdNivelRiesgo.Id]
	}

	duracion := inversion.Duracion

	if duracion == "" {
		duracion = "Sin definir"
	}

	return models.ResultadoInversion{
		Id:                    inversion.Id,
		IdUsuario:             inversion.IdUsuario,
		Nombre:                inversion.Nombre,
		Monto:                 inversion.Monto,
		Rentabilidad:          inversion.Rentabilidad,
		Rendimiento:           rendimiento,
		RendimientoPorcentaje: porcentajeDecimal(rendimiento, inversion.Monto),
		Riesgo:                riesgo,
		Duracion:              duracion,
	}
}

// CalcularInversiones calcula todo lo de la pantalla de Inversiones.
func CalcularInversiones(inversiones []models.Inversion, niveles []models.NivelRiesgo, historial []models.HistorialInversion) models.ResultadoInversiones {
	mapa := mapaNiveles(niveles)

	resultado := models.ResultadoInversiones{
		Inversiones:    []models.ResultadoInversion{},
		TotalPorRiesgo: map[string]float64{"Bajo": 0, "Medio": 0, "Alto": 0},
		Carteras:       []models.ResultadoParte{},
	}

	ids := map[int]bool{}

	for _, inversion := range inversiones {
		item := armarInversion(inversion, mapa)

		resultado.Inversiones = append(resultado.Inversiones, item)
		resultado.BalanceInversiones += item.Monto
		resultado.RendimientoTotal += item.Rendimiento

		if item.Riesgo != "" {
			resultado.TotalPorRiesgo[item.Riesgo] += item.Monto
		}

		ids[inversion.Id] = true
	}

	sort.Slice(resultado.Inversiones, func(i, j int) bool {
		return resultado.Inversiones[i].Id < resultado.Inversiones[j].Id
	})

	resultado.RendimientoPorcentaje = porcentajeDecimal(resultado.RendimientoTotal, resultado.BalanceInversiones)

	// Cartera: qué parte del total es cada inversión
	for _, item := range resultado.Inversiones {
		resultado.Carteras = append(resultado.Carteras, models.ResultadoParte{
			Etiqueta:   item.Nombre,
			Monto:      item.Monto,
			Porcentaje: porcentaje(item.Monto, resultado.BalanceInversiones),
		})
	}

	// Alerta si la mitad o más del dinero está en riesgo alto
	enRiesgoAlto := porcentaje(resultado.TotalPorRiesgo["Alto"], resultado.BalanceInversiones)

	if resultado.BalanceInversiones > 0 && enRiesgoAlto >= 50 {
		resultado.AlertaConcentracion = models.ResultadoAlerta{
			Mostrar: true,
			Mensaje: fmt.Sprintf("El %d%% de tu dinero invertido está en riesgo alto. Considera diversificar en opciones de menor riesgo.", enRiesgoAlto),
		}
	}

	resultado.Crecimiento = calcularCrecimiento(historial, ids)

	return resultado
}

// calcularCrecimiento suma, mes a mes, el valor de las inversiones indicadas.
func calcularCrecimiento(historial []models.HistorialInversion, ids map[int]bool) []models.ResultadoPunto {
	porPeriodo := map[string]float64{}

	for _, punto := range historial {
		if punto.IdInversion != nil && ids[punto.IdInversion.Id] {
			porPeriodo[punto.Fecha.Format("2006-01")] += punto.Valor
		}
	}

	puntos := []models.ResultadoPunto{}

	for periodo, valor := range porPeriodo {
		puntos = append(puntos, models.ResultadoPunto{Periodo: periodo, Valor: valor})
	}

	sort.Slice(puntos, func(i, j int) bool {
		return puntos[i].Periodo < puntos[j].Periodo
	})

	return puntos
}
