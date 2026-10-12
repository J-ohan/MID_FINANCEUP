package routers

// ============================================================================
// ROUTER.GO — La "tabla de direcciones" del MID
// ============================================================================
//
// Aquí se dice qué función atiende cada URL. Cada línea tiene esta forma:
//
//	beego.Router( RUTA , CONTROLADOR , "método_http:Función" )
//
// Ejemplo:
//
//	beego.Router("/v1/metas", &controllers.FinanzasController{}, "post:PostMeta")
//
// se lee así: "cuando llegue un POST a /v1/metas, ejecutar la función
// PostMeta del controlador de finanzas".
//
//   - :id dentro de la ruta es un valor que cambia, por ejemplo /v1/metas/6.
//     En el controlador se lee con c.Ctx.Input.Param(":id").
//   - Si una misma ruta atiende varios métodos, se separan con punto y coma:
//     "put:PutMeta;delete:DeleteMeta" → PUT edita y DELETE elimina.
//   - Métodos HTTP: get = consultar, post = crear, put = editar, delete = eliminar.
//
// En esta rama solo está el módulo de Finanzas (controllers/Finanzas.go).
// ============================================================================

import (
	"api_mid_financeup/controllers" // Donde están los controladores

	beego "github.com/beego/beego/v2/server/web" // El framework web
)

// init se ejecuta solo, una vez, cuando arranca el MID: registra todas las rutas.
func init() {

	// Panel principal: tarjetas, salud, dona, metas e inversiones de un usuario
	beego.Router("/v1/resumen-financiero/:id", &controllers.FinanzasController{}, "get:GetResumenFinanciero")
	// Categorías para los selects y filtros
	beego.Router("/v1/categorias", &controllers.FinanzasController{}, "get:GetCategorias")

	// Libro mayor
	// Movimientos de un usuario (acepta filtros: ?tipo=&categoria=&desde=&hasta=&q=)
	beego.Router("/v1/movimientos/usuario/:id", &controllers.FinanzasController{}, "get:GetMovimientosUsuario")
	// Crear un movimiento
	beego.Router("/v1/movimientos", &controllers.FinanzasController{}, "post:PostMovimiento")
	// Importar movimientos desde la plantilla de Excel
	beego.Router("/v1/movimientos/importar", &controllers.FinanzasController{}, "post:PostImportarMovimientos")
	// Editar (PUT) o eliminar (DELETE) un movimiento
	beego.Router("/v1/movimientos/:id", &controllers.FinanzasController{}, "put:PutMovimiento;delete:DeleteMovimiento")

	// Metas
	// Metas de un usuario con sus totales
	beego.Router("/v1/metas/usuario/:id", &controllers.FinanzasController{}, "get:GetMetasUsuario")
	// Crear una meta
	beego.Router("/v1/metas", &controllers.FinanzasController{}, "post:PostMeta")
	// Editar (PUT) o eliminar (DELETE) una meta
	beego.Router("/v1/metas/:id", &controllers.FinanzasController{}, "put:PutMeta;delete:DeleteMeta")
	// Abonar dinero a una meta
	beego.Router("/v1/metas/:id/abono", &controllers.FinanzasController{}, "put:PutAbonoMeta")

	// Inversiones
	// Inversiones de un usuario con su resumen y gráfica
	beego.Router("/v1/inversiones/usuario/:id", &controllers.FinanzasController{}, "get:GetInversionesUsuario")
	// Crear una inversión
	beego.Router("/v1/inversiones", &controllers.FinanzasController{}, "post:PostInversion")
	// Editar (PUT) o eliminar (DELETE) una inversión
	beego.Router("/v1/inversiones/:id", &controllers.FinanzasController{}, "put:PutInversion;delete:DeleteInversion")

	// Admin (datos de todos los usuarios)
	beego.Router("/v1/admin/movimientos", &controllers.FinanzasController{}, "get:GetAdminMovimientos")
	beego.Router("/v1/admin/metas", &controllers.FinanzasController{}, "get:GetAdminMetas")
	beego.Router("/v1/admin/inversiones", &controllers.FinanzasController{}, "get:GetAdminInversiones")
	// Libro mayor agrupado por usuario, con el nombre de cada uno
	beego.Router("/v1/admin/libro-mayor", &controllers.FinanzasController{}, "get:GetAdminLibroMayor")
}
