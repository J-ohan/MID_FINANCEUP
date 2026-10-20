package finanzas

import (
	"fmt"
	"math"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// METAS DE AHORRO (tabla meta)
// GET /v1/metas/usuario/:id
// PUT /v1/metas/:id/abono
//
// Crear, editar o borrar una meta no esta aqui: la pagina lo hace
// directo en crud_finanzas (/v1/meta), como dice la guia.
// ============================================================

// armarMeta pasa una meta al formato de la pagina y calcula:
//   - porcentaje: cuanto lleva de su objetivo (maximo 100).
//     Ejemplo: lleva 250 de 1000 -> 25 %.
//   - cumplida: si ya ahorro lo que queria.
func armarMeta(meta models.Meta) map[string]interface{} {

	porcentaje := 0.0
	if meta.MontoObjetivo > 0 {
		porcentaje = math.Min(100, math.Round(meta.MontoActual/meta.MontoObjetivo*100))
	}

	return map[string]interface{}{
		"id":         meta.Id,
		"nombre":     meta.Nombre,
		"icono":      meta.Icono,
		"actual":     meta.MontoActual,
		"objetivo":   meta.MontoObjetivo,
		"porcentaje": porcentaje,
		"cumplida":   meta.MontoObjetivo > 0 && meta.MontoActual >= meta.MontoObjetivo,
	}
}

// MetasUsuario devuelve las metas del usuario con su porcentaje, cuantas
// estan cumplidas y los totales (todo lo ahorrado y todo lo que quiere ahorrar).
func MetasUsuario(idUsuario int) (map[string]interface{}, error) {

	// 0. El usuario debe existir
	if err := usuarioExiste(idUsuario); err != nil {
		return nil, err
	}

	// 1. Pedir al CRUD las metas activas del usuario
	var metas []models.Meta

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/meta?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudFinanzas, idUsuario), &metas); err != nil {
		return nil, err
	}

	// 2. Armar cada meta y sumar los totales
	lista := []map[string]interface{}{}
	totalAhorrado := 0.0
	totalObjetivo := 0.0
	cumplidas := 0

	for _, meta := range metas {

		item := armarMeta(meta)

		if item["cumplida"].(bool) {
			cumplidas++
		}

		totalAhorrado += meta.MontoActual
		totalObjetivo += meta.MontoObjetivo

		lista = append(lista, item)
	}

	return map[string]interface{}{
		"metas":          lista,
		"total_ahorrado": totalAhorrado,
		"total_objetivo": totalObjetivo,
		"cumplidas":      cumplidas,
	}, nil
}

// AbonarMeta suma dinero a lo que lleva ahorrado una meta.
//
// Ojo: el PUT del CRUD reemplaza el registro COMPLETO. Por eso primero
// se trae la meta entera (en un map, para no perder ningun campo), se
// cambia solo MontoActual y se manda todo de vuelta.
func AbonarMeta(idMeta int, valor float64) (map[string]interface{}, error) {

	// 1. El id de la meta y el abono deben ser mayores a 0
	if idMeta <= 0 {
		return nil, errIdMeta
	}

	if valor <= 0 {
		return nil, errValorAbono
	}

	// 2. Traer la meta completa (404 si no existe)
	var meta map[string]interface{}
	direccion := fmt.Sprintf("%s/meta/%d", helpers.UrlCrudFinanzas, idMeta)

	if err := helpers.ConsultarUno(direccion, &meta, mensajeMetaNoExiste); err != nil {
		return nil, err
	}

	// Una meta borrada (Activo = false) no recibe abonos
	if activa, _ := meta["Activo"].(bool); !activa {
		return nil, errMetaBorrada
	}

	// 3. Sumar el abono a lo que ya llevaba
	actual, _ := meta["MontoActual"].(float64)
	meta["MontoActual"] = actual + valor

	// 4. Guardar la meta completa con el nuevo valor
	if err := helpers.Actualizar(direccion, meta); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       idMeta,
		"actual":   meta["MontoActual"],
		"objetivo": meta["MontoObjetivo"],
	}, nil
}
