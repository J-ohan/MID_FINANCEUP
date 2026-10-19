// @APIVersion 1.0.0
// @Title FinanceUp - API MID
// @Description Une la informacion de los 5 CRUD de FinanceUp
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"api_mid_financeup/controllers"
	"api_mid_financeup/controllers/finanzas"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {


	// General (para probar que los CRUD esten encendidos)
	beego.Router("/v1/estado", &controllers.GeneralController{}, "get:Estado")

	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")

	// Alianzas (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")

	// Finanzas (crud_finanzas) - controladores en controllers/finanzas/
	// Crear, editar o borrar movimientos, metas e inversiones, y la lista
	// de categorias, la pagina los hace directo en crud_finanzas (8082).
	beego.Router("/v1/resumen-financiero/:id", &finanzas.FinanzasController{}, "get:GetResumenFinanciero")
	beego.Router("/v1/movimientos/usuario/:id", &finanzas.FinanzasController{}, "get:GetMovimientos")
	beego.Router("/v1/movimientos/usuario/:id/exportar", &finanzas.FinanzasController{}, "get:GetExportarMovimientos")
	beego.Router("/v1/movimientos/importar", &finanzas.FinanzasController{}, "post:PostImportarMovimientos")
	beego.Router("/v1/metas/usuario/:id", &finanzas.FinanzasController{}, "get:GetMetas")
	beego.Router("/v1/metas/:id/abono", &finanzas.FinanzasController{}, "put:PutAbonoMeta")
	beego.Router("/v1/inversiones/usuario/:id", &finanzas.FinanzasController{}, "get:GetInversiones")
	beego.Router("/v1/admin/finanzas", &finanzas.FinanzasController{}, "get:GetAdminFinanzas")
	beego.Router("/v1/admin/libro-mayor", &finanzas.FinanzasController{}, "get:GetAdminLibroMayor")
}
