package finanzas

import (
	"net/http"

	"api_mid_financeup/helpers"
)

// ============================================================
// ERRORES DE FINANZAS
// Todos los mensajes de error del modulo estan aqui, en un solo
// lugar. Si un mensaje cambia, se cambia solo en este archivo.
//
// Cada error trae el codigo HTTP que recibe la pagina:
//   400 = la pagina mando algo mal (falta un dato o esta mal escrito)
//   404 = lo que se pide no existe (o fue borrado)
//   502 = un CRUD esta apagado o no respondio
//         (ese lo arma helpers/crud.go, no esta en esta lista)
// ============================================================

// LimiteFilasExcel es el maximo de filas que se aceptan en una importacion.
const LimiteFilasExcel = 1000

// Usuario
var (
	errIdUsuario = helpers.NuevoError(http.StatusBadRequest, "El id del usuario es obligatorio y debe ser mayor a 0")
)

// mensajeUsuarioNoExiste se usa con helpers.ConsultarUno, que arma el 404.
const mensajeUsuarioNoExiste = "El usuario no existe"

// Filtros del libro mayor y de la exportacion
var (
	errTipoFiltro  = helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser todos, ingreso o gasto")
	errFechaFiltro = helpers.NuevoError(http.StatusBadRequest, "Las fechas deben tener el formato AAAA-MM-DD")
	errRangoFechas = helpers.NuevoError(http.StatusBadRequest, "La fecha 'desde' no puede ser mayor que la fecha 'hasta'")
)

// errCategoriaNoExiste: el filtro pide una categoria que no esta en la base.
func errCategoriaNoExiste(nombre string) error {
	return helpers.NuevoError(http.StatusBadRequest, "La categoria \""+nombre+"\" no existe")
}

// Importar Excel (cada fila mala se reporta con uno de estos motivos)
var (
	errModoImportar    = helpers.NuevoError(http.StatusBadRequest, "El modo debe ser validar o confirmar")
	errSinMovimientos  = helpers.NuevoError(http.StatusBadRequest, "No llego ningun movimiento")
	errDemasiadasFilas = helpers.NuevoError(http.StatusBadRequest, "El archivo tiene demasiadas filas (maximo 1000)")
	errFechaFila       = helpers.NuevoError(http.StatusBadRequest, "La fecha debe ser AAAA-MM-DD")
	errConceptoOValor  = helpers.NuevoError(http.StatusBadRequest, "Falta el concepto o el valor")
	errConceptoLargo   = helpers.NuevoError(http.StatusBadRequest, "El concepto no puede tener mas de 120 caracteres")
	errMetodoPagoLargo = helpers.NuevoError(http.StatusBadRequest, "El metodo de pago no puede tener mas de 50 caracteres")
	errTipoMovimiento  = helpers.NuevoError(http.StatusBadRequest, "El tipo debe ser ingreso o gasto")
)

// errNoSeImporto: el CRUD fallo a la mitad y se deshizo la importacion.
func errNoSeImporto(concepto string) error {
	return helpers.NuevoError(http.StatusBadGateway, "No se pudo guardar \""+concepto+"\", no se importo ninguna fila")
}

// Metas
var (
	errIdMeta      = helpers.NuevoError(http.StatusBadRequest, "El id de la meta debe ser mayor a 0")
	errValorAbono  = helpers.NuevoError(http.StatusBadRequest, "El valor del abono debe ser mayor a 0")
	errMetaBorrada = helpers.NuevoError(http.StatusNotFound, mensajeMetaNoExiste)
)

// mensajeMetaNoExiste se usa con helpers.ConsultarUno, que arma el 404.
const mensajeMetaNoExiste = "La meta no existe"

// Avisos: no detienen la respuesta, solo le cuentan a la pagina que
// una parte no se pudo cargar (por ejemplo, la grafica de inversiones).
const avisoGraficaInversiones = "No se pudo cargar el historial de algunas inversiones; la grafica puede estar incompleta"
