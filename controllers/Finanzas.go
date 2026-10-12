// ============================================================================
// FINANZAS.GO — Controlador del módulo de Finanzas en el MID
// ============================================================================
//
// ¿QUÉ ES EL MID?
//   El MID es el "intermediario" entre el frontend (Angular) y los CRUD.
//   El frontend nunca habla directo con la base de datos: le pide todo al MID.
//   El MID le pide los datos a los CRUD, los revisa, los une, hace los
//   cálculos y le devuelve al frontend la información lista para mostrar.
//
//     Frontend (Angular)  →  MID (este archivo, puerto 8080)  →  CRUD  →  Base de datos
//
// ¿CON QUÉ CRUD HABLA ESTE ARCHIVO?
//   - crud_finanzas (puerto 8082): categorías, movimientos, metas, inversiones...
//   - crud_auth     (puerto 8081): para revisar que el usuario exista y saber su nombre.
//
// ¿CÓMO ESTÁ ORGANIZADO?
//   1. Tablas:       cómo llegan los datos desde los CRUD (una plantilla por tabla).
//   2. Datos:        cómo llegan los datos que envía el frontend (los formularios).
//   3. Respuestas:   cómo salen los datos que el MID le devuelve al frontend.
//   4. Métodos:      una función por cada ruta (URL) del módulo, agrupadas en
//                    Finanzas, Libro mayor, Metas, Inversiones y Admin.
//   5. Validaciones: reglas que deben cumplir los datos antes de guardarlos.
//   6. Consultas:    funciones que hablan con los CRUD (leer, crear, actualizar, borrar).
//   7. Cálculos:     sumas, porcentajes y transformaciones de los datos.
//   8. Ayudas HTTP:  funciones pequeñas para leer la petición y responder.
//
// PEQUEÑO GLOSARIO DE GO (para leer el código más fácil)
//   struct          Una "plantilla" con varios campos, como una ficha o un formulario.
//   `json:"x"`      Dice cómo se llama ese campo cuando se convierte a JSON.
//   err y nil       Casi todas las funciones devuelven un error llamado "err".
//                   Si err es nil (que significa "nada"), todo salió bien.
//                   Si err NO es nil, algo falló.
//   *Tipo           Un puntero: una referencia a un dato. Puede valer nil (vacío).
//   &variable       Le pasa a otra función la variable "original" para que la pueda llenar o cambiar.
//   []Tipo          Una lista de elementos de ese tipo.
//   map[A]B         Un diccionario: para cada clave A guarda un valor B.
//   for _, x := range lista   Recorre la lista, un elemento a la vez.
//   :=              Crea una variable nueva y le pone un valor.
//   defer           Deja algo pendiente para hacerlo justo antes de salir de la función.
//   (c *FinanzasController)   Indica que esa función pertenece al controlador de finanzas.
//                   Dentro de ella, "c" es el controlador: con él se lee la petición
//                   y se envía la respuesta.
// ============================================================================

package controllers

import (
	"bytes"         // Para armar el cuerpo (body) de las peticiones que se envían a los CRUD
	"encoding/json" // Para convertir JSON a structs de Go y al revés
	"fmt"           // Para armar textos con valores, por ejemplo URLs con un id
	"io"            // Para leer el cuerpo de las respuestas de los CRUD
	"math"          // Funciones matemáticas: redondear, máximo, mínimo, potencias
	"net/http"      // Para hacer peticiones HTTP y usar los códigos de respuesta (200, 404...)
	"sort"          // Para ordenar listas
	"strconv"       // Para convertir texto a número y número a texto
	"strings"       // Funciones para trabajar con textos (quitar espacios, minúsculas...)
	"time"          // Para manejar fechas
	"unicode"       // Para pasar una letra a mayúscula
	"unicode/utf8"  // Para contar letras (no bytes) en textos con tildes o emojis

	beego "github.com/beego/beego/v2/server/web" // El framework web que usa todo el MID
)

// Valores fijos que se usan en todo el archivo.
const (
	// Dirección del CRUD de finanzas. Todas las tablas de finanzas se piden aquí.
	urlCrudFinanzas = "http://localhost:8082/v1"

	// Dirección del CRUD de usuarios (auth). Se usa para revisar que el usuario exista.
	urlCrudUsuarios = "http://localhost:8081/v1"

	// Máximo de registros que se le piden al CRUD en cada consulta.
	// Si no se envía "limit", los CRUD hechos con bee devuelven solo 10 registros.
	limiteCrud = 10000
)

// ============================================================================
// 1. TABLAS: JSON que el MID recibe de los CRUD
// ----------------------------------------------------------------------------
// Cada struct es una "plantilla" de una tabla de la base de datos.
// Los nombres de los campos son los mismos que usan los modelos de crud_finanzas,
// así Go puede llenar la plantilla automáticamente con el JSON que llega.
//
// Se incluyen TODAS las columnas de cada tabla porque, cuando se actualiza un
// registro (PUT), el CRUD reemplaza el registro completo: si faltara una
// columna en la plantilla, su valor se borraría en la base de datos.
//
// Nota: "Ref" y "Usuario" no están en este archivo porque ya existen en auth.go
// (que es del mismo paquete). Ref representa una llave foránea tal como la
// manda el CRUD, por ejemplo: "IdCategoria": {"Id": 3}
// ============================================================================

// Tabla finanzas.categoria — las categorías de los movimientos (Arriendo, Salud, Salario...)
type Categoria struct {
	Id          int    `json:"Id"`          // Número único de la categoría
	Nombre      string `json:"Nombre"`      // Nombre que se muestra, ej. "Gastos hormiga"
	Descripcion string `json:"Descripcion"` // Explicación corta de la categoría
	Icono       string `json:"Icono"`       // Emoji que se muestra en pantalla, ej. "☕"
	Activo      bool   `json:"Activo"`      // true = se usa; false = está "borrada"
}

// Tabla finanzas.movimiento_ingreso_egreso — cada ingreso o gasto del Libro mayor
type MovimientoIngresoEgreso struct {
	Id            int       `json:"Id"`            // Número único del movimiento
	IdUsuario     int       `json:"IdUsuario"`     // Dueño del movimiento
	IdCategoria   *Ref      `json:"IdCategoria"`   // Categoría (puede venir vacía = nil)
	Nombre        string    `json:"Nombre"`        // Concepto, ej. "Mercado"
	Monto         float64   `json:"Monto"`         // Valor en pesos
	EsIngreso     bool      `json:"EsIngreso"`     // true = ingreso; false = gasto
	Fecha         time.Time `json:"Fecha"`         // Día del movimiento
	MetodoPago    string    `json:"MetodoPago"`    // Efectivo, Tarjeta, Transferencia...
	Observaciones string    `json:"Observaciones"` // Notas opcionales
	Activo        bool      `json:"Activo"`        // false = eliminado
}

// Tabla finanzas.meta — las metas de ahorro del usuario
type Meta struct {
	Id                 int       `json:"Id"`                 // Número único de la meta
	IdUsuario          int       `json:"IdUsuario"`          // Dueño de la meta
	IdEditarMeta       *Ref      `json:"IdEditarMeta"`       // Relación con editar_meta (puede ser nil)
	IdMovimientoDinero *Ref      `json:"IdMovimientoDinero"` // Relación con movimiento_meta (puede ser nil)
	Nombre             string    `json:"Nombre"`             // Ej. "Vacaciones"
	Descripcion        string    `json:"Descripcion"`        // Detalle opcional
	MontoObjetivo      float64   `json:"MontoObjetivo"`      // Cuánto se quiere ahorrar
	MontoActual        float64   `json:"MontoActual"`        // Cuánto lleva ahorrado
	FechaLimite        time.Time `json:"FechaLimite"`        // Fecha para cumplirla (opcional)
	Color              string    `json:"Color"`              // Color de la tarjeta (opcional)
	Icono              string    `json:"Icono"`              // Emoji de la meta
	Activo             bool      `json:"Activo"`             // false = eliminada
}

// Tabla finanzas.nivel_riesgo — los niveles de riesgo: Bajo, Medio y Alto
type NivelRiesgo struct {
	Id     int    `json:"Id"`     // Número único del nivel
	Nombre string `json:"Nombre"` // "Bajo", "Medio" o "Alto"
	Activo bool   `json:"Activo"` // false = no se usa
}

// Tabla finanzas.inversion — las inversiones del usuario
type Inversion struct {
	Id                 int       `json:"Id"`                 // Número único de la inversión
	IdUsuario          int       `json:"IdUsuario"`          // Dueño de la inversión
	IdTipoInversion    *Ref      `json:"IdTipoInversion"`    // Tipo (Acciones, Bonos...). Puede ser nil
	IdNivelRiesgo      *Ref      `json:"IdNivelRiesgo"`      // Nivel de riesgo (obligatorio)
	IdMovimientoDinero *Ref      `json:"IdMovimientoDinero"` // Relación con movimiento_inversion (puede ser nil)
	Nombre             string    `json:"Nombre"`             // Ej. "CDT a un año"
	Monto              float64   `json:"Monto"`              // Dinero invertido
	Rentabilidad       float64   `json:"Rentabilidad"`       // Ganancia en porcentaje, ej. 12.5 (%)
	Rendimiento        float64   `json:"Rendimiento"`        // Ganancia (o pérdida) en pesos
	Duracion           string    `json:"Duracion"`           // Texto, ej. "6 meses", "Flexible"
	FechaInicio        time.Time `json:"FechaInicio"`        // Cuándo empezó
	FechaFin           time.Time `json:"FechaFin"`           // Cuándo termina (opcional)
	Activo             bool      `json:"Activo"`             // false = eliminada
}

// Tabla finanzas.historial_inversion — cuánto valía cada inversión en una fecha.
// Sirve para dibujar la gráfica de crecimiento.
type HistorialInversion struct {
	Id          int       `json:"Id"`          // Número único del registro
	IdInversion *Ref      `json:"IdInversion"` // A qué inversión pertenece
	Fecha       time.Time `json:"Fecha"`       // Día en que se tomó el valor
	Valor       float64   `json:"Valor"`       // Cuánto valía la inversión ese día
	Activo      bool      `json:"Activo"`      // false = no se tiene en cuenta
}

// ============================================================================
// 2. DATOS: JSON que el MID recibe del frontend
// ----------------------------------------------------------------------------
// Son los datos de los formularios. Usan nombres en minúscula con guion bajo
// (id_usuario, metodo_pago), igual que el resto del MID del equipo.
// ============================================================================

// Lo que envía el formulario de "Nuevo movimiento" o "Editar movimiento".
// Rutas: POST /v1/movimientos y PUT /v1/movimientos/:id
// Ejemplo:
//
//	{"id_usuario": 1, "fecha": "2026-03-20", "concepto": "Mercado",
//	 "categoria": "Alimentacion", "tipo": "gasto", "valor": 450000}
type DatosMovimiento struct {
	IdUsuario int     `json:"id_usuario"` // Dueño (solo se usa al crear)
	Fecha     string  `json:"fecha"`      // Fecha en texto: "2026-03-20"
	Concepto  string  `json:"concepto"`   // Qué fue el movimiento
	Categoria string  `json:"categoria"`  // NOMBRE de la categoría: "Salario", "Gastos hormiga"...
	Tipo      string  `json:"tipo"`       // "ingreso" o "gasto"
	Valor     float64 `json:"valor"`      // Valor en pesos (mayor que 0)
	// Estos dos son punteros (*string) para saber si el frontend los envió o no:
	// si no vienen (nil), al editar se deja el valor que ya tenía.
	MetodoPago    *string `json:"metodo_pago"`
	Observaciones *string `json:"observaciones"`
}

