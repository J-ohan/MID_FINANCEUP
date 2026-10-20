package helpers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// ============================================================
// FILTRO DE AUTENTICACION
// Corre ANTES de cada ruta (se activa en main.go). Hace 3 cosas:
//   1. Deja pasar sin token las rutas publicas (login, registro...).
//   2. En las demas exige "Authorization: Bearer <token>" (si no, 401).
//   3. Revisa que el usuario solo pida SUS datos: si el id de la URL,
//      el ?usuario= o el "id_usuario" del body no es el suyo, responde 403.
//      El rol "admin" puede ver los datos de todos.
// ============================================================

// rutasPublicas: "METODO /ruta" que no necesitan token.
var rutasPublicas = map[string]bool{
	"GET /v1/estado":                true,
	"POST /v1/registro":             true,
	"POST /v1/login":                true,
	"POST /v1/login/verificar":      true,
	"POST /v1/login/reenviar":       true,
	"POST /v1/login/social":         true,
	"POST /v1/recuperar-contrasena": true,
	"PUT /v1/recuperar-contrasena":  true,
}

// Catalogos que se pueden ver sin iniciar sesion (solo lectura).
// Si llevan ?usuario= (progreso de un usuario), si piden token.
var catalogoPublico = regexp.MustCompile(`^/v1/(cursos|ofertas)(/\d+)?$`)

// Rutas donde el :id es el id de un USUARIO (no de una meta, curso, etc.)
var rutasConIdUsuario = []*regexp.Regexp{
	regexp.MustCompile(`^/v1/(perfil|dashboard|resumen-financiero|progreso-educativo)/([^/]+)(/|$)`),
	regexp.MustCompile(`^/v1/(movimientos|metas|inversiones|solicitud-credito|pqr)/usuario/([^/]+)(/|$)`),
}

// FiltroAutenticacion es el filtro que se registra en main.go.
func FiltroAutenticacion(ctx *context.Context) {

	metodo := ctx.Input.Method()
	ruta := strings.TrimRight(ctx.Input.URL(), "/")

	// Las peticiones OPTIONS son del navegador (CORS), no traen token
	if metodo == http.MethodOptions {
		return
	}

	// 1. Rutas publicas
	if rutasPublicas[metodo+" "+ruta] {
		return
	}

	if metodo == http.MethodGet && catalogoPublico.MatchString(ruta) && ctx.Input.Query("usuario") == "" {
		return
	}

	// 2. Leer y revisar el token
	token := strings.TrimPrefix(ctx.Input.Header("Authorization"), "Bearer ")

	datos, err := LeerToken(token)

	if err != nil {
		responderFiltro(ctx, err)
		return
	}

	// El token temporal del login no sirve para entrar a las demas rutas
	if datos.Rol == RolVerificacion {
		responderFiltro(ctx, NuevoError(http.StatusUnauthorized, "La sesion no es valida. Inicia sesion de nuevo."))
		return
	}

	// Se guarda para que los controladores sepan quien es (IdUsuarioDeLaSesion)
	ctx.Input.SetData("token", datos)

	// 3. Permisos
	if strings.HasPrefix(ruta, "/v1/admin/") && datos.Rol != "admin" {
		responderFiltro(ctx, NuevoError(http.StatusForbidden, "Solo un administrador puede ver esta informacion"))
		return
	}

	if datos.Rol == "admin" {
		return
	}

	for _, id := range idsDeUsuarioPedidos(ctx, ruta) {
		if id != datos.IdUsuario {
			responderFiltro(ctx, NuevoError(http.StatusForbidden, "No puedes ver ni cambiar datos de otro usuario"))
			return
		}
	}
}

// idsDeUsuarioPedidos busca de quien son los datos que se piden:
// en la URL (/v1/perfil/5), en ?usuario=5 y en el "id_usuario" del body.
func idsDeUsuarioPedidos(ctx *context.Context, ruta string) []int {

	ids := []int{}

	for _, patron := range rutasConIdUsuario {
		if partes := patron.FindStringSubmatch(ruta); partes != nil {
			ids = append(ids, idDeTexto(partes[2]))
		}
	}

	if texto := ctx.Input.Query("usuario"); texto != "" {
		ids = append(ids, idDeTexto(texto))
	}

	var body map[string]interface{}

	// Go acepta "ID_USUARIO" o "Id_Usuario" al leer el JSON, asi que se
	// revisan todas las formas de escribir la llave
	if json.Unmarshal(ctx.Input.RequestBody, &body) == nil {
		for llave, valor := range body {
			if strings.EqualFold(llave, "id_usuario") {
				id, esNumero := valor.(float64)
				if !esNumero || id != float64(int(id)) {
					id = -1
				}
				ids = append(ids, int(id))
			}
		}
	}

	return ids
}

// idDeTexto convierte el id; si no son solo numeros devuelve -1 (no es de nadie).
func idDeTexto(texto string) int {

	id, err := strconv.Atoi(texto)

	if err != nil || !SoloNumeros(texto) {
		return -1
	}

	return id
}

func responderFiltro(ctx *context.Context, err error) {

	codigo := http.StatusUnauthorized

	if errorApi, ok := err.(*ErrorApi); ok {
		codigo = errorApi.Codigo
	}

	ctx.Output.SetStatus(codigo)
	ctx.Output.JSON(map[string]interface{}{"error": err.Error()}, false, false)
}

// IdUsuarioDeLaSesion devuelve el id del usuario que inicio sesion
// (lo dejo el filtro). Devuelve 0 si la ruta es publica.
func IdUsuarioDeLaSesion(c *beego.Controller) int {

	if datos, ok := c.Ctx.Input.GetData("token").(DatosToken); ok {
		return datos.IdUsuario
	}

	return 0
}

// EsAdmin dice si el usuario que inicio sesion es administrador.
func EsAdmin(c *beego.Controller) bool {
	datos, ok := c.Ctx.Input.GetData("token").(DatosToken)
	return ok && datos.Rol == "admin"
}

// DuenoQuePide devuelve el id del usuario de la sesion para revisar que
// los datos sean suyos, o 0 si es admin (puede ver los de todos).
func DuenoQuePide(c *beego.Controller) int {

	if EsAdmin(c) {
		return 0
	}

	return IdUsuarioDeLaSesion(c)
}

// TokenDeLaSesion devuelve los datos del token de la peticion (para el logout).
func TokenDeLaSesion(c *beego.Controller) (DatosToken, bool) {
	datos, ok := c.Ctx.Input.GetData("token").(DatosToken)
	return datos, ok
}
