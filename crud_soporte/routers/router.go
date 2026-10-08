// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	":/Users/Alejandro Zorro/Desktop/Finance_Up/MID_FINANCEUP/crud_soporte/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/registro_actividad",
			beego.NSInclude(
				&controllers.RegistroActividadController{},
			),
		),

		beego.NSNamespace("/estado_pqr",
			beego.NSInclude(
				&controllers.EstadoPqrController{},
			),
		),

		beego.NSNamespace("/pqr",
			beego.NSInclude(
				&controllers.PqrController{},
			),
		),

		beego.NSNamespace("/adjunto",
			beego.NSInclude(
				&controllers.AdjuntoController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
