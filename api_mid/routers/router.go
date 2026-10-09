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
	// Estado de los CRUD
	beego.Router("/v1/estado", &controllers.MidController{}, "get:Estado")

	// Usuarios (crud_auth)
	beego.Router("/v1/perfil/:id", &controllers.MidController{}, "get:GetPerfil")
	beego.Router("/v1/registro", &controllers.MidController{}, "post:PostRegistro")
	beego.Router("/v1/login", &controllers.MidController{}, "post:PostLogin")

	// Finanzas (crud_finanzas)
	beego.Router("/v1/resumen-financiero/:id", &controllers.MidController{}, "get:GetResumenFinanciero")

	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.MidController{}, "get:GetProgresoEducativo")

	// Creditos (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.MidController{}, "get:GetOfertas")
	beego.Router("/v1/solicitud-credito", &controllers.MidController{}, "post:PostSolicitudCredito")
	beego.Router("/v1/solicitud-credito/usuario/:id", &controllers.MidController{}, "get:GetSolicitudesUsuario")

	// PQR (crud_soporte)
	beego.Router("/v1/pqr", &controllers.MidController{}, "post:PostPqr")
	beego.Router("/v1/pqr/usuario/:id", &controllers.MidController{}, "get:GetPqrUsuario")

	// Todo junto
	beego.Router("/v1/dashboard/:id", &controllers.MidController{}, "get:GetDashboard")

	// Pasarela a cualquier CRUD
	beego.Router("/v1/crud/:conjunto/*", &controllers.MidController{}, "*:Reenviar")
}
