// Package finanzas tiene toda la logica del modulo de finanzas del MID:
// resumen, libro mayor, Excel, metas, inversiones y la vista del admin.
//
// Aqui NO se crea, edita ni borra un registro suelto: eso la pagina lo
// hace directo en crud_finanzas (puerto 8082), como dice la guia.
// El MID solo hace lo que necesita logica: juntar tablas, validar o calcular.
package finanzas

import (
	"fmt"
	"strings"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// FUNCIONES DE APOYO
// Funciones pequenas que usan los demas archivos de esta carpeta.
// Empiezan con minuscula, por eso solo se pueden usar aqui adentro
// (los controladores no las ven).
// ============================================================

// categoriasPorId trae TODAS las categorias en una sola llamada al CRUD
// y las guarda en un mapa por su Id. Asi, para poner el nombre de la
// categoria a cada movimiento, no hay que llamar al CRUD una y otra vez.
func categoriasPorId() (map[int]models.Categoria, error) {

	var categorias []models.Categoria

	if err := helpers.ConsultarLista(helpers.UrlCrudFinanzas+"/categoria?limit=1000", &categorias); err != nil {
		return nil, err
	}

	mapa := map[int]models.Categoria{}
	for _, categoria := range categorias {
		mapa[categoria.Id] = categoria
	}

	return mapa, nil
}

// categoriasPorNombre hace lo contrario: dado el nombre de una categoria
// (en minuscula, ej. "vivienda") devuelve su Id. Lo usa el Excel, porque
// la plantilla trae el nombre y el CRUD necesita el Id.
func categoriasPorNombre(categorias map[int]models.Categoria) map[string]int {

	mapa := map[string]int{}
	for _, categoria := range categorias {
		mapa[strings.ToLower(strings.TrimSpace(categoria.Nombre))] = categoria.Id
	}

	return mapa
}

// nombresUsuarios pide a crud_auth todos los usuarios y devuelve el
// nombre completo de cada uno por su Id ("Sharith R"...).
// Es privada para no chocar con NombresUsuarios del service de auth.
func nombresUsuarios() (map[int]string, error) {

	// Solo nos interesan 3 campos del usuario
	var usuarios []struct {
		Id       int    `json:"Id"`
		Nombre   string `json:"Nombre"`
		Apellido string `json:"Apellido"`
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?limit=10000", &usuarios); err != nil {
		return nil, err
	}

	nombres := map[int]string{}
	for _, usuario := range usuarios {
		nombres[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	return nombres, nil
}

// usuarioExiste pregunta a crud_auth si el usuario existe.
// Es lo primero que hacen las funciones que reciben un id de usuario:
//   - id en 0 o negativo  -> error 400
//   - el usuario no existe -> error 404
//   - crud_auth apagado    -> error 502
func usuarioExiste(idUsuario int) error {

	if idUsuario <= 0 {
		return errIdUsuario
	}

	var usuario struct {
		Id int `json:"Id"`
	}

	return helpers.ConsultarUno(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, idUsuario), &usuario, mensajeUsuarioNoExiste)
}

// tipoMovimiento convierte el campo EsIngreso (true/false) en la palabra
// que usa la pagina: "ingreso" o "gasto".
func tipoMovimiento(movimiento models.Movimiento) string {
	if movimiento.EsIngreso {
		return "ingreso"
	}
	return "gasto"
}

// armarMovimiento pasa un movimiento del formato del CRUD (Nombre, Monto,
// EsIngreso...) al formato que usa la pagina (concepto, valor, tipo...),
// y le agrega el nombre y el icono de su categoria.
func armarMovimiento(movimiento models.Movimiento, categorias map[int]models.Categoria) map[string]interface{} {

	// Si el movimiento no tiene categoria, queda con nombre e icono vacios
	categoria := models.Categoria{}
	if movimiento.IdCategoria != nil {
		categoria = categorias[movimiento.IdCategoria.Id]
	}

	return map[string]interface{}{
		"id":            movimiento.Id,
		"fecha":         movimiento.Fecha.Format("2006-01-02"),
		"concepto":      movimiento.Nombre,
		"id_categoria":  categoria.Id,
		"categoria":     categoria.Nombre,
		"icono":         categoria.Icono,
		"tipo":          tipoMovimiento(movimiento),
		"valor":         movimiento.Monto,
		"metodoPago":    movimiento.MetodoPago, // mismo nombre que usa el cliente
		"observaciones": movimiento.Observaciones,
	}
}

// saludFinanciera dice como van las finanzas del usuario:
//   - "Alto riesgo" si gasta mas de lo que gana (balance negativo).
//   - "Atencion" si ahorra menos del 10 % de lo que gana.
//   - "En orden" en cualquier otro caso.
func saludFinanciera(balance float64, porcentajeAhorro float64) map[string]string {

	if balance < 0 {
		return map[string]string{"estado": "Alto riesgo", "mensaje": "Estas gastando mas de lo que ganas."}
	}

	if porcentajeAhorro < 10 {
		return map[string]string{"estado": "Atencion", "mensaje": "Tu margen de ahorro es bajo, revisa tus gastos."}
	}

	return map[string]string{"estado": "En orden", "mensaje": "Tus finanzas van bien, sigue asi."}
}
