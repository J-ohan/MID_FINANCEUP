// @APIVersion 1.0.0
// @Title FinanceUp - API MID
// @Description Une la informacion de los 5 CRUD de FinanceUp
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers


// ROUTER.GO — La "tabla de direcciones" del MID

//
// Aquí se dice qué función atiende cada URL. Cada línea tiene esta forma:
//
//	beego.Router( RUTA , CONTROLADOR , "método_http:Función" )
//
// Ejemplo:
//
//	beego.Router("/v1/metas", &controllers.FinanzasController{}, "post:CrearMeta")
//
// se lee así: "cuando llegue un POST a /v1/metas, ejecutar la función
// CrearMeta del controlador de finanzas".
//
//   - :id dentro de la ruta es un valor que cambia, por ejemplo /v1/metas/6.
//     En el controlador se lee con c.Ctx.Input.Param(":id").
//   - Si una misma ruta atiende varios métodos, se separan con punto y coma:
//     "put:EditarMeta;delete:EliminarMeta" → PUT edita y DELETE elimina.
//   - Métodos HTTP: get = consultar, post = crear, put = editar, delete = eliminar.
// ============================================================================

import (
	"api_mid_financeup/controllers" // Donde están todos los controladores (finanzas.go, auth.go...)

	beego "github.com/beego/beego/v2/server/web" // El framework web
)

// init se ejecuta solo, una vez, cuando arranca el MID: registra todas las rutas.
func init() {
	// General
	beego.Router("/v1/estado", &controllers.GeneralController{}, "get:Estado")
	beego.Router("/v1/dashboard/:id", &controllers.GeneralController{}, "get:GetDashboard")

	// Usuarios (crud_auth)
	beego.Router("/v1/perfil/:id", &controllers.AuthController{}, "get:GetPerfil")
	beego.Router("/v1/registro", &controllers.AuthController{}, "post:PostRegistro")
	beego.Router("/v1/login", &controllers.AuthController{}, "post:PostLogin")

	// ------------------------------------------------------------------------
	// FINANZAS (todo esto está en controllers/finanzas.go)
	// ------------------------------------------------------------------------

	// Panel principal: tarjetas, salud, dona, metas e inversiones de un usuario
	beego.Router("/v1/resumen-financiero/:id", &controllers.FinanzasController{}, "get:ObtenerResumenFinanciero")
	// Categorías para los selects y filtros
	beego.Router("/v1/categorias", &controllers.FinanzasController{}, "get:ObtenerCategorias")

	// Finanzas - Libro mayor
	// Movimientos de un usuario (acepta filtros: ?tipo=&categoria=&desde=&hasta=&q=)
	beego.Router("/v1/movimientos/usuario/:id", &controllers.FinanzasController{}, "get:ObtenerMovimientos")
	// Crear un movimiento
	beego.Router("/v1/movimientos", &controllers.FinanzasController{}, "post:CrearMovimiento")
	// Importar movimientos desde la plantilla de Excel
	beego.Router("/v1/movimientos/importar", &controllers.FinanzasController{}, "post:ImportarMovimientos")
	// Editar (PUT) o eliminar (DELETE) un movimiento
	beego.Router("/v1/movimientos/:id", &controllers.FinanzasController{}, "put:EditarMovimiento;delete:EliminarMovimiento")

	// Finanzas - Metas
	// Metas de un usuario con sus totales
	beego.Router("/v1/metas/usuario/:id", &controllers.FinanzasController{}, "get:ObtenerMetas")
	// Crear una meta
	beego.Router("/v1/metas", &controllers.FinanzasController{}, "post:CrearMeta")
	// Editar (PUT) o eliminar (DELETE) una meta
	beego.Router("/v1/metas/:id", &controllers.FinanzasController{}, "put:EditarMeta;delete:EliminarMeta")
	// Abonar dinero a una meta
	beego.Router("/v1/metas/:id/abono", &controllers.FinanzasController{}, "put:AbonarMeta")

	// Finanzas - Inversiones
	// Inversiones de un usuario con su resumen y gráfica
	beego.Router("/v1/inversiones/usuario/:id", &controllers.FinanzasController{}, "get:ObtenerInversiones")
	// Crear una inversión
	beego.Router("/v1/inversiones", &controllers.FinanzasController{}, "post:CrearInversion")
	// Editar (PUT) o eliminar (DELETE) una inversión
	beego.Router("/v1/inversiones/:id", &controllers.FinanzasController{}, "put:EditarInversion;delete:EliminarInversion")

	// Finanzas - Admin (datos de todos los usuarios)
	beego.Router("/v1/admin/movimientos", &controllers.FinanzasController{}, "get:ObtenerMovimientosAdmin")
	beego.Router("/v1/admin/metas", &controllers.FinanzasController{}, "get:ObtenerMetasAdmin")
	beego.Router("/v1/admin/inversiones", &controllers.FinanzasController{}, "get:ObtenerInversionesAdmin")
	// Libro mayor agrupado por usuario, con el nombre de cada uno
	beego.Router("/v1/admin/libro-mayor", &controllers.FinanzasController{}, "get:ObtenerLibroMayorAdmin")

	// ------------------------------------------------------------------------
	// OTROS MÓDULOS (no son de finanzas)
	// ------------------------------------------------------------------------

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
