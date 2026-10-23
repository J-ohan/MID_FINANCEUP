// @APIVersion 1.0.0
// @Title FinanceUp - API MID
// @Description Une la informacion de los 5 CRUD de FinanceUp
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"api_mid_financeup/controllers"
	"api_mid_financeup/controllers/finanzas"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// General (para probar que los CRUD esten encendidos)
	beego.Router("/v1/estado", &controllers.GeneralController{}, "get:Estado")
	beego.Router("/v1/dashboard/:id", &controllers.GeneralController{}, "get:GetDashboard")

	// Usuarios (crud_auth)
	beego.Router("/v1/registro", &controllers.AuthController{}, "post:PostRegistro")
	beego.Router("/v1/login", &controllers.AuthController{}, "post:PostLogin")
	beego.Router("/v1/login/verificar", &controllers.AuthController{}, "post:PostVerificarCodigo")
	beego.Router("/v1/login/reenviar", &controllers.AuthController{}, "post:PostReenviarCodigo")
	beego.Router("/v1/logout", &controllers.AuthController{}, "post:PostLogout")
	beego.Router("/v1/login/social", &controllers.AuthController{}, "post:PostLoginSocial")
	beego.Router("/v1/recuperar-contrasena", &controllers.AuthController{}, "post:PostRecuperarContrasena;put:PutRestablecerContrasena")
	beego.Router("/v1/perfil/:id", &controllers.AuthController{}, "get:GetPerfil")
	beego.Router("/v1/perfil/:id/contrasena", &controllers.AuthController{}, "put:PutCambiarContrasena")
	beego.Router("/v1/admin/usuarios", &controllers.AuthController{}, "get:GetAdminUsuarios")
	beego.Router("/v1/admin/usuarios/:id/rol", &controllers.AuthController{}, "put:PutRolUsuario")
	beego.Router("/v1/admin/usuarios/:id/estado", &controllers.AuthController{}, "put:PutEstadoUsuario")

	// Educacion (crud_educacion)
	beego.Router("/v1/cursos", &controllers.EducacionController{}, "get:GetCursos")
	beego.Router("/v1/cursos/:id", &controllers.EducacionController{}, "get:GetCurso")
	beego.Router("/v1/cursos/:id/reiniciar", &controllers.EducacionController{}, "post:PostReiniciarCurso")
	beego.Router("/v1/clase/:id", &controllers.EducacionController{}, "get:GetClase")
	beego.Router("/v1/clase/:id/completar", &controllers.EducacionController{}, "post:PostCompletarLeccion")
	beego.Router("/v1/progreso-educativo/:id", &controllers.EducacionController{}, "get:GetProgresoEducativo")
	beego.Router("/v1/certificado", &controllers.EducacionController{}, "post:PostCertificado")

	// Alianzas (crud_negocio)
	beego.Router("/v1/ofertas", &controllers.NegocioController{}, "get:GetOfertas")
	beego.Router("/v1/ofertas/:id", &controllers.NegocioController{}, "get:GetOferta")
	beego.Router("/v1/solicitud-credito", &controllers.NegocioController{}, "post:PostSolicitudCredito")
	beego.Router("/v1/solicitud-credito/usuario/:id", &controllers.NegocioController{}, "get:GetSolicitudesUsuario")

	// PQR (crud_soporte)
	beego.Router("/v1/pqr", &controllers.SoporteController{}, "post:PostPqr")
	beego.Router("/v1/pqr/usuario/:id", &controllers.SoporteController{}, "get:GetPqrUsuario")
	beego.Router("/v1/pqr/radicado/:radicado", &controllers.SoporteController{}, "get:GetPqr")
	beego.Router("/v1/admin/pqr/:id/respuesta", &controllers.SoporteController{}, "post:PostRespuestaPqrAdmin")
	beego.Router("/v1/admin/pqr", &controllers.SoporteController{}, "get:GetPqrAdmin")

	// Finanzas (crud_finanzas) - controladores en controllers/finanzas/
	// Crear, editar o borrar movimientos, metas e inversiones, y la lista
	// de categorias, la pagina los hace directo en crud_finanzas (8082).
	beego.Router("/v1/resumen-financiero/:id", &finanzas.FinanzasController{}, "get:GetResumenFinanciero")
	beego.Router("/v1/movimientos/usuario/:id", &finanzas.FinanzasController{}, "get:GetMovimientos")
	beego.Router("/v1/movimientos/usuario/:id/exportar", &finanzas.FinanzasController{}, "get:GetExportarMovimientos")
	beego.Router("/v1/movimientos/importar", &finanzas.FinanzasController{}, "post:PostImportarMovimientos")
	beego.Router("/v1/metas/usuario/:id", &finanzas.FinanzasController{}, "get:GetMetas")
	beego.Router("/v1/metas/:id/abono", &finanzas.FinanzasController{}, "put:PutAbonoMeta")
	beego.Router("/v1/inversiones/usuario/:id", &finanzas.FinanzasController{}, "get:GetInversiones")
	beego.Router("/v1/inversiones/usuario/:id/exportar", &finanzas.FinanzasController{}, "get:GetExportarInversiones")
	beego.Router("/v1/admin/finanzas", &finanzas.FinanzasController{}, "get:GetAdminFinanzas")
	beego.Router("/v1/admin/libro-mayor", &finanzas.FinanzasController{}, "get:GetAdminLibroMayor")
}
