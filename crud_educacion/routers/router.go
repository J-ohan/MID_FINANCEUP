// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	":/Users/Alejandro Zorro/Desktop/Finance_Up/MID_FINANCEUP/crud_educacion/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/progreso_educativo",
			beego.NSInclude(
				&controllers.ProgresoEducativoController{},
			),
		),

		beego.NSNamespace("/progreso_leccion",
			beego.NSInclude(
				&controllers.ProgresoLeccionController{},
			),
		),

		beego.NSNamespace("/modulo_educativo",
			beego.NSInclude(
				&controllers.ModuloEducativoController{},
			),
		),

		beego.NSNamespace("/contenido",
			beego.NSInclude(
				&controllers.ContenidoController{},
			),
		),

		beego.NSNamespace("/leccion",
			beego.NSInclude(
				&controllers.LeccionController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
