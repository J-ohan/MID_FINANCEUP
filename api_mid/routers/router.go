package routers

import (
	"api_mid_financeup/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

// Todas las rutas del MID. Cada linea dice: ruta, controlador y "metodo:Funcion".
func init() {
	ns := beego.NewNamespace("/v1",
		// Estado de los servicios
		beego.NSRouter("/estado", &controllers.DashboardController{}, "get:Estado"),

		// Usuarios (crud_auth)
		beego.NSRouter("/perfil/:id", &controllers.AuthController{}, "get:ObtenerPerfil"),
		beego.NSRouter("/registro", &controllers.AuthController{}, "post:Registrar"),
		beego.NSRouter("/login", &controllers.AuthController{}, "post:IniciarSesion"),

		// Finanzas (crud_finanzas)
		beego.NSRouter("/resumen-financiero/:id", &controllers.FinanzasController{}, "get:ResumenFinanciero"),

		// Educacion (crud_educacion)
		beego.NSRouter("/progreso-educativo/:id", &controllers.EducacionController{}, "get:ProgresoEducativo"),

		// Creditos (crud_negocio)
		beego.NSRouter("/ofertas", &controllers.NegocioController{}, "get:Ofertas"),
		beego.NSRouter("/solicitud-credito", &controllers.NegocioController{}, "post:CrearSolicitud"),
		beego.NSRouter("/solicitud-credito/usuario/:id", &controllers.NegocioController{}, "get:SolicitudesDeUsuario"),

		// PQR (crud_soporte)
		beego.NSRouter("/pqr", &controllers.SoporteController{}, "post:CrearPqr"),
		beego.NSRouter("/pqr/usuario/:id", &controllers.SoporteController{}, "get:PqrDeUsuario"),

		// Panel con todo junto (los 5 CRUD)
		beego.NSRouter("/dashboard/:id", &controllers.DashboardController{}, "get:Dashboard"),

		// Pasarela: cualquier operacion directa sobre un CRUD
		beego.NSRouter("/crud/:conjunto/*", &controllers.PasarelaController{}, "*:Reenviar"),
	)
	beego.AddNamespace(ns)
}
