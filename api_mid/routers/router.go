// @APIVersion 1.0.0
// @Title FinanceUp - API MID
// @Description Une la informacion de los 5 CRUD de FinanceUp
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"api_mid_financeup/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")

	// Alianzas (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")

	// Finanzas (crud_finanzas)
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
