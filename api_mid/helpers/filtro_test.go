package helpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

// Pruebas del filtro de autenticacion. Se corren con: go test ./helpers

// pasarPorFiltro simula una peticion y dice que codigo respondio el filtro
// (0 = la dejo pasar).
func pasarPorFiltro(metodo string, ruta string, token string, body string) int {

	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(body))
	if token != "" {
		peticion.Header.Set("Authorization", "Bearer "+token)
	}

	respuesta := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(respuesta, peticion)
	ctx.Input.RequestBody = []byte(body)

	FiltroAutenticacion(ctx)

	if !ctx.ResponseWriter.Started {
		return 0
	}

	return respuesta.Code
}

func TestFiltroPermisos(t *testing.T) {

	usuario1, _ := CrearToken(1, "user")
	admin, _ := CrearToken(9, "admin")
	verificacion, _ := CrearTokenVerificacion(1)

	casos := []struct {
		nombre              string
		metodo, ruta, token string
		body                string
		esperado            int
	}{
		// Rutas publicas: pasan sin token
		{"login sin token", "POST", "/v1/login", "", `{"email":"a@b.co"}`, 0},
		{"registro sin token", "POST", "/v1/registro", "", ``, 0},
		{"recuperar (PUT) sin token", "PUT", "/v1/recuperar-contrasena", "", ``, 0},
		{"estado sin token", "GET", "/v1/estado", "", ``, 0},
		{"catalogo sin token", "GET", "/v1/cursos/3", "", ``, 0},
		{"preflight del navegador", "OPTIONS", "/v1/perfil/1", "", ``, 0},

		// Sin token o con token falso: 401
		{"perfil sin token", "GET", "/v1/perfil/1", "", ``, http.StatusUnauthorized},
		{"token inventado", "GET", "/v1/perfil/1", "abc.def.ghi", ``, http.StatusUnauthorized},
		{"catalogo con progreso de un usuario", "GET", "/v1/cursos/3?usuario=2", "", ``, http.StatusUnauthorized},

		// Con token: solo sus datos
		{"su perfil", "GET", "/v1/perfil/1", usuario1, ``, 0},
		{"perfil de otro", "GET", "/v1/perfil/2", usuario1, ``, http.StatusForbidden},
		{"clave de otro", "PUT", "/v1/perfil/2/contrasena", usuario1, ``, http.StatusForbidden},
		{"sus finanzas", "GET", "/v1/resumen-financiero/1", usuario1, ``, 0},
		{"finanzas de otro", "GET", "/v1/movimientos/usuario/2", usuario1, ``, http.StatusForbidden},
		{"importar a nombre de otro", "POST", "/v1/movimientos/importar", usuario1, `{"id_usuario":2}`, http.StatusForbidden},
		{"importar a su nombre", "POST", "/v1/movimientos/importar", usuario1, `{"id_usuario":1}`, 0},
		{"progreso de otro por ?usuario=", "GET", "/v1/cursos/3?usuario=2", usuario1, ``, http.StatusForbidden},
		{"una meta (el :id no es de usuario)", "PUT", "/v1/metas/55/abono", usuario1, `{"valor":5}`, 0},
		{"admin siendo usuario", "GET", "/v1/admin/finanzas", usuario1, ``, http.StatusForbidden},

		// Intentos de saltarse el filtro
		{"id_usuario en mayusculas", "POST", "/v1/solicitud-credito", usuario1, `{"ID_USUARIO":2}`, http.StatusForbidden},
		{"id_usuario repetido", "POST", "/v1/solicitud-credito", usuario1, `{"id_usuario":1,"Id_Usuario":2}`, http.StatusForbidden},
		{"id_usuario como texto", "POST", "/v1/solicitud-credito", usuario1, `{"id_usuario":"2"}`, http.StatusForbidden},
		{"id con signo en la URL", "GET", "/v1/perfil/+2", usuario1, ``, http.StatusForbidden},
		{"id con ceros a la izquierda", "GET", "/v1/perfil/01", usuario1, ``, 0},
		{"token del paso 2 del login", "GET", "/v1/perfil/1", verificacion, ``, http.StatusUnauthorized},

		// El admin puede ver todo
		{"admin ve perfil de otro", "GET", "/v1/perfil/2", admin, ``, 0},
		{"admin ve vistas de admin", "GET", "/v1/admin/finanzas", admin, ``, 0},
	}

	for _, caso := range casos {
		if obtenido := pasarPorFiltro(caso.metodo, caso.ruta, caso.token, caso.body); obtenido != caso.esperado {
			t.Errorf("%s: esperaba %d, llego %d", caso.nombre, caso.esperado, obtenido)
		}
	}
}
