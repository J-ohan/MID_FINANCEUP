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

// Rutas que no necesitan token ("METODO /ruta").
var rutasPublicas = map[string]bool{
	"GET /v1/estado":                true,
	"POST /v1/registro":             true,
	"POST /v1/login":                true,
	"POST /v1/login/verificar":      true,
	"POST /v1/login/reenviar":       true,
	"POST /v1/login/social":         true,
	"POST /v1/recuperar-contrasena": true,
	"PUT /v1/recuperar-contrasena":  true,
	"POST /v1/verificar-correo":     true,
	"PUT /v1/verificar-correo":      true,
}

// Catalogos publicos (solo lectura). Con ?usuario= si piden token.
var catalogoPublico = regexp.MustCompile(`^/v1/(cursos|ofertas)(/\d+)?$`)

// Rutas donde el :id es el id de un USUARIO.
var rutasConIdUsuario = []*regexp.Regexp{
	regexp.MustCompile(`^/v1/(perfil|dashboard|resumen-financiero|progreso-educativo)/(\d+)(/|$)`),
	regexp.MustCompile(`^/v1/(movimientos|metas|inversiones|solicitud-credito|pqr)/usuario/(\d+)(/|$)`),
}

// FiltroAutenticacion revisa el token y los permisos antes de cada ruta.
func FiltroAutenticacion(ctx *context.Context) {

	metodo := ctx.Input.Method()
	ruta := strings.TrimRight(ctx.Input.URL(), "/")

	// OPTIONS lo manda el navegador (CORS) y no trae token
	if metodo == http.MethodOptions || rutasPublicas[metodo+" "+ruta] {
		return
	}

	if metodo == http.MethodGet && catalogoPublico.MatchString(ruta) && ctx.Input.Query("usuario") == "" {
		return
	}

	datos, err := LeerToken(strings.TrimPrefix(ctx.Input.Header("Authorization"), "Bearer "))
	if err != nil {
		responderFiltro(ctx, err)
		return
	}

	// Para que los controladores sepan quien es
	ctx.Input.SetData("token", datos)

	if strings.HasPrefix(ruta, "/v1/admin/") && datos.Rol != "admin" {
		responderFiltro(ctx, NuevoError(http.StatusForbidden, "Solo un administrador puede ver esta informacion"))
		return
	}

	if datos.Rol == "admin" {
		return
	}

	for _, id := range idsPedidos(ctx, ruta) {
		if id != datos.IdUsuario {
			responderFiltro(ctx, NuevoError(http.StatusForbidden, "No puedes ver ni cambiar datos de otro usuario"))
			return
		}
	}
}

// idsPedidos busca de quien son los datos: en la URL, en ?usuario= y en el "id_usuario" del body.
func idsPedidos(ctx *context.Context, ruta string) []int {

	ids := []int{}

	for _, patron := range rutasConIdUsuario {
		if partes := patron.FindStringSubmatch(ruta); partes != nil {
			id, _ := strconv.Atoi(partes[2])
			ids = append(ids, id)
		}
	}

	if texto := ctx.Input.Query("usuario"); texto != "" {
		id, _ := strconv.Atoi(texto)
		ids = append(ids, id)
	}

	var body map[string]interface{}
	if json.Unmarshal(ctx.Input.RequestBody, &body) == nil {
		if valor, existe := body["id_usuario"]; existe {
			id, _ := valor.(float64)
			ids = append(ids, int(id))
		}
	}

	return ids
}

// responderFiltro corta la peticion y responde {"error": "..."} con su codigo.
func responderFiltro(ctx *context.Context, err error) {

	codigo := http.StatusUnauthorized
	if errorApi, ok := err.(*ErrorApi); ok {
		codigo = errorApi.Codigo
	}

	ctx.Output.SetStatus(codigo)
	ctx.Output.JSON(map[string]interface{}{"error": err.Error()}, false, false)
}

// IdUsuarioDeLaSesion devuelve el id del usuario del token (0 en rutas publicas).
func IdUsuarioDeLaSesion(c *beego.Controller) int {
	datos, _ := TokenDeLaSesion(c)
	return datos.IdUsuario
}

// TokenDeLaSesion devuelve los datos del token de la peticion.
func TokenDeLaSesion(c *beego.Controller) (DatosToken, bool) {
	datos, ok := c.Ctx.Input.GetData("token").(DatosToken)
	return datos, ok
}
