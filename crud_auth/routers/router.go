// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	":/Users/Alejandro Zorro/Desktop/Finance_Up/MID_FINANCEUP/crud_auth/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/tipo_documento",
			beego.NSInclude(
				&controllers.TipoDocumentoController{},
			),
		),

		beego.NSNamespace("/usuario",
			beego.NSInclude(
				&controllers.UsuarioController{},
			),
		),

		beego.NSNamespace("/credencial",
			beego.NSInclude(
				&controllers.CredencialController{},
			),
		),

		beego.NSNamespace("/usuario_rol",
			beego.NSInclude(
				&controllers.UsuarioRolController{},
			),
		),

		beego.NSNamespace("/auditoria_login",
			beego.NSInclude(
				&controllers.AuditoriaLoginController{},
			),
		),

		beego.NSNamespace("/rol",
			beego.NSInclude(
				&controllers.RolController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