// Lo que envía la carga masiva desde la plantilla de Excel.
// Ruta: POST /v1/movimientos/importar
type DatosImportacion struct {
	IdUsuario int               `json:"id_usuario"` // A quién se le cargan los movimientos
	Modo      string            `json:"modo"`       // "validar" (solo revisa) o "confirmar" (guarda)
	Filas     []DatosMovimiento `json:"filas"`      // Cada fila del Excel, con el mismo formato de arriba
}

// Lo que envía el formulario de "Nueva meta" o "Editar meta".
// Rutas: POST /v1/metas y PUT /v1/metas/:id
type DatosMeta struct {
	IdUsuario int     `json:"id_usuario"` // Dueño (solo se usa al crear)
	Nombre    string  `json:"nombre"`     // Nombre de la meta
	Icono     string  `json:"icono"`      // Emoji (si no viene se usa 🎯)
	Actual    float64 `json:"actual"`     // Cuánto lleva ahorrado
	Objetivo  float64 `json:"objetivo"`   // Cuánto quiere ahorrar
}

// Lo que envía el botón "Abonar" de una meta.
// Ruta: PUT /v1/metas/:id/abono
type DatosAbono struct {
	Monto float64 `json:"monto"` // Dinero que se suma a la meta
}

// Lo que envía el formulario de "Nueva inversión" o "Editar inversión".
// Rutas: POST /v1/inversiones y PUT /v1/inversiones/:id
type DatosInversion struct {
	IdUsuario   int     `json:"id_usuario"`  // Dueño (solo se usa al crear)
	Nombre      string  `json:"nombre"`      // Nombre de la inversión
	Monto       float64 `json:"monto"`       // Dinero invertido
	Rendimiento float64 `json:"rendimiento"` // Ganancia o pérdida en pesos (puede ser negativa)
	Riesgo      string  `json:"riesgo"`      // "Bajo", "Medio" o "Alto"
	Duracion    string  `json:"duracion"`    // "6 meses", "1 año", "Flexible"...
}

// Filtros del Libro mayor. No llegan en el cuerpo sino en la URL, por ejemplo:
// /v1/movimientos/usuario/1?tipo=gasto&categoria=Salud&desde=2026-03-01&hasta=2026-03-31&q=mercado
type FiltrosMovimientos struct {
	Tipo      string // "todos", "ingreso" o "gasto"
	Categoria string // "Todas" o el nombre de una categoría
	Desde     string // Fecha inicial "AAAA-MM-DD" (opcional)
	Hasta     string // Fecha final "AAAA-MM-DD" (opcional)
	Texto     string // Palabra a buscar dentro del concepto (opcional)
}

// ============================================================================
// 3. RESPUESTAS: JSON que el MID envía al frontend
// ----------------------------------------------------------------------------
// Es la información ya procesada y lista para mostrar en pantalla.
// ============================================================================

// Una categoría, como la necesita el frontend para los selects y filtros.
type RespuestaCategoria struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
	Icono  string `json:"icono"`
}

// Un movimiento del Libro mayor, ya con el nombre de su categoría y su ícono.
type RespuestaMovimiento struct {
	Id            int     `json:"id"`
	IdUsuario     int     `json:"id_usuario"`
	Fecha         string  `json:"fecha"`     // "2026-03-20"
	Concepto      string  `json:"concepto"`  // Qué fue
	Categoria     string  `json:"categoria"` // Nombre de la categoría ("Sin categoria" si no tiene)
	Icono         string  `json:"icono"`     // Emoji de la categoría
	Tipo          string  `json:"tipo"`      // "ingreso" o "gasto"
	Valor         float64 `json:"valor"`
	MetodoPago    string  `json:"metodo_pago"`
	Observaciones string  `json:"observaciones"`
}

// Las tarjetas de totales del Libro mayor.
// Se calculan solo con los movimientos que pasaron los filtros.
type RespuestaTotales struct {
	TotalIngresos         float64 `json:"total_ingresos"`          // Suma de ingresos
	TotalGastos           float64 `json:"total_gastos"`            // Suma de gastos
	Ahorro                float64 `json:"ahorro"`                  // Ingresos - gastos
	PorcentajeAhorro      int     `json:"porcentaje_ahorro"`       // Qué % de los ingresos quedó
	PorcentajeGastado     int     `json:"porcentaje_gastado"`      // Qué % de los ingresos se gastó
	GastosHormiga         float64 `json:"gastos_hormiga"`          // Gastos de la categoría "Gastos hormiga"
	TotalAhorros          float64 `json:"total_ahorros"`           // Movimientos de la categoría "Ahorro"
	CapacidadAhorro       float64 `json:"capacidad_ahorro"`        // Lo que quedaría sin los gastos hormiga
	ProyeccionAhorroAnual float64 `json:"proyeccion_ahorro_anual"` // El ahorro multiplicado por 12 meses
}

// Respuesta completa del Libro mayor: la lista de movimientos y sus totales.
type RespuestaLibroMayor struct {
	Movimientos []RespuestaMovimiento `json:"movimientos"`
	Totales     RespuestaTotales      `json:"totales"`
}

// Una fila del Excel que tuvo un problema al importar.
type RespuestaErrorFila struct {
	Fila     int    `json:"fila"`     // Número de la fila (empieza en 1)
	Concepto string `json:"concepto"` // Concepto que traía esa fila
	Motivo   string `json:"motivo"`   // Por qué no se puede importar
}

// Resultado de importar el Excel.
type RespuestaImportacion struct {
	Modo      string               `json:"modo"`      // "validar" o "confirmar"
	Validos   int                  `json:"validos"`   // Cuántas filas están bien
	Guardados int                  `json:"guardados"` // Cuántas se guardaron (0 en modo validar)
	Errores   []RespuestaErrorFila `json:"errores"`   // Las filas con problemas
}

// Una meta, ya con su porcentaje calculado.
type RespuestaMeta struct {
	Id         int     `json:"id"`
	IdUsuario  int     `json:"id_usuario"`
	Nombre     string  `json:"nombre"`
	Icono      string  `json:"icono"`
	Actual     float64 `json:"actual"`     // Cuánto lleva
	Objetivo   float64 `json:"objetivo"`   // Cuánto quiere
	Porcentaje int     `json:"porcentaje"` // Avance de 0 a 100
	Faltante   float64 `json:"faltante"`   // Cuánto le falta (nunca negativo)
	Cumplida   bool    `json:"cumplida"`   // true si ya llegó al objetivo
}

// La lista de metas con los totales que muestra la pantalla de Metas.
type RespuestaMetas struct {
	Metas           []RespuestaMeta `json:"metas"`
	TotalMetas      int             `json:"total_metas"`      // Cuántas metas hay
	MetasAlcanzadas int             `json:"metas_alcanzadas"` // Cuántas están cumplidas
	TotalAhorrado   float64         `json:"total_ahorrado"`   // Suma de lo ahorrado en todas
	TotalObjetivo   float64         `json:"total_objetivo"`   // Suma de los objetivos
	ProgresoGeneral int             `json:"progreso_general"` // Avance total en %
}

// Una inversión, ya con su riesgo en texto y su rendimiento calculado.
type RespuestaInversion struct {
	Id                    int     `json:"id"`
	IdUsuario             int     `json:"id_usuario"`
	Nombre                string  `json:"nombre"`
	Monto                 float64 `json:"monto"`
	Rentabilidad          float64 `json:"rentabilidad"` // % guardado en la tabla
	Rendimiento           float64 `json:"rendimiento"`  // Ganancia o pérdida en pesos
	RendimientoPorcentaje float64 `json:"rendimiento_porcentaje"`
	Riesgo                string  `json:"riesgo"`   // "Bajo", "Medio" o "Alto"
	Duracion              string  `json:"duracion"` // "Sin definir" si no tiene
}

// Una parte de un total. Se usa para la dona de gastos y para la cartera de inversiones.
// Ejemplo: {"etiqueta": "Arriendo", "monto": 1500000, "porcentaje": 61}
type RespuestaParte struct {
	Etiqueta   string  `json:"etiqueta"`
	Monto      float64 `json:"monto"`
	Porcentaje int     `json:"porcentaje"`
}

// Un punto de la gráfica de crecimiento. Ejemplo: {"periodo": "2026-05", "valor": 2053000}
type RespuestaPunto struct {
	Periodo string  `json:"periodo"` // Año y mes
	Valor   float64 `json:"valor"`   // Suma del valor de las inversiones ese mes
}

// Un aviso para mostrar (o no) en pantalla.
type RespuestaAlerta struct {
	Mostrar bool   `json:"mostrar"`
	Mensaje string `json:"mensaje"`
}

// Todo lo que muestra la pantalla de Inversiones.
type RespuestaInversiones struct {
	Inversiones           []RespuestaInversion `json:"inversiones"`
	BalanceInversiones    float64              `json:"balance_inversiones"`    // Total invertido
	RendimientoTotal      float64              `json:"rendimiento_total"`      // Ganancia total en pesos
	RendimientoPorcentaje float64              `json:"rendimiento_porcentaje"` // Ganancia total en %
	TotalPorRiesgo        map[string]float64   `json:"total_por_riesgo"`       // Ej. {"Bajo": 500000, "Medio": 0, "Alto": 1000000}
	AlertaConcentracion   RespuestaAlerta      `json:"alerta_concentracion"`   // Aviso si hay mucho en riesgo alto
	Carteras              []RespuestaParte     `json:"carteras"`               // Qué % del total es cada inversión
	Crecimiento           []RespuestaPunto     `json:"crecimiento"`            // Puntos de la gráfica
}

// El estado de "salud financiera" del panel principal.
type RespuestaSalud struct {
	Estado  string `json:"estado"`  // "En orden", "Atención" o "Alto riesgo"
	Mensaje string `json:"mensaje"` // Frase para el usuario
}

// Todo lo que muestra el panel principal de Finanzas.
// Conserva las claves de la primera versión (balance, gastos_por_categoria,
// inversiones) para no romper a quien ya las use.
type RespuestaResumen struct {
	IdUsuario          int                   `json:"id_usuario"`
	TotalIngresos      float64               `json:"total_ingresos"`
	TotalGastos        float64               `json:"total_gastos"`
	Balance            float64               `json:"balance"`
	Disponible         float64               `json:"disponible"` // igual a balance (nombre del frontend)
	PorcentajeAhorro   int                   `json:"porcentaje_ahorro"`
	Salud              RespuestaSalud        `json:"salud"`
	GastosPorCategoria map[string]float64    `json:"gastos_por_categoria"`
	Distribucion       []RespuestaParte      `json:"distribucion"` // lo mismo con %, para la dona
	UltimosMovimientos []RespuestaMovimiento `json:"ultimos_movimientos"`
	Metas              []RespuestaMeta       `json:"metas"`
	TotalInvertido     float64               `json:"total_invertido"`
	GananciaEstimada   float64               `json:"ganancia_estimada"`
	Inversiones        []RespuestaInversion  `json:"inversiones"`
	Crecimiento        []RespuestaPunto      `json:"crecimiento"`
}

// Admin: una fila del libro mayor. El ingreso y el egreso van en columnas separadas
// (si es ingreso, egreso queda en 0, y al revés).
type RespuestaFilaLibro struct {
	Id        int     `json:"id"`
	Fecha     string  `json:"fecha"`
	Concepto  string  `json:"concepto"`
	Categoria string  `json:"categoria"`
	Ingreso   float64 `json:"ingreso"`
	Egreso    float64 `json:"egreso"`
}

