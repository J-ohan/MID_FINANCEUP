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
	// General
	beego.Router("/v1/estado", &controllers.GeneralController{}, "get:Estado")
	beego.Router("/v1/dashboard/:id", &controllers.GeneralController{}, "get:GetDashboard")

	// Usuarios (crud_auth)
	beego.Router("/v1/perfil/:id", &controllers.AuthController{}, "get:GetPerfil")
	beego.Router("/v1/registro", &controllers.AuthController{}, "post:PostRegistro")
	beego.Router("/v1/login", &controllers.AuthController{}, "post:PostLogin")

	// Finanzas (crud_finanzas)
	beego.Router("/v1/resumen-financiero/:id", &controllers.FinanzasController{}, "get:GetResumenFinanciero")

	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")

	// Creditos (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")
	beego.Router("/v1/solicitud-credito", &controllers.NegocioController{}, "post:PostSolicitudCredito")
	beego.Router("/v1/solicitud-credito/usuario/:id", &controllers.NegocioController{}, "get:GetSolicitudesUsuario")

	// PQR (crud_soporte)
	beego.Router("/v1/pqr", &controllers.SoporteController{}, "post:PostPqr")
	beego.Router("/v1/pqr/usuario/:id", &controllers.SoporteController{}, "get:GetPqrUsuario")
}
