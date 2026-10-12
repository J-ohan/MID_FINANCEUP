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