// Admin: el libro mayor de UN usuario.
type RespuestaLibroUsuario struct {
	IdUsuario     int                  `json:"id_usuario"`
	NombreUsuario string               `json:"nombre_usuario"` // Ej. "Harold Arciniegas"
	Filas         []RespuestaFilaLibro `json:"filas"`
	TotalIngresos float64              `json:"total_ingresos"`
	TotalEgresos  float64              `json:"total_egresos"`
	Resultado     float64              `json:"resultado"` // Ingresos - egresos
}

// Admin: los libros de todos los usuarios y los totales de toda la plataforma.
type RespuestaLibrosAdmin struct {
	Libros                  []RespuestaLibroUsuario `json:"libros"`
	TotalIngresosPlataforma float64                 `json:"total_ingresos_plataforma"`
	TotalEgresosPlataforma  float64                 `json:"total_egresos_plataforma"`
	UsuariosActivos         int                     `json:"usuarios_activos"` // Usuarios que tienen movimientos
}

// ============================================================================
// 4. MÉTODOS
// ----------------------------------------------------------------------------
// Cada método atiende UNA ruta. Las rutas se conectan con estos métodos en
// routers/router.go, por ejemplo:
//
//	beego.Router("/v1/metas", &controllers.FinanzasController{}, "post:PostMeta")
//
// Casi todos siguen los mismos pasos:
//   1. Leer lo que llegó (el id de la URL o el JSON del formulario).
//   2. Validar los datos.
//   3. Consultar o guardar en los CRUD.
//   4. Calcular y responder.
//
// Patrón que se repite mucho:
//
//	if err != nil {
//		c.responderError(err) // le responde al frontend con el error
//		return                // y termina la función aquí mismo
//	}
// ============================================================================

// FinanzasController es el controlador del módulo. Todos los métodos de abajo
// le pertenecen. "beego.Controller" le da las herramientas para leer la
// petición (c.Ctx.Input) y para responder (c.Data["json"], c.ServeJSON()).
type FinanzasController struct {
	beego.Controller
}

// ----------------------------------------------------------------------------
// FINANZAS (panel principal)
// ----------------------------------------------------------------------------

