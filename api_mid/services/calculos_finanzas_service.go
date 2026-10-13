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

