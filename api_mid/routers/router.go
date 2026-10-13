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
 
	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")
 
	// Creditos (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")
	beego.Router("/v1/solicitud-credito", &controllers.NegocioController{}, "post:PostSolicitudCredito")
	beego.Router("/v1/solicitud-credito/usuario/:id", &controllers.NegocioController{}, "get:GetSolicitudesUsuario")
 }
