package routers

import (
	"api_mid_financeup/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	// Educacion (crud_educacion)
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")

	// Alianzas (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")
}