// GetResumenFinanciero arma todo lo que muestra el panel principal de Finanzas.
//
//	Ruta:     GET /v1/resumen-financiero/:id      (:id = id del usuario)
//	Devuelve: tarjetas (ingresos, gastos, disponible, % de ahorro), salud
//	          financiera, dona de gastos, últimos 5 movimientos, metas,
//	          inversiones y gráfica de crecimiento.
//	Errores:  400 si el id no es un número, 404 si el usuario no existe,
//	          502 si algún CRUD no responde.
func (c *FinanzasController) GetResumenFinanciero() {

	// 1. Leer el id del usuario que viene en la URL
	idUsuario, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Revisar que el usuario exista (le pregunta a crud_auth).
	//    El "_" significa que no necesitamos los datos del usuario, solo saber si existe.
	if _, err := consultarUsuario(idUsuario); err != nil {
		c.responderError(err)
		return
	}

	// 3. Traer de crud_finanzas todo lo que se necesita.
	//    Si cualquier consulta falla, se responde el error y se termina.
	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	movimientos, err := consultarMovimientos(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	metas, err := consultarMetas(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	inversiones, err := consultarInversiones(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	niveles, err := consultarNivelesRiesgo()
	if err != nil {
		c.responderError(err)
		return
	}

	// El historial no detiene la respuesta si falla: la gráfica simplemente sale vacía.
	historial := consultarHistorial()

	// 4. Hacer los cálculos
	lista := armarMovimientos(movimientos, categorias)                      // movimientos con su categoría, del más nuevo al más viejo
	totales := calcularTotales(lista)                                       // ingresos, gastos, ahorro...
	resumenInversiones := armarInversiones(inversiones, niveles, historial) // total invertido, ganancia, gráfica...

	// Gastos por categoría: se calcula como lista (para la dona) y también
	// como diccionario {"Arriendo": 1500000, ...} (formato de la primera versión).
	distribucion := calcularDistribucion(lista)
	gastosPorCategoria := map[string]float64{}
	for _, parte := range distribucion {
		gastosPorCategoria[parte.Etiqueta] = parte.Monto
	}

	// 5. Armar la respuesta y enviarla con código 200 (todo bien)
	c.responder(http.StatusOK, RespuestaResumen{
		IdUsuario:          idUsuario,
		TotalIngresos:      totales.TotalIngresos,
		TotalGastos:        totales.TotalGastos,
		Balance:            totales.Ahorro,
		Disponible:         totales.Ahorro,
		PorcentajeAhorro:   totales.PorcentajeAhorro,
		Salud:              calcularSalud(totales),
		GastosPorCategoria: gastosPorCategoria,
		Distribucion:       distribucion,
		UltimosMovimientos: primerosMovimientos(lista, 5), // solo los 5 más recientes
		Metas:              armarMetas(metas).Metas,
		TotalInvertido:     resumenInversiones.BalanceInversiones,
		GananciaEstimada:   resumenInversiones.RendimientoTotal,
		Inversiones:        resumenInversiones.Inversiones,
		Crecimiento:        resumenInversiones.Crecimiento,
	})
}

// GetCategorias devuelve las categorías activas con su ícono.
// El frontend las usa para llenar los selects y los filtros del Libro mayor.
//
//	Ruta:     GET /v1/categorias
//	Devuelve: [{"id": 1, "nombre": "Arriendo", "icono": "🏠"}, ...]
func (c *FinanzasController) GetCategorias() {

	// 1. Pedir las categorías activas a crud_finanzas
	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Pasar cada categoría al formato del frontend (solo id, nombre e ícono)
	lista := []RespuestaCategoria{}
	for _, categoria := range categorias {
		lista = append(lista, RespuestaCategoria{Id: categoria.Id, Nombre: categoria.Nombre, Icono: categoria.Icono})
	}

	// 3. Responder la lista
	c.responder(http.StatusOK, lista)
}

// ----------------------------------------------------------------------------
// LIBRO MAYOR (tablas movimiento_ingreso_egreso + categoria)
// ----------------------------------------------------------------------------

// GetMovimientosUsuario devuelve los movimientos de un usuario (filtrados)
// y los totales de esos movimientos.
//
//	Ruta:     GET /v1/movimientos/usuario/:id?tipo=&categoria=&desde=&hasta=&q=
//	          Todos los filtros son opcionales.
//	Devuelve: {"movimientos": [...], "totales": {...}}
//	Errores:  400 si el id o algún filtro está mal, 404 si el usuario no existe.
func (c *FinanzasController) GetMovimientosUsuario() {

	// 1. Leer el id del usuario y los filtros que vienen en la URL
	idUsuario, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	filtros, err := c.filtrosDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Revisar que el usuario exista (crud_auth)
	if _, err := consultarUsuario(idUsuario); err != nil {
		c.responderError(err)
		return
	}

	// 3. Traer los movimientos del usuario y las categorías (crud_finanzas)
	movimientos, err := consultarMovimientos(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Unir cada movimiento con su categoría, aplicar los filtros y sumar
	lista := filtrarMovimientos(armarMovimientos(movimientos, categorias), filtros)

	c.responder(http.StatusOK, RespuestaLibroMayor{Movimientos: lista, Totales: calcularTotales(lista)})
}

// PostMovimiento crea un ingreso o un gasto nuevo.
//
//	Ruta:     POST /v1/movimientos
//	Recibe:   DatosMovimiento (ver sección 2)
//	Devuelve: el movimiento creado, con su id (código 201 = creado)
//	Errores:  400 JSON mal formado, 422 datos inválidos o categoría que no existe,
//	          404 usuario que no existe.
func (c *FinanzasController) PostMovimiento() {

	// 1. Leer el JSON que envió el frontend y llenar la plantilla "datos"
	var datos DatosMovimiento
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar: que venga el usuario y que los datos cumplan las reglas
	if err := validarIdUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	if err := validarMovimiento(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Revisar que el usuario exista y buscar la categoría por su nombre
	if _, err := consultarUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	categoria, err := buscarCategoria(categorias, datos.Categoria)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Pasar los datos al formato de la tabla y guardarlos en crud_finanzas.
	//    Se envía "&movimiento" para que, al guardar, se le ponga el id que asignó la base.
	movimiento := nuevoMovimiento(datos.IdUsuario, datos, categoria)
	if err := crearEnCrud("movimiento_ingreso_egreso", &movimiento); err != nil {
		c.responderError(err)
		return
	}

	// 5. Responder el movimiento creado (201 = creado)
	c.responder(http.StatusCreated, armarMovimiento(movimiento, mapaCategorias(categorias)))
}

// PutMovimiento edita un movimiento que ya existe. Usa las mismas reglas que al crear.
//
//	Ruta:     PUT /v1/movimientos/:id      (:id = id del movimiento)
//	Recibe:   DatosMovimiento. Si no vienen metodo_pago u observaciones,
//	          se dejan como estaban.
//	Devuelve: el movimiento actualizado
//	Errores:  400, 422 y 404 si el movimiento no existe o está eliminado.
func (c *FinanzasController) PutMovimiento() {

	// 1. Leer el id del movimiento (de la URL) y el JSON enviado
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	var datos DatosMovimiento
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar los datos
	if err := validarMovimiento(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Buscar el movimiento como está hoy y la categoría nueva
	movimiento, err := consultarMovimiento(id)
	if err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	categoria, err := buscarCategoria(categorias, datos.Categoria)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Armar el movimiento con los cambios. Se conserva el mismo dueño y el mismo id.
	cambios := nuevoMovimiento(movimiento.IdUsuario, datos, categoria)
	cambios.Id = movimiento.Id
	// Si el frontend no envió estos campos (nil), se dejan los valores que ya tenía
	if datos.MetodoPago == nil {
		cambios.MetodoPago = movimiento.MetodoPago
	}
	if datos.Observaciones == nil {
		cambios.Observaciones = movimiento.Observaciones
	}

	// 5. Guardar los cambios en crud_finanzas
	if err := actualizarEnCrud("movimiento_ingreso_egreso", id, cambios); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarMovimiento(cambios, mapaCategorias(categorias)))
}

// DeleteMovimiento "elimina" un movimiento.
// En realidad no lo borra de la base: lo marca con activo = false,
// así deja de aparecer en todas las consultas pero no se pierde el dato.
//
//	Ruta:     DELETE /v1/movimientos/:id
//	Devuelve: {"id": 9, "mensaje": "Movimiento eliminado"}
//	Errores:  400 id inválido, 404 si no existe.
func (c *FinanzasController) DeleteMovimiento() {

	// 1. Leer el id de la URL
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Buscar el movimiento
	movimiento, err := consultarMovimiento(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Marcarlo como inactivo y guardarlo
	movimiento.Activo = false
	if err := actualizarEnCrud("movimiento_ingreso_egreso", id, movimiento); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, map[string]interface{}{"id": id, "mensaje": "Movimiento eliminado"})
}

// PostImportarMovimientos hace la carga masiva desde la plantilla de Excel.
// Funciona en dos pasos:
//
//   - modo "validar":   revisa todas las filas y dice cuáles están bien y cuáles
//     tienen errores. NO guarda nada.
//
//   - modo "confirmar": guarda las filas que están bien. Si una falla al
//     guardarse, borra las que ya había guardado (rollback), para que no
//     queden importaciones a medias.
//
// Datos de la ruta:
//
//	Ruta:     POST /v1/movimientos/importar
//	Recibe:   DatosImportacion
//	Devuelve: RespuestaImportacion (200 al validar, 201 al confirmar)
func (c *FinanzasController) PostImportarMovimientos() {

	// 1. Leer el JSON enviado
	var datos DatosImportacion
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Revisar el modo, el usuario y que haya filas
	datos.Modo = strings.ToLower(strings.TrimSpace(datos.Modo)) // " Validar " → "validar"
	if datos.Modo != "validar" && datos.Modo != "confirmar" {
		c.responderError(invalido("El modo debe ser validar o confirmar"))
		return
	}

	if err := validarIdUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	if len(datos.Filas) == 0 {
		c.responderError(invalido("No hay filas para importar"))
		return
	}

	if _, err := consultarUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Revisar fila por fila: las buenas van a "validos" y las malas a "errores"
	validos := []MovimientoIngresoEgreso{}
	errores := []RespuestaErrorFila{}
	hoy := time.Now().Format("2006-01-02") // fecha de hoy en texto, ej. "2026-10-11"

	for i, fila := range datos.Filas {

		// Si la fila no trae fecha, se usa la de hoy
		if strings.TrimSpace(fila.Fecha) == "" {
			fila.Fecha = hoy
		}

		// Si la fila no cumple las reglas, se anota el error y se pasa a la siguiente
		// (i + 1 porque "i" empieza en 0 y las filas se cuentan desde 1)
		if err := validarMovimiento(&fila); err != nil {
			errores = append(errores, RespuestaErrorFila{Fila: i + 1, Concepto: fila.Concepto, Motivo: err.Error()})
			continue
		}

		// Si la categoría no existe, también es un error
		categoria, err := buscarCategoria(categorias, fila.Categoria)
		if err != nil {
			errores = append(errores, RespuestaErrorFila{Fila: i + 1, Concepto: fila.Concepto, Motivo: err.Error()})
			continue
		}

		// La fila está bien: se prepara para guardarla
		validos = append(validos, nuevoMovimiento(datos.IdUsuario, fila, categoria))
	}

	respuesta := RespuestaImportacion{Modo: datos.Modo, Validos: len(validos), Errores: errores}

	// En modo "validar" solo se informa el resultado, no se guarda nada
	if datos.Modo == "validar" {
		c.responder(http.StatusOK, respuesta)
		return
	}

	if len(validos) == 0 {
		c.responderError(invalido("Ninguna fila es valida, no se importo nada"))
		return
	}

	// 4. Modo "confirmar": guardar una por una las filas válidas.
	//    Se van anotando los ids guardados para poder deshacer si algo falla.
	guardados := []int{}

	for i := range validos {
		if err := crearEnCrud("movimiento_ingreso_egreso", &validos[i]); err != nil {
			// Rollback: borrar las que ya se habían guardado
			for _, idGuardado := range guardados {
				borrarEnCrud("movimiento_ingreso_egreso", idGuardado)
			}
			c.responderError(errorFinanzas{http.StatusBadGateway, "No se pudo guardar el movimiento \"" + validos[i].Nombre + "\". No se importo ninguna fila"})
			return
		}
		guardados = append(guardados, validos[i].Id)
	}

	// 5. Responder cuántas se guardaron (201 = creado)
	respuesta.Guardados = len(guardados)
	c.responder(http.StatusCreated, respuesta)
}

// ----------------------------------------------------------------------------
// METAS (tabla meta)
// ----------------------------------------------------------------------------

// GetMetasUsuario devuelve las metas de un usuario con su porcentaje, lo que
// les falta y si están cumplidas, más los totales de la pantalla de Metas.
//
//	Ruta:     GET /v1/metas/usuario/:id
//	Devuelve: RespuestaMetas
//	Errores:  400 id inválido, 404 usuario que no existe.
func (c *FinanzasController) GetMetasUsuario() {

	// 1. Leer el id del usuario de la URL
	idUsuario, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Revisar que el usuario exista (crud_auth)
	if _, err := consultarUsuario(idUsuario); err != nil {
		c.responderError(err)
		return
	}

	// 3. Traer sus metas (crud_finanzas)
	metas, err := consultarMetas(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Calcular porcentajes y totales, y responder
	c.responder(http.StatusOK, armarMetas(metas))
}

// PostMeta crea una meta de ahorro nueva.
//
//	Ruta:     POST /v1/metas
//	Recibe:   DatosMeta, ej. {"id_usuario": 1, "nombre": "Viaje", "actual": 0, "objetivo": 2000000}
//	Devuelve: la meta creada (201)
//	Errores:  400, 422 datos inválidos, 404 usuario que no existe.
func (c *FinanzasController) PostMeta() {

	// 1. Leer el JSON enviado
	var datos DatosMeta
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar los datos
	if err := validarIdUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	if err := validarMeta(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Revisar que el usuario exista (crud_auth)
	if _, err := consultarUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	// 4. Armar la meta con el formato de la tabla y guardarla.
	//    El formulario no usa editar_meta, por eso IdEditarMeta queda vacío (nil).
	meta := Meta{
		IdUsuario:     datos.IdUsuario,
		Nombre:        datos.Nombre,
		Icono:         datos.Icono,
		MontoObjetivo: datos.Objetivo,
		MontoActual:   datos.Actual,
		Activo:        true,
	}

	if err := crearEnCrud("meta", &meta); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusCreated, armarMeta(meta))
}

// PutMeta edita el nombre, el ícono, el monto actual y el objetivo de una meta.
//
//	Ruta:     PUT /v1/metas/:id
//	Recibe:   DatosMeta
//	Devuelve: la meta actualizada
//	Errores:  400, 422, 404 si la meta no existe.
func (c *FinanzasController) PutMeta() {

	// 1. Leer el id de la meta y el JSON enviado
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	var datos DatosMeta
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar los datos
	if err := validarMeta(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Buscar la meta como está hoy
	meta, err := consultarMeta(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Cambiar solo los campos del formulario (los demás quedan igual) y guardar
	meta.Nombre = datos.Nombre
	meta.MontoObjetivo = datos.Objetivo
	meta.MontoActual = datos.Actual
	meta.Icono = datos.Icono

	if err := actualizarEnCrud("meta", id, meta); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarMeta(meta))
}

// PutAbonoMeta suma un abono (dinero) a lo que lleva ahorrado una meta.
//
//	Ruta:     PUT /v1/metas/:id/abono
//	Recibe:   {"monto": 150000}
//	Devuelve: la meta con el nuevo monto y su porcentaje
//	Errores:  422 si el monto es 0 o negativo, 404 si la meta no existe.
func (c *FinanzasController) PutAbonoMeta() {

	// 1. Leer el id de la meta y el JSON enviado
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	var datos DatosAbono
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. El abono tiene que ser mayor que 0
	if datos.Monto <= 0 {
		c.responderError(invalido("El monto del abono debe ser mayor a 0"))
		return
	}

	// 3. Buscar la meta
	meta, err := consultarMeta(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Sumar el abono a lo que ya llevaba y guardar
	meta.MontoActual += datos.Monto

	if err := actualizarEnCrud("meta", id, meta); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarMeta(meta))
}

// DeleteMeta "elimina" una meta (la marca con activo = false, no la borra).
//
//	Ruta:     DELETE /v1/metas/:id
//	Devuelve: {"id": 6, "mensaje": "Meta eliminada"}
//	Errores:  400 id inválido, 404 si no existe.
func (c *FinanzasController) DeleteMeta() {

	// 1. Leer el id de la URL
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Buscar la meta
	meta, err := consultarMeta(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Marcarla como inactiva y guardar
	meta.Activo = false
	if err := actualizarEnCrud("meta", id, meta); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, map[string]interface{}{"id": id, "mensaje": "Meta eliminada"})
}

// ----------------------------------------------------------------------------
// INVERSIONES (tablas inversion + nivel_riesgo + historial_inversion)
// ----------------------------------------------------------------------------

// GetInversionesUsuario devuelve las inversiones de un usuario y todo lo que
// muestra la pantalla de Inversiones: total invertido, rendimiento, reparto
// por riesgo, alerta de concentración y gráfica de crecimiento.
//
//	Ruta:     GET /v1/inversiones/usuario/:id
//	Devuelve: RespuestaInversiones
//	Errores:  400 id inválido, 404 usuario que no existe.
func (c *FinanzasController) GetInversionesUsuario() {

	// 1. Leer el id del usuario de la URL
	idUsuario, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Revisar que el usuario exista (crud_auth)
	if _, err := consultarUsuario(idUsuario); err != nil {
		c.responderError(err)
		return
	}

	// 3. Traer sus inversiones, los niveles de riesgo y el historial (crud_finanzas)
	inversiones, err := consultarInversiones(idUsuario)
	if err != nil {
		c.responderError(err)
		return
	}

	niveles, err := consultarNivelesRiesgo()
	if err != nil {
		c.responderError(err)
		return
	}

	historial := consultarHistorial()

	// 4. Calcular el resumen y responder
	c.responder(http.StatusOK, armarInversiones(inversiones, niveles, historial))
}

// PostInversion crea una inversión nueva.
//
//	Ruta:     POST /v1/inversiones
//	Recibe:   DatosInversion, ej. {"id_usuario": 4, "nombre": "CDT", "monto": 1000000,
//	          "rendimiento": 90000, "riesgo": "Bajo", "duracion": "1 año"}
//	Devuelve: la inversión creada (201)
//	Errores:  400, 422 datos inválidos o riesgo que no existe, 404 usuario.
func (c *FinanzasController) PostInversion() {

	// 1. Leer el JSON enviado
	var datos DatosInversion
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar los datos
	if err := validarIdUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	if err := validarInversion(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Revisar que el usuario exista y buscar el nivel de riesgo por su nombre
	if _, err := consultarUsuario(datos.IdUsuario); err != nil {
		c.responderError(err)
		return
	}

	niveles, err := consultarNivelesRiesgo()
	if err != nil {
		c.responderError(err)
		return
	}

	nivel, err := buscarNivelRiesgo(niveles, datos.Riesgo)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Armar la inversión con el formato de la tabla y guardarla.
	//    La rentabilidad (%) se calcula a partir del rendimiento en pesos.
	inversion := Inversion{
		IdUsuario:     datos.IdUsuario,
		IdNivelRiesgo: &Ref{Id: nivel.Id},
		Nombre:        datos.Nombre,
		Monto:         datos.Monto,
		Rendimiento:   datos.Rendimiento,
		Rentabilidad:  calcularRentabilidad(datos.Rendimiento, datos.Monto),
		Duracion:      datos.Duracion,
		FechaInicio:   time.Now(), // empieza hoy
		Activo:        true,
	}

	if err := crearEnCrud("inversion", &inversion); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusCreated, armarInversion(inversion, mapaNiveles(niveles)))
}

// PutInversion edita una inversión que ya existe.
//
//	Ruta:     PUT /v1/inversiones/:id
//	Recibe:   DatosInversion
//	Devuelve: la inversión actualizada
//	Errores:  400, 422, 404 si la inversión no existe.
func (c *FinanzasController) PutInversion() {

	// 1. Leer el id de la inversión y el JSON enviado
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	var datos DatosInversion
	if err := c.leerDatos(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 2. Validar los datos
	if err := validarInversion(&datos); err != nil {
		c.responderError(err)
		return
	}

	// 3. Buscar la inversión como está hoy y el nivel de riesgo nuevo
	inversion, err := consultarInversion(id)
	if err != nil {
		c.responderError(err)
		return
	}

	niveles, err := consultarNivelesRiesgo()
	if err != nil {
		c.responderError(err)
		return
	}

	nivel, err := buscarNivelRiesgo(niveles, datos.Riesgo)
	if err != nil {
		c.responderError(err)
		return
	}

	// 4. Cambiar los campos del formulario (los demás quedan igual) y guardar
	inversion.Nombre = datos.Nombre
	inversion.Monto = datos.Monto
	inversion.Rendimiento = datos.Rendimiento
	inversion.Rentabilidad = calcularRentabilidad(datos.Rendimiento, datos.Monto)
	inversion.Duracion = datos.Duracion
	inversion.IdNivelRiesgo = &Ref{Id: nivel.Id}

	if err := actualizarEnCrud("inversion", id, inversion); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarInversion(inversion, mapaNiveles(niveles)))
}

// DeleteInversion "elimina" una inversión (la marca con activo = false, no la borra).
//
//	Ruta:     DELETE /v1/inversiones/:id
//	Devuelve: {"id": 6, "mensaje": "Inversion eliminada"}
//	Errores:  400 id inválido, 404 si no existe.
func (c *FinanzasController) DeleteInversion() {

	// 1. Leer el id de la URL
	id, err := c.idDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Buscar la inversión
	inversion, err := consultarInversion(id)
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Marcarla como inactiva y guardar
	inversion.Activo = false
	if err := actualizarEnCrud("inversion", id, inversion); err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, map[string]interface{}{"id": id, "mensaje": "Inversion eliminada"})
}

// ----------------------------------------------------------------------------
// ADMIN (datos de TODOS los usuarios)
// Estas rutas reutilizan los mismos cálculos de arriba, pero sin filtrar por
// usuario: por eso se pasa 0 como id de usuario a las consultas.
// ----------------------------------------------------------------------------

// GetAdminMovimientos devuelve los movimientos de todos los usuarios y sus totales.
//
//	Ruta:     GET /v1/admin/movimientos?tipo=&categoria=&desde=&hasta=&q=
//	Devuelve: RespuestaLibroMayor
func (c *FinanzasController) GetAdminMovimientos() {

	// 1. Leer los filtros de la URL
	filtros, err := c.filtrosDeLaURL()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Traer los movimientos de todos (0 = sin filtrar por usuario) y las categorías
	movimientos, err := consultarMovimientos(0)
	if err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Unir con la categoría, filtrar y sumar
	lista := filtrarMovimientos(armarMovimientos(movimientos, categorias), filtros)

	c.responder(http.StatusOK, RespuestaLibroMayor{Movimientos: lista, Totales: calcularTotales(lista)})
}

// GetAdminMetas devuelve las metas de todos los usuarios y sus totales.
//
//	Ruta:     GET /v1/admin/metas
//	Devuelve: RespuestaMetas
func (c *FinanzasController) GetAdminMetas() {

	// Traer las metas de todos (0 = sin filtrar por usuario)
	metas, err := consultarMetas(0)
	if err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarMetas(metas))
}

// GetAdminInversiones devuelve las inversiones de todos los usuarios y sus totales.
//
//	Ruta:     GET /v1/admin/inversiones
//	Devuelve: RespuestaInversiones
func (c *FinanzasController) GetAdminInversiones() {

	// Traer las inversiones de todos (0 = sin filtrar por usuario)
	inversiones, err := consultarInversiones(0)
	if err != nil {
		c.responderError(err)
		return
	}

	niveles, err := consultarNivelesRiesgo()
	if err != nil {
		c.responderError(err)
		return
	}

	c.responder(http.StatusOK, armarInversiones(inversiones, niveles, consultarHistorial()))
}

// GetAdminLibroMayor devuelve el libro mayor agrupado por usuario, con el
// nombre de cada uno y los totales de toda la plataforma.
//
//	Ruta:     GET /v1/admin/libro-mayor
//	Devuelve: RespuestaLibrosAdmin
func (c *FinanzasController) GetAdminLibroMayor() {

	// 1. Traer los nombres de todos los usuarios (crud_auth)
	nombres, err := consultarNombresUsuarios()
	if err != nil {
		c.responderError(err)
		return
	}

	// 2. Traer todos los movimientos y las categorías (crud_finanzas)
	movimientos, err := consultarMovimientos(0)
	if err != nil {
		c.responderError(err)
		return
	}

	categorias, err := consultarCategorias()
	if err != nil {
		c.responderError(err)
		return
	}

	// 3. Agrupar los movimientos por usuario y responder
	c.responder(http.StatusOK, armarLibrosAdmin(armarMovimientos(movimientos, categorias), nombres))
}

// ============================================================================
// 5. VALIDACIONES
// ----------------------------------------------------------------------------
// Revisan que los datos cumplan las reglas antes de guardarlos.
// Si algo está mal, devuelven un error con código 422 ("datos inválidos") y
// un mensaje claro que el frontend le puede mostrar al usuario.
// Si todo está bien, devuelven nil.
// ============================================================================

// invalido crea un error 422 con el mensaje indicado.
func invalido(mensaje string) error {
	return errorFinanzas{http.StatusUnprocessableEntity, mensaje}
}

// validarIdUsuario revisa que el formulario traiga un id de usuario (mayor que 0).
func validarIdUsuario(idUsuario int) error {
	if idUsuario <= 0 {
		return invalido("El id_usuario es obligatorio")
	}
	return nil
}

// validarMovimiento revisa los datos de un movimiento.
// También los "limpia": quita espacios sobrantes y pasa el tipo a minúsculas.
// Recibe un puntero (*DatosMovimiento) para poder hacer esa limpieza sobre los datos originales.
func validarMovimiento(datos *DatosMovimiento) error {
	// Limpiar: " Mercado " → "Mercado", "Gasto" → "gasto"
	datos.Concepto = strings.TrimSpace(datos.Concepto)
	datos.Categoria = strings.TrimSpace(datos.Categoria)
	datos.Tipo = strings.ToLower(strings.TrimSpace(datos.Tipo))
	datos.Fecha = strings.TrimSpace(datos.Fecha)

	// Reglas (se revisan en orden; se devuelve el primer error que se encuentre)
	if datos.Concepto == "" {
		return invalido("El concepto es obligatorio")
	}
	if utf8.RuneCountInString(datos.Concepto) > 120 { // en la base el concepto tiene máximo 120 letras
		return invalido("El concepto no puede tener mas de 120 caracteres")
	}
	if datos.Valor <= 0 {
		return invalido("El valor debe ser mayor a 0")
	}
	if datos.Tipo != "ingreso" && datos.Tipo != "gasto" {
		return invalido("El tipo debe ser ingreso o gasto")
	}
	if datos.Categoria == "" {
		return invalido("La categoria es obligatoria")
	}
	// time.Parse intenta leer la fecha con el formato AAAA-MM-DD.
	// (En Go el formato se escribe con la fecha de ejemplo 2006-01-02.)
	if _, err := time.Parse("2006-01-02", datos.Fecha); err != nil {
		return invalido("La fecha debe tener el formato AAAA-MM-DD")
	}
	if datos.MetodoPago != nil && utf8.RuneCountInString(*datos.MetodoPago) > 50 {
		return invalido("El metodo de pago no puede tener mas de 50 caracteres")
	}
	return nil // todo bien
}

// validarMeta revisa los datos de una meta. Si no trae ícono, le pone 🎯.
func validarMeta(datos *DatosMeta) error {
	// Limpiar espacios sobrantes
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Icono = strings.TrimSpace(datos.Icono)
	if datos.Icono == "" {
		datos.Icono = "🎯" // ícono por defecto
	}

	// Reglas
	if datos.Nombre == "" {
		return invalido("El nombre de la meta es obligatorio")
	}
	if utf8.RuneCountInString(datos.Nombre) > 120 {
		return invalido("El nombre no puede tener mas de 120 caracteres")
	}
	if datos.Objetivo <= 0 {
		return invalido("El monto objetivo debe ser mayor a 0")
	}
	if datos.Actual < 0 {
		return invalido("El monto actual no puede ser negativo")
	}
	if utf8.RuneCountInString(datos.Icono) > 50 {
		return invalido("El icono no puede tener mas de 50 caracteres")
	}
	return nil // todo bien
}

// validarInversion revisa los datos de una inversión.
// También deja el riesgo bien escrito: "bajo" o "BAJO" → "Bajo".
func validarInversion(datos *DatosInversion) error {
	// Limpiar
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Duracion = strings.TrimSpace(datos.Duracion)
	datos.Riesgo = capitalizar(datos.Riesgo)

	// Reglas
	if datos.Nombre == "" {
		return invalido("El nombre de la inversion es obligatorio")
	}
	if utf8.RuneCountInString(datos.Nombre) > 120 {
		return invalido("El nombre no puede tener mas de 120 caracteres")
	}
	if datos.Duracion == "" {
		return invalido("La duracion es obligatoria")
	}
	if utf8.RuneCountInString(datos.Duracion) > 30 {
		return invalido("La duracion no puede tener mas de 30 caracteres")
	}
	if datos.Monto <= 0 {
		return invalido("El monto debe ser mayor a 0")
	}
	if datos.Riesgo != "Bajo" && datos.Riesgo != "Medio" && datos.Riesgo != "Alto" {
		return invalido("El riesgo debe ser Bajo, Medio o Alto")
	}
	// No se puede perder más dinero del que se invirtió
	if datos.Rendimiento < -datos.Monto {
		return invalido("La perdida no puede ser mayor que el monto invertido")
	}
	// La base guarda la rentabilidad con máximo 4 cifras enteras (menos de 10.000 %)
	if math.Abs(calcularRentabilidad(datos.Rendimiento, datos.Monto)) >= 10000 {
		return invalido("El rendimiento no es coherente con el monto invertido")
	}
	return nil // todo bien
}

// ============================================================================
// 6. CONSULTAS A LOS CRUD
// ----------------------------------------------------------------------------
// Aquí están todas las funciones que hablan con crud_finanzas y crud_auth.
// Los métodos de arriba nunca hacen peticiones HTTP directamente: llaman
// a estas funciones.
// ============================================================================

// errorFinanzas es un error que además guarda el código HTTP con el que el
// MID le debe responder al frontend. Ejemplos:
//
//	errorFinanzas{404, "El usuario no existe"}
//	errorFinanzas{422, "El valor debe ser mayor a 0"}
type errorFinanzas struct {
	estado  int    // código HTTP: 400, 404, 409, 422, 502...
	mensaje string // texto que se le muestra al usuario
}

// Error hace que errorFinanzas funcione como cualquier error de Go
// (devuelve el mensaje cuando se pide err.Error()).
func (e errorFinanzas) Error() string {
	return e.mensaje
}

// noExiste dice si un error es de tipo "404 no existe".
func noExiste(err error) bool {
	e, ok := err.(errorFinanzas) // ¿el error es un errorFinanzas?
	return ok && e.estado == http.StatusNotFound
}

// nombreDelCrud devuelve un nombre amigable del CRUD según la URL,
// para usarlo en los mensajes de error.
func nombreDelCrud(url string) string {
	if strings.HasPrefix(url, urlCrudUsuarios) {
		return "la API de usuarios"
	}
	return "la API de finanzas"
}

// consultarCrud hace un GET a un CRUD y llena "destino" con el JSON que responde.
// "destino" debe ser una variable pasada con &, por ejemplo: consultarCrud(url, &metas)
//
// Posibles errores:
//   - 502 si el CRUD está apagado o responde con un error.
//   - 404 si el registro pedido no existe.
func consultarCrud(url string, destino interface{}) error {
	// 1. Hacer la petición GET
	response, err := http.Get(url)
	if err != nil {
		// El CRUD no contestó (por ejemplo, está apagado)
		return errorFinanzas{http.StatusBadGateway, "No fue posible comunicarse con " + nombreDelCrud(url)}
	}
	// Cerrar la respuesta al terminar la función (siempre hay que hacerlo)
	defer response.Body.Close()

	// 2. Leer todo el cuerpo de la respuesta
	body, _ := io.ReadAll(response.Body)

	// 3. Si el código no es 200, algo falló en el CRUD
	if response.StatusCode != http.StatusOK {
		return errorFinanzas{http.StatusBadGateway, "Error en " + nombreDelCrud(url) + " (codigo " + strconv.Itoa(response.StatusCode) + ")"}
	}

	// 4. Convertir el JSON a la plantilla de Go
	if err := json.Unmarshal(body, destino); err != nil {
		// Los CRUD hechos con bee responden los errores como un texto y con código 200.
		// Si ese texto dice "no row found", es que el registro no existe.
		if strings.Contains(string(body), "no row found") {
			return errorFinanzas{http.StatusNotFound, "No existe el registro"}
		}
		return errorFinanzas{http.StatusBadGateway, "Error en " + nombreDelCrud(url) + ": " + string(body)}
	}
	return nil // todo bien: "destino" ya tiene los datos
}

// enviarCrud hace un POST (crear), PUT (actualizar) o DELETE (borrar) a crud_finanzas.
//
//   - POST: el CRUD responde 201 con el registro creado. Ese registro se copia
//     de vuelta en "registro", así el MID conoce el id que le dio la base de datos.
//   - PUT y DELETE: el CRUD responde el texto "OK" si todo salió bien.
func enviarCrud(metodo string, url string, registro interface{}) error {
	// 1. Convertir el registro a JSON (DELETE no lleva cuerpo, por eso puede ser nil)
	var cuerpo io.Reader
	if registro != nil {
		jsonRegistro, _ := json.Marshal(registro)
		cuerpo = bytes.NewBuffer(jsonRegistro)
	}

	// 2. Armar la petición y decir que el cuerpo es JSON
	peticion, _ := http.NewRequest(metodo, url, cuerpo)
	peticion.Header.Set("Content-Type", "application/json")

	// 3. Enviarla
	response, err := http.DefaultClient.Do(peticion)
	if err != nil {
		return errorFinanzas{http.StatusBadGateway, "No fue posible comunicarse con " + nombreDelCrud(url)}
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	// 4a. Si era un POST, el CRUD debe responder 201 (creado)
	if metodo == http.MethodPost {
		if response.StatusCode != http.StatusCreated {
			return errorFinanzas{http.StatusBadGateway, "La API de finanzas no guardo el registro: " + string(body)}
		}
		// Copiar lo que respondió el CRUD (ya trae el id nuevo) en "registro"
		json.Unmarshal(body, registro)
		return nil
	}

	// 4b. Si era PUT o DELETE, el CRUD debe responder "OK"
	if strings.TrimSpace(string(body)) != `"OK"` {
		return errorFinanzas{http.StatusBadGateway, "La API de finanzas no actualizo el registro: " + string(body)}
	}
	return nil
}

// crearEnCrud guarda un registro nuevo en una tabla de finanzas.
// Ejemplo: crearEnCrud("meta", &meta)
func crearEnCrud(tabla string, registro interface{}) error {
	return enviarCrud(http.MethodPost, urlCrudFinanzas+"/"+tabla, registro)
}

// actualizarEnCrud reemplaza un registro existente (identificado por su id).
// Ejemplo: actualizarEnCrud("meta", 6, meta)
func actualizarEnCrud(tabla string, id int, registro interface{}) error {
	return enviarCrud(http.MethodPut, fmt.Sprintf("%s/%s/%d", urlCrudFinanzas, tabla, id), registro)
}

// borrarEnCrud borra un registro de verdad (no solo lo desactiva).
// Solo se usa para deshacer una importación de Excel que falló a la mitad.
func borrarEnCrud(tabla string, id int) error {
	return enviarCrud(http.MethodDelete, fmt.Sprintf("%s/%s/%d", urlCrudFinanzas, tabla, id), nil)
}

// filtroActivos arma la parte final de la URL para pedir solo registros activos.
//   - Con un id de usuario:  ?query=id_usuario:5,activo:true&limit=10000
//   - Con 0 (para el admin): ?query=activo:true&limit=10000   (de todos los usuarios)
func filtroActivos(idUsuario int) string {
	if idUsuario > 0 {
		return fmt.Sprintf("?query=id_usuario:%d,activo:true&limit=%d", idUsuario, limiteCrud)
	}
	return fmt.Sprintf("?query=activo:true&limit=%d", limiteCrud)
}

// --- Consultas a crud_auth (usuarios) ---

// consultarUsuario trae un usuario por su id. Si no existe, devuelve un error 404.
func consultarUsuario(id int) (Usuario, error) {
	var usuario Usuario
	err := consultarCrud(fmt.Sprintf("%s/usuario/%d", urlCrudUsuarios, id), &usuario)
	if noExiste(err) {
		return usuario, errorFinanzas{http.StatusNotFound, "El usuario no existe"}
	}
	return usuario, err
}

// consultarNombresUsuarios trae todos los usuarios y devuelve un diccionario
// id → "Nombre Apellido", por ejemplo {1: "Harold Arciniegas", 2: "Fabio Zorro"}.
func consultarNombresUsuarios() (map[int]string, error) {
	var usuarios []Usuario
	if err := consultarCrud(fmt.Sprintf("%s/usuario?limit=%d", urlCrudUsuarios, limiteCrud), &usuarios); err != nil {
		return nil, err
	}
	nombres := map[int]string{}
	for _, usuario := range usuarios {
		nombres[usuario.Id] = nombreCompleto(usuario)
	}
	return nombres, nil
}

// --- Consultas a crud_finanzas ---
// Las que terminan en "s" (consultarMetas) traen una LISTA.
// Las que no (consultarMeta) traen UN registro por su id y devuelven 404
// si no existe o si está eliminado (activo = false).

// consultarCategorias trae todas las categorías activas.
func consultarCategorias() ([]Categoria, error) {
	var categorias []Categoria
	err := consultarCrud(urlCrudFinanzas+"/categoria"+filtroActivos(0), &categorias)
	return categorias, err
}

// consultarMovimientos trae los movimientos activos de un usuario (o de todos si idUsuario es 0).
func consultarMovimientos(idUsuario int) ([]MovimientoIngresoEgreso, error) {
	var movimientos []MovimientoIngresoEgreso
	err := consultarCrud(urlCrudFinanzas+"/movimiento_ingreso_egreso"+filtroActivos(idUsuario), &movimientos)
	return movimientos, err
}

// consultarMovimiento trae un movimiento por su id.
func consultarMovimiento(id int) (MovimientoIngresoEgreso, error) {
	var movimiento MovimientoIngresoEgreso
	err := consultarCrud(fmt.Sprintf("%s/movimiento_ingreso_egreso/%d", urlCrudFinanzas, id), &movimiento)
	if noExiste(err) || (err == nil && !movimiento.Activo) {
		return movimiento, errorFinanzas{http.StatusNotFound, "El movimiento no existe"}
	}
	return movimiento, err
}

// consultarMetas trae las metas activas de un usuario (o de todos si idUsuario es 0).
func consultarMetas(idUsuario int) ([]Meta, error) {
	var metas []Meta
	err := consultarCrud(urlCrudFinanzas+"/meta"+filtroActivos(idUsuario), &metas)
	return metas, err
}

// consultarMeta trae una meta por su id.
func consultarMeta(id int) (Meta, error) {
	var meta Meta
	err := consultarCrud(fmt.Sprintf("%s/meta/%d", urlCrudFinanzas, id), &meta)
	if noExiste(err) || (err == nil && !meta.Activo) {
		return meta, errorFinanzas{http.StatusNotFound, "La meta no existe"}
	}
	return meta, err
}

// consultarInversiones trae las inversiones activas de un usuario (o de todos si idUsuario es 0).
func consultarInversiones(idUsuario int) ([]Inversion, error) {
	var inversiones []Inversion
	err := consultarCrud(urlCrudFinanzas+"/inversion"+filtroActivos(idUsuario), &inversiones)
	return inversiones, err
}

// consultarInversion trae una inversión por su id.
func consultarInversion(id int) (Inversion, error) {
	var inversion Inversion
	err := consultarCrud(fmt.Sprintf("%s/inversion/%d", urlCrudFinanzas, id), &inversion)
	if noExiste(err) || (err == nil && !inversion.Activo) {
		return inversion, errorFinanzas{http.StatusNotFound, "La inversion no existe"}
	}
	return inversion, err
}

// consultarNivelesRiesgo trae los niveles de riesgo (Bajo, Medio, Alto).
func consultarNivelesRiesgo() ([]NivelRiesgo, error) {
	var niveles []NivelRiesgo
	err := consultarCrud(urlCrudFinanzas+"/nivel_riesgo"+filtroActivos(0), &niveles)
	return niveles, err
}

// consultarHistorial trae el historial de valor de las inversiones.
// A diferencia de las demás, esta NO devuelve error: si falla (por ejemplo,
// si crud_finanzas todavía no tiene la tabla historial_inversion), devuelve
// una lista vacía y la gráfica de crecimiento simplemente sale vacía.
func consultarHistorial() []HistorialInversion {
	var historial []HistorialInversion
	if err := consultarCrud(urlCrudFinanzas+"/historial_inversion"+filtroActivos(0), &historial); err != nil {
		return []HistorialInversion{}
	}
	return historial
}

// buscarCategoria busca una categoría por su nombre, sin importar mayúsculas
// ni tildes ("alimentación" encuentra "Alimentacion"). Si no existe, error 422.
func buscarCategoria(categorias []Categoria, nombre string) (Categoria, error) {
	for _, categoria := range categorias {
		if normalizar(categoria.Nombre) == normalizar(nombre) {
			return categoria, nil // la encontró
		}
	}
	return Categoria{}, invalido("La categoria \"" + nombre + "\" no existe")
}

// buscarNivelRiesgo busca un nivel de riesgo por su nombre ("bajo" encuentra "Bajo").
// Si no existe, error 422.
func buscarNivelRiesgo(niveles []NivelRiesgo, nombre string) (NivelRiesgo, error) {
	for _, nivel := range niveles {
		if normalizar(nivel.Nombre) == normalizar(nombre) {
			return nivel, nil // lo encontró
		}
	}
	return NivelRiesgo{}, invalido("El nivel de riesgo \"" + nombre + "\" no existe")
}

// ============================================================================
// 7. CÁLCULOS
// ----------------------------------------------------------------------------
// Funciones que transforman los datos de las tablas al formato del frontend
// y hacen las sumas y porcentajes. No hablan con ningún CRUD.
// ============================================================================

// porcentaje calcula qué % es "parte" de "total", redondeado a número entero.
// Ejemplo: porcentaje(25, 200) = 13. Si el total es 0 devuelve 0 (para no dividir entre 0).
func porcentaje(parte, total float64) int {
	if total == 0 {
		return 0
	}
	return int(math.Round(parte / total * 100))
}

// porcentajeDecimal es igual que porcentaje pero con un decimal.
// Ejemplo: porcentajeDecimal(125000, 1000000) = 12.5
func porcentajeDecimal(parte, total float64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(parte/total*1000) / 10
}

// sinTildes reemplaza cada letra con tilde por la misma letra sin tilde.
var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// normalizar deja un texto listo para compararlo con otro:
// sin espacios a los lados, en minúsculas y sin tildes.
// Ejemplo: " Alimentación " → "alimentacion"
func normalizar(texto string) string {
	return sinTildes.Replace(strings.ToLower(strings.TrimSpace(texto)))
}

// capitalizar pone en mayúscula la primera letra de cada palabra.
// Ejemplo: "harold arciniegas" → "Harold Arciniegas"
func capitalizar(texto string) string {
	palabras := strings.Fields(strings.ToLower(texto)) // separa el texto en palabras
	for i, palabra := range palabras {
		letras := []rune(palabra)              // las letras de la palabra
		letras[0] = unicode.ToUpper(letras[0]) // la primera, en mayúscula
		palabras[i] = string(letras)
	}
	return strings.Join(palabras, " ") // vuelve a unir las palabras con espacios
}

// nombreCompleto une el nombre y el apellido de un usuario, bien escritos.
func nombreCompleto(usuario Usuario) string {
	return capitalizar(usuario.Nombre + " " + usuario.Apellido)
}

// --- Libro mayor ---

// mapaCategorias convierte la lista de categorías en un diccionario id → categoría,
// para encontrar rápido la categoría de cada movimiento.
func mapaCategorias(categorias []Categoria) map[int]Categoria {
	mapa := map[int]Categoria{}
	for _, categoria := range categorias {
		mapa[categoria.Id] = categoria
	}
	return mapa
}

// nuevoMovimiento pasa los datos del formulario al formato de la tabla.
// Ejemplo: tipo "ingreso" → EsIngreso true; concepto → Nombre; valor → Monto.
func nuevoMovimiento(idUsuario int, datos DatosMovimiento, categoria Categoria) MovimientoIngresoEgreso {
	// La fecha ya fue validada, así que aquí se puede convertir sin revisar el error
	fecha, _ := time.Parse("2006-01-02", datos.Fecha)

	movimiento := MovimientoIngresoEgreso{
		IdUsuario:   idUsuario,
		IdCategoria: &Ref{Id: categoria.Id},
		Nombre:      datos.Concepto,
		Monto:       datos.Valor,
		EsIngreso:   datos.Tipo == "ingreso", // true si el tipo es "ingreso"
		Fecha:       fecha,
		Activo:      true,
	}
	// Los campos opcionales solo se copian si vinieron en el formulario
	if datos.MetodoPago != nil {
		movimiento.MetodoPago = strings.TrimSpace(*datos.MetodoPago)
	}
	if datos.Observaciones != nil {
		movimiento.Observaciones = strings.TrimSpace(*datos.Observaciones)
	}
	return movimiento
}

// armarMovimiento hace lo contrario: pasa una fila de la tabla al formato del
// frontend, con el nombre y el ícono de su categoría.
func armarMovimiento(movimiento MovimientoIngresoEgreso, categorias map[int]Categoria) RespuestaMovimiento {
	// Valores por defecto (si el movimiento no tiene categoría, se quedan así)
	respuesta := RespuestaMovimiento{
		Id:            movimiento.Id,
		IdUsuario:     movimiento.IdUsuario,
		Fecha:         movimiento.Fecha.Format("2006-01-02"), // fecha a texto "AAAA-MM-DD"
		Concepto:      movimiento.Nombre,
		Categoria:     "Sin categoria",
		Icono:         "💰",
		Tipo:          "gasto",
		Valor:         movimiento.Monto,
		MetodoPago:    movimiento.MetodoPago,
		Observaciones: movimiento.Observaciones,
	}

	if movimiento.EsIngreso {
		respuesta.Tipo = "ingreso"
	}

	// Si tiene categoría, se busca su nombre y su ícono
	if movimiento.IdCategoria != nil {
		if categoria, ok := categorias[movimiento.IdCategoria.Id]; ok {
			respuesta.Categoria = categoria.Nombre
			if categoria.Icono != "" {
				respuesta.Icono = categoria.Icono
			}
		}
	}
	return respuesta
}

// armarMovimientos pasa TODOS los movimientos al formato del frontend y los
// ordena del más reciente al más antiguo (si tienen la misma fecha, por id).
func armarMovimientos(movimientos []MovimientoIngresoEgreso, categorias []Categoria) []RespuestaMovimiento {
	mapa := mapaCategorias(categorias)

	lista := []RespuestaMovimiento{}
	for _, movimiento := range movimientos {
		lista = append(lista, armarMovimiento(movimiento, mapa))
	}

	// Ordenar: la función dice cuándo el elemento "i" va antes que el "j"
	sort.SliceStable(lista, func(i, j int) bool {
		if lista[i].Fecha != lista[j].Fecha {
			return lista[i].Fecha > lista[j].Fecha // fecha más nueva primero
		}
		return lista[i].Id > lista[j].Id // misma fecha: id más alto primero
	})
	return lista
}

// filtrarMovimientos deja solo los movimientos que cumplen TODOS los filtros activos.
// Los filtros vacíos no se tienen en cuenta.
func filtrarMovimientos(lista []RespuestaMovimiento, filtros FiltrosMovimientos) []RespuestaMovimiento {
	texto := normalizar(filtros.Texto)
	categoria := normalizar(filtros.Categoria)

	resultado := []RespuestaMovimiento{}
	for _, movimiento := range lista {
		// "continue" salta al siguiente movimiento sin agregar este
		if filtros.Tipo != "" && filtros.Tipo != "todos" && movimiento.Tipo != filtros.Tipo {
			continue
		}
		if categoria != "" && categoria != "todas" && normalizar(movimiento.Categoria) != categoria {
			continue
		}
		// Las fechas "AAAA-MM-DD" se pueden comparar como texto
		if filtros.Desde != "" && movimiento.Fecha < filtros.Desde {
			continue
		}
		if filtros.Hasta != "" && movimiento.Fecha > filtros.Hasta {
			continue
		}
		if texto != "" && !strings.Contains(normalizar(movimiento.Concepto), texto) {
			continue
		}
		// Pasó todos los filtros
		resultado = append(resultado, movimiento)
	}
	return resultado
}

// calcularTotales suma los movimientos y calcula las tarjetas del Libro mayor.
func calcularTotales(lista []RespuestaMovimiento) RespuestaTotales {
	totales := RespuestaTotales{} // empieza todo en 0

	for _, movimiento := range lista {
		if movimiento.Tipo == "ingreso" {
			totales.TotalIngresos += movimiento.Valor
		} else {
			totales.TotalGastos += movimiento.Valor
			if normalizar(movimiento.Categoria) == "gastos hormiga" {
				totales.GastosHormiga += movimiento.Valor
			}
		}
		if normalizar(movimiento.Categoria) == "ahorro" {
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

// calcularSalud decide el estado de "salud financiera" del panel:
//   - "Alto riesgo" si gasta más de lo que gana.
//   - "Atención"    si ahorra menos del 10 % de lo que gana.
//   - "En orden"    en cualquier otro caso.
func calcularSalud(totales RespuestaTotales) RespuestaSalud {
	if totales.Ahorro < 0 {
		return RespuestaSalud{Estado: "Alto riesgo", Mensaje: "Estás gastando más de lo que ganas."}
	}

	// Qué % de los ingresos se está ahorrando
	ahorro := 0.0
	if totales.TotalIngresos > 0 {
		ahorro = totales.Ahorro / totales.TotalIngresos * 100
	}

	if ahorro < 10 {
		return RespuestaSalud{Estado: "Atención", Mensaje: "Tu margen de ahorro es bajo, revisa tus gastos."}
	}
	return RespuestaSalud{Estado: "En orden", Mensaje: "Tus finanzas van bien, sigue así."}
}

// calcularDistribucion suma los gastos por categoría (para la dona del panel)
// y los ordena de mayor a menor.
func calcularDistribucion(lista []RespuestaMovimiento) []RespuestaParte {
	montos := map[string]float64{} // categoría → total gastado
	total := 0.0

	for _, movimiento := range lista {
		if movimiento.Tipo == "gasto" {
			montos[movimiento.Categoria] += movimiento.Valor
			total += movimiento.Valor
		}
	}

	// Pasar el diccionario a lista, calculando el % de cada categoría
	partes := []RespuestaParte{}
	for etiqueta, monto := range montos {
		partes = append(partes, RespuestaParte{Etiqueta: etiqueta, Monto: monto, Porcentaje: porcentaje(monto, total)})
	}

	// Ordenar de mayor a menor monto (si empatan, por nombre)
	sort.Slice(partes, func(i, j int) bool {
		if partes[i].Monto != partes[j].Monto {
			return partes[i].Monto > partes[j].Monto
		}
		return partes[i].Etiqueta < partes[j].Etiqueta
	})
	return partes
}

// primerosMovimientos devuelve solo los primeros "cantidad" movimientos de la lista.
func primerosMovimientos(lista []RespuestaMovimiento, cantidad int) []RespuestaMovimiento {
	if len(lista) > cantidad {
		return lista[:cantidad] // del primero hasta "cantidad"
	}
	return lista
}

// armarLibrosAdmin agrupa los movimientos por usuario para la vista del admin.
// En cada libro las filas van de la más antigua a la más reciente, y al final
// se calculan los totales de cada usuario y de toda la plataforma.
func armarLibrosAdmin(lista []RespuestaMovimiento, nombres map[int]string) RespuestaLibrosAdmin {
	libros := map[int]*RespuestaLibroUsuario{} // id del usuario → su libro
	respuesta := RespuestaLibrosAdmin{Libros: []RespuestaLibroUsuario{}}

	// La lista viene del más nuevo al más viejo, por eso se recorre al revés
	for i := len(lista) - 1; i >= 0; i-- {
		movimiento := lista[i]

		// Si el usuario todavía no tiene libro, se le crea uno vacío
		libro, ok := libros[movimiento.IdUsuario]
		if !ok {
			libro = &RespuestaLibroUsuario{
				IdUsuario:     movimiento.IdUsuario,
				NombreUsuario: nombres[movimiento.IdUsuario],
				Filas:         []RespuestaFilaLibro{},
			}
			libros[movimiento.IdUsuario] = libro
		}

		// Agregar la fila: el valor va en "ingreso" o en "egreso" según el tipo
		fila := RespuestaFilaLibro{Id: movimiento.Id, Fecha: movimiento.Fecha, Concepto: movimiento.Concepto, Categoria: movimiento.Categoria}
		if movimiento.Tipo == "ingreso" {
			fila.Ingreso = movimiento.Valor
			libro.TotalIngresos += movimiento.Valor
		} else {
			fila.Egreso = movimiento.Valor
			libro.TotalEgresos += movimiento.Valor
		}
		libro.Filas = append(libro.Filas, fila)
	}

	// Calcular el resultado de cada libro y sumar los totales de la plataforma
	for _, libro := range libros {
		libro.Resultado = libro.TotalIngresos - libro.TotalEgresos
		respuesta.TotalIngresosPlataforma += libro.TotalIngresos
		respuesta.TotalEgresosPlataforma += libro.TotalEgresos
		respuesta.Libros = append(respuesta.Libros, *libro)
	}

	// Ordenar los libros por id de usuario
	sort.Slice(respuesta.Libros, func(i, j int) bool {
		return respuesta.Libros[i].IdUsuario < respuesta.Libros[j].IdUsuario
	})
	respuesta.UsuariosActivos = len(respuesta.Libros)
	return respuesta
}

// --- Metas ---

// Las metas viejas de la base guardan el ícono como una palabra en inglés.
// Este diccionario las traduce al emoji que muestra el frontend.
var iconosMeta = map[string]string{"shield": "🛟", "plane": "✈️", "home": "🏠", "car": "🚙"}

// iconoMeta devuelve el emoji de una meta: traduce las palabras viejas y
// usa 🎯 si la meta no tiene ícono.
func iconoMeta(icono string) string {
	if emoji, ok := iconosMeta[icono]; ok {
		return emoji
	}
	if icono == "" {
		return "🎯"
	}
	return icono
}

// armarMeta pasa una meta de la tabla al formato del frontend y calcula
// su porcentaje (máximo 100), lo que le falta y si está cumplida.
func armarMeta(meta Meta) RespuestaMeta {
	avance := 0
	if meta.MontoObjetivo > 0 {
		// Ej. 500.000 de 2.000.000 = 25 %. Se limita a 100 aunque se haya pasado.
		avance = int(math.Min(100, math.Round(meta.MontoActual/meta.MontoObjetivo*100)))
	}

	return RespuestaMeta{
		Id:         meta.Id,
		IdUsuario:  meta.IdUsuario,
		Nombre:     meta.Nombre,
		Icono:      iconoMeta(meta.Icono),
		Actual:     meta.MontoActual,
		Objetivo:   meta.MontoObjetivo,
		Porcentaje: avance,
		Faltante:   math.Max(0, meta.MontoObjetivo-meta.MontoActual), // nunca negativo
		Cumplida:   meta.MontoActual >= meta.MontoObjetivo,
	}
}

// armarMetas pasa todas las metas al formato del frontend y calcula los
// totales de la pantalla de Metas.
func armarMetas(metas []Meta) RespuestaMetas {
	respuesta := RespuestaMetas{Metas: []RespuestaMeta{}}

	for _, meta := range metas {
		item := armarMeta(meta)
		respuesta.Metas = append(respuesta.Metas, item)
		respuesta.TotalAhorrado += item.Actual
		respuesta.TotalObjetivo += item.Objetivo
		if item.Cumplida {
			respuesta.MetasAlcanzadas++ // suma 1
		}
	}

	// Ordenar por id para que siempre salgan en el mismo orden
	sort.Slice(respuesta.Metas, func(i, j int) bool {
		return respuesta.Metas[i].Id < respuesta.Metas[j].Id
	})

	respuesta.TotalMetas = len(respuesta.Metas)
	respuesta.ProgresoGeneral = porcentaje(respuesta.TotalAhorrado, respuesta.TotalObjetivo)
	return respuesta
}

// --- Inversiones ---

// mapaNiveles convierte la lista de niveles en un diccionario id → nombre,
// por ejemplo {1: "Bajo", 2: "Medio", 3: "Alto"}.
func mapaNiveles(niveles []NivelRiesgo) map[int]string {
	mapa := map[int]string{}
	for _, nivel := range niveles {
		mapa[nivel.Id] = nivel.Nombre
	}
	return mapa
}

// calcularRentabilidad pasa el rendimiento en pesos a porcentaje del monto,
// con dos decimales. Ejemplo: 90.000 de 1.000.000 = 9 (%).
func calcularRentabilidad(rendimiento, monto float64) float64 {
	if monto == 0 {
		return 0
	}
	return math.Round(rendimiento/monto*10000) / 100
}

// rendimientoDe devuelve la ganancia en pesos de una inversión.
// Si la inversión es vieja y no tiene ese dato, lo calcula con su rentabilidad (%).
func rendimientoDe(inversion Inversion) float64 {
	if inversion.Rendimiento == 0 && inversion.Rentabilidad != 0 {
		return math.Round(inversion.Monto * inversion.Rentabilidad / 100)
	}
	return inversion.Rendimiento
}

// armarInversion pasa una inversión de la tabla al formato del frontend,
// con el nombre de su nivel de riesgo y su rendimiento calculado.
func armarInversion(inversion Inversion, niveles map[int]string) RespuestaInversion {
	rendimiento := rendimientoDe(inversion)

	// Buscar el nombre del nivel de riesgo ("Bajo", "Medio" o "Alto")
	riesgo := ""
	if inversion.IdNivelRiesgo != nil {
		riesgo = niveles[inversion.IdNivelRiesgo.Id]
	}

	duracion := inversion.Duracion
	if duracion == "" {
		duracion = "Sin definir"
	}

	return RespuestaInversion{
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

// armarInversiones calcula todo lo de la pantalla de Inversiones:
// la lista, los totales, el reparto por riesgo, la cartera, la alerta y la gráfica.
func armarInversiones(inversiones []Inversion, niveles []NivelRiesgo, historial []HistorialInversion) RespuestaInversiones {
	mapa := mapaNiveles(niveles)

	// Empezar con las listas vacías y los tres riesgos en 0
	respuesta := RespuestaInversiones{
		Inversiones:    []RespuestaInversion{},
		TotalPorRiesgo: map[string]float64{"Bajo": 0, "Medio": 0, "Alto": 0},
		Carteras:       []RespuestaParte{},
	}
	ids := map[int]bool{} // ids de las inversiones incluidas (para la gráfica)

	// Sumar cada inversión a los totales
	for _, inversion := range inversiones {
		item := armarInversion(inversion, mapa)
		respuesta.Inversiones = append(respuesta.Inversiones, item)
		respuesta.BalanceInversiones += item.Monto
		respuesta.RendimientoTotal += item.Rendimiento
		if item.Riesgo != "" {
			respuesta.TotalPorRiesgo[item.Riesgo] += item.Monto
		}
		ids[inversion.Id] = true
	}

	// Ordenar por id
	sort.Slice(respuesta.Inversiones, func(i, j int) bool {
		return respuesta.Inversiones[i].Id < respuesta.Inversiones[j].Id
	})

	respuesta.RendimientoPorcentaje = porcentajeDecimal(respuesta.RendimientoTotal, respuesta.BalanceInversiones)

	// Cartera: qué parte del total invertido es cada inversión
	for _, item := range respuesta.Inversiones {
		respuesta.Carteras = append(respuesta.Carteras, RespuestaParte{
			Etiqueta:   item.Nombre,
			Monto:      item.Monto,
			Porcentaje: porcentaje(item.Monto, respuesta.BalanceInversiones),
		})
	}

	// Alerta si la mitad o más del dinero está en riesgo alto
	enRiesgoAlto := porcentaje(respuesta.TotalPorRiesgo["Alto"], respuesta.BalanceInversiones)
	if respuesta.BalanceInversiones > 0 && enRiesgoAlto >= 50 {
		respuesta.AlertaConcentracion = RespuestaAlerta{
			Mostrar: true,
			Mensaje: fmt.Sprintf("El %d%% de tu dinero invertido está en riesgo alto. Considera diversificar en opciones de menor riesgo.", enRiesgoAlto),
		}
	}

	// Gráfica de crecimiento con el historial de estas inversiones
	respuesta.Crecimiento = calcularCrecimiento(historial, ids)
	return respuesta
}

// calcularCrecimiento suma, mes a mes, el valor de las inversiones indicadas
// en "ids" y devuelve los puntos de la gráfica ordenados por fecha.
func calcularCrecimiento(historial []HistorialInversion, ids map[int]bool) []RespuestaPunto {
	porPeriodo := map[string]float64{} // "2026-05" → suma del valor ese mes
	for _, punto := range historial {
		// Solo se cuentan los puntos de las inversiones que se están mostrando
		if punto.IdInversion != nil && ids[punto.IdInversion.Id] {
			porPeriodo[punto.Fecha.Format("2006-01")] += punto.Valor
		}
	}

	// Pasar el diccionario a lista
	puntos := []RespuestaPunto{}
	for periodo, valor := range porPeriodo {
		puntos = append(puntos, RespuestaPunto{Periodo: periodo, Valor: valor})
	}

	// Ordenar del mes más viejo al más nuevo
	sort.Slice(puntos, func(i, j int) bool {
		return puntos[i].Periodo < puntos[j].Periodo
	})
	return puntos
}

// ============================================================================
// 8. AYUDAS HTTP (funciones del controlador para leer y responder)
// ============================================================================

// responder envía una respuesta al frontend con el código HTTP indicado
// (200 = bien, 201 = creado...) y los datos convertidos a JSON.
func (c *FinanzasController) responder(estado int, datos interface{}) {
	c.Ctx.Output.SetStatus(estado)
	c.Data["json"] = datos
	c.ServeJSON()
}

// responderError envía un error al frontend, por ejemplo:
//
//	código 404 con {"error": "El usuario no existe"}
//
// Si el error es un errorFinanzas usa su código; si no, usa 500 (error interno).
func (c *FinanzasController) responderError(err error) {
	estado := http.StatusInternalServerError
	if e, ok := err.(errorFinanzas); ok {
		estado = e.estado
	}
	c.responder(estado, map[string]interface{}{"error": err.Error()})
}

// idDeLaURL lee el ":id" de la ruta y lo convierte a número.
// Ejemplo: en /v1/metas/6 devuelve 6. Si no es un número mayor que 0, error 400.
func (c *FinanzasController) idDeLaURL() (int, error) {
	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || id <= 0 {
		return 0, errorFinanzas{http.StatusBadRequest, "El ID debe ser un numero mayor que cero"}
	}
	return id, nil
}

// leerDatos convierte el JSON que envió el frontend a la plantilla "destino".
// Si el JSON está mal escrito, error 400.
func (c *FinanzasController) leerDatos(destino interface{}) error {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, destino); err != nil {
		return errorFinanzas{http.StatusBadRequest, "El cuerpo de la peticion no es un JSON valido"}
	}
	return nil
}

// filtrosDeLaURL lee los filtros del Libro mayor que vienen en la URL
// (?tipo=gasto&categoria=Salud&desde=...&hasta=...&q=...) y revisa que estén bien.
func (c *FinanzasController) filtrosDeLaURL() (FiltrosMovimientos, error) {
	// c.GetString("tipo") devuelve el valor de ?tipo=... (o "" si no viene)
	filtros := FiltrosMovimientos{
		Tipo:      strings.ToLower(strings.TrimSpace(c.GetString("tipo"))),
		Categoria: c.GetString("categoria"),
		Desde:     strings.TrimSpace(c.GetString("desde")),
		Hasta:     strings.TrimSpace(c.GetString("hasta")),
		Texto:     c.GetString("q"),
	}

	// El tipo solo puede ser vacío, "todos", "ingreso" o "gasto"
	if filtros.Tipo != "" && filtros.Tipo != "todos" && filtros.Tipo != "ingreso" && filtros.Tipo != "gasto" {
		return filtros, errorFinanzas{http.StatusBadRequest, "El filtro tipo debe ser todos, ingreso o gasto"}
	}

	// Las fechas, si vienen, deben tener el formato AAAA-MM-DD
	for _, fecha := range []string{filtros.Desde, filtros.Hasta} {
		if fecha == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", fecha); err != nil {
			return filtros, errorFinanzas{http.StatusBadRequest, "Las fechas de los filtros deben tener el formato AAAA-MM-DD"}
		}
	}
	return filtros, nil
}
