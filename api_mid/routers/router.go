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
	beego.Router("/v1/resumen-financiero/:id", &controllers.FinanzasController{}, "get:GetResumenFinanciero")
	beego.Router("/v1/movimientos/usuario/:id", &controllers.FinanzasController{}, "get:GetMovimientos")
	beego.Router("/v1/movimientos/importar", &controllers.FinanzasController{}, "post:PostImportarMovimientos")
	beego.Router("/v1/metas/usuario/:id", &controllers.FinanzasController{}, "get:GetMetas")
	// Pendiente: abono de metas, inversiones y vista de admin
	// (los services ya estan en services/finanzas.go)
}
