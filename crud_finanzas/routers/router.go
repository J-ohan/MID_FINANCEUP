// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	":/Users/Alejandro Zorro/Desktop/Finance_Up/MID_FINANCEUP/crud_finanzas/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/inversion",
			beego.NSInclude(
				&controllers.InversionController{},
			),
		),

		beego.NSNamespace("/movimiento_meta",
			beego.NSInclude(
				&controllers.MovimientoMetaController{},
			),
		),

		beego.NSNamespace("/tipo_ingreso_meta",
			beego.NSInclude(
				&controllers.TipoIngresoMetaController{},
			),
		),

		beego.NSNamespace("/meta",
			beego.NSInclude(
				&controllers.MetaController{},
			),
		),

		beego.NSNamespace("/tipo_ingreso",
			beego.NSInclude(
				&controllers.TipoIngresoController{},
			),
		),

		beego.NSNamespace("/tipo_inversion",
			beego.NSInclude(
				&controllers.TipoInversionController{},
			),
		),

		beego.NSNamespace("/nivel_riesgo",
			beego.NSInclude(
				&controllers.NivelRiesgoController{},
			),
		),

		beego.NSNamespace("/movimiento_inversion",
			beego.NSInclude(
				&controllers.MovimientoInversionController{},
			),
		),

		beego.NSNamespace("/editar_meta",
			beego.NSInclude(
				&controllers.EditarMetaController{},
			),
		),

		beego.NSNamespace("/solicitud_consolidacion",
			beego.NSInclude(
				&controllers.SolicitudConsolidacionController{},
			),
		),

		beego.NSNamespace("/categoria",
			beego.NSInclude(
				&controllers.CategoriaController{},
			),
		),

		beego.NSNamespace("/movimiento_ingreso_egreso",
			beego.NSInclude(
				&controllers.MovimientoIngresoEgresoController{},
			),
		),

		beego.NSNamespace("/finanzas",
			beego.NSInclude(
				&controllers.FinanzasController{},
			),
		),

		beego.NSNamespace("/tipo_ingreso_inversion",
			beego.NSInclude(
				&controllers.TipoIngresoInversionController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
