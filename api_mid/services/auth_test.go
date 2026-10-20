package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// PRUEBAS DE AUTH (login, registro, Google, codigos, contrasenas)
// Se corren con: go test ./services -run Auth
//
// No necesitan nada encendido:
//   - crudAuthDePrueba hace de cuenta que es crud_auth (en memoria)
//   - googleDePrueba hace de cuenta que es Google
//   - los correos no se envian: se guardan en "correos" para revisarlos
// ============================================================

type crudAuthDePrueba struct {
	mu     sync.Mutex
	tablas map[string]map[int]map[string]interface{}
	ultimo map[string]int
}

// Nombres del ?query= del CRUD -> nombre del campo en el JSON
var camposQuery = map[string]string{
	"email": "Email", "id_usuario": "IdUsuario", "proposito": "Proposito", "activo": "Activo",
	"nombre_rol": "NombreRol", "proveedor": "Proveedor", "id_externo": "IdExterno",
}

func valorComoTexto(valor interface{}) string {
	if ref, ok := valor.(map[string]interface{}); ok {
		return fmt.Sprint(ref["Id"])
	}
	return fmt.Sprint(valor)
}

func (c *crudAuthDePrueba) guardar(tabla string, registro map[string]interface{}) map[string]interface{} {
	if c.tablas[tabla] == nil {
		c.tablas[tabla] = map[int]map[string]interface{}{}
	}
	c.ultimo[tabla]++
	registro["Id"] = float64(c.ultimo[tabla])
	registro["FechaCreacion"] = time.Now().UTC().Format(time.RFC3339Nano)
	c.tablas[tabla][c.ultimo[tabla]] = registro
	return registro
}

func (c *crudAuthDePrueba) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	responder := func(codigo int, datos interface{}) {
		w.WriteHeader(codigo)
		json.NewEncoder(w).Encode(datos)
	}

	partes := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // [tabla] o [tabla, id]
	tabla := partes[0]
	id := 0
	if len(partes) > 1 {
		id, _ = strconv.Atoi(partes[1])
	}

	switch r.Method {
	case http.MethodPost:
		var registro map[string]interface{}
		json.NewDecoder(r.Body).Decode(&registro)
		// La tabla usuario no deja dos correos iguales (como la base de verdad)
		if tabla == "usuario" {
			for _, otro := range c.tablas["usuario"] {
				if otro["Email"] == registro["Email"] {
					responder(400, "duplicate key value violates unique constraint")
					return
				}
			}
		}
		responder(201, c.guardar(tabla, registro))

	case http.MethodPut:
		var registro map[string]interface{}
		json.NewDecoder(r.Body).Decode(&registro)
		registro["Id"] = float64(id)
		c.tablas[tabla][id] = registro
		responder(200, "OK")

	case http.MethodDelete:
		delete(c.tablas[tabla], id)
		responder(200, "OK")

	case http.MethodGet:
		if id > 0 {
			if registro, existe := c.tablas[tabla][id]; existe {
				responder(200, registro)
			} else {
				responder(200, "<QuerySeter> no row found")
			}
			return
		}

		filtros := map[string]string{}
		if query := r.URL.Query().Get("query"); query != "" {
			for _, par := range strings.Split(query, ",") {
				kv := strings.SplitN(par, ":", 2)
				filtros[camposQuery[kv[0]]] = kv[1]
			}
		}

		lista := []map[string]interface{}{}
		for i := c.ultimo[tabla]; i >= 1; i-- { // del mas nuevo al mas viejo
			registro, existe := c.tablas[tabla][i]
			if !existe {
				continue
			}
			cumple := true
			for campo, valor := range filtros {
				if valorComoTexto(registro[campo]) != valor {
					cumple = false
				}
			}
			if cumple {
				lista = append(lista, registro)
			}
		}
		if r.URL.Query().Get("limit") == "1" && len(lista) > 1 {
			lista = lista[:1]
		}
		responder(200, lista)
	}
}

// contar dice cuantos registros de una tabla cumplen una condicion
func (c *crudAuthDePrueba) contar(tabla string, cumple func(map[string]interface{}) bool) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := 0
	for _, registro := range c.tablas[tabla] {
		if cumple(registro) {
			total++
		}
	}
	return total
}

type correoEnviado struct{ para, asunto, mensaje string }

// entorno prepara todo para una prueba y devuelve el CRUD y los correos enviados
func entorno(t *testing.T, conRolUsuario bool) (*crudAuthDePrueba, *[]correoEnviado) {

	crud := &crudAuthDePrueba{tablas: map[string]map[int]map[string]interface{}{}, ultimo: map[string]int{}}
	crud.guardar("rol", map[string]interface{}{"NombreRol": "Administrador", "Activo": true})
	if conRolUsuario {
		crud.guardar("rol", map[string]interface{}{"NombreRol": "Usuario", "Activo": true})
	}

	servidor := httptest.NewServer(crud)
	t.Cleanup(servidor.Close)
	helpers.UrlCrudAuth = servidor.URL

	correos := &[]correoEnviado{}
	anterior := helpers.EnviarCorreo
	helpers.EnviarCorreo = func(para, asunto, mensaje string) error {
		*correos = append(*correos, correoEnviado{para, asunto, mensaje})
		return nil
	}
	t.Cleanup(func() { helpers.EnviarCorreo = anterior })

	return crud, correos
}

var codigoEnCorreo = regexp.MustCompile(`es: (\d{6})`)

func codigoDelCorreo(t *testing.T, correo correoEnviado) string {
	partes := codigoEnCorreo.FindStringSubmatch(correo.mensaje)
	if partes == nil {
		t.Fatalf("el correo no trae un codigo: %q", correo.mensaje)
	}
	return partes[1]
}

func codigoHttp(err error) int {
	if errorApi, ok := err.(*helpers.ErrorApi); ok {
		return errorApi.Codigo
	}
	return 0
}

// registrar crea un usuario de prueba y devuelve su id
func registrar(t *testing.T, email string) int {
	resultado, err := Registrar(models.DatosRegistro{Nombre: "Sharith", Apellido: "Rodriguez", Email: email, Contrasena: "Clave1234", AceptaTerminos: true})
	if err != nil {
		t.Fatalf("no se pudo registrar: %v", err)
	}
	return resultado["id_usuario"].(int)
}

// sinEsperaEntreCodigos hace que el ultimo codigo parezca de hace 2 minutos
func sinEsperaEntreCodigos(crud *crudAuthDePrueba) {
	crud.mu.Lock()
	defer crud.mu.Unlock()
	for _, codigo := range crud.tablas["codigo_verificacion"] {
		codigo["FechaCreacion"] = time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339Nano)
	}
}

func TestAuthRegistroValidaciones(t *testing.T) {

	entorno(t, true)

	casos := []struct {
		nombre string
		datos  models.DatosRegistro
	}{
		{"clave corta", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "a@b.co", Contrasena: "Ab1", AceptaTerminos: true}},
		{"clave sin mayuscula", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "a@b.co", Contrasena: "clave1234", AceptaTerminos: true}},
		{"clave sin numero", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "a@b.co", Contrasena: "ClaveClave", AceptaTerminos: true}},
		{"clave de mas de 72", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "a@b.co", Contrasena: "A1" + strings.Repeat("x", 80), AceptaTerminos: true}},
		{"correo sin punto", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "ana@correo", Contrasena: "Clave1234", AceptaTerminos: true}},
		{"correo con nombre", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "Ana <a@b.co>", Contrasena: "Clave1234", AceptaTerminos: true}},
		{"nombre corto", models.DatosRegistro{Nombre: "An", Apellido: "Perez", Email: "a@b.co", Contrasena: "Clave1234", AceptaTerminos: true}},
		{"sin terminos", models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "a@b.co", Contrasena: "Clave1234"}},
	}

	for _, caso := range casos {
		if _, err := Registrar(caso.datos); codigoHttp(err) != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, llego %v", caso.nombre, err)
		}
	}
}

func TestAuthRegistroEntregaTokenYNoDejaRepetirCorreo(t *testing.T) {

	entorno(t, true)

	resultado, err := Registrar(models.DatosRegistro{Nombre: "Sharith", Apellido: "Rodriguez", Email: " Sharith@Correo.com ", Contrasena: "Clave1234", AceptaTerminos: true})

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}

	if resultado["token"] == "" || resultado["rol"] != "user" || resultado["email"] != "sharith@correo.com" {
		t.Errorf("respuesta del registro incompleta: %v", resultado)
	}

	if _, err := Registrar(models.DatosRegistro{Nombre: "Otra", Apellido: "Persona", Email: "sharith@correo.com", Contrasena: "Clave1234", AceptaTerminos: true}); codigoHttp(err) != http.StatusConflict {
		t.Errorf("el correo repetido debia dar 409, llego %v", err)
	}
}

func TestAuthRegistroSeDeshaceSiFallaElRol(t *testing.T) {

	// Sin el rol "Usuario" en la base, el registro falla a la mitad
	crud, _ := entorno(t, false)

	if _, err := Registrar(models.DatosRegistro{Nombre: "Ana", Apellido: "Perez", Email: "ana@correo.com", Contrasena: "Clave1234", AceptaTerminos: true}); err == nil {
		t.Fatal("esperaba error al no existir el rol Usuario")
	}

	usuarios := crud.contar("usuario", func(r map[string]interface{}) bool { return true })
	credenciales := crud.contar("credencial", func(r map[string]interface{}) bool { return true })

	if usuarios != 0 || credenciales != 0 {
		t.Errorf("el registro no se deshizo: %d usuarios y %d credenciales", usuarios, credenciales)
	}
}

func TestAuthLoginCompletoConCodigoPorCorreo(t *testing.T) {

	_, correos := entorno(t, true)
	id := registrar(t, "sharith@correo.com")

	// Paso 1: correo y clave -> llega el codigo al correo
	paso1, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"})

	if err != nil || paso1["requiere_codigo"] != true || len(*correos) != 1 || (*correos)[0].para != "sharith@correo.com" {
		t.Fatalf("paso 1 fallo: %v %v correos=%v", paso1, err, *correos)
	}

	codigo := codigoDelCorreo(t, (*correos)[0])

	// Codigo equivocado
	if _, err := VerificarCodigo(models.DatosCodigo{IdUsuario: id, Codigo: "000000"}); codigoHttp(err) != http.StatusUnauthorized && codigo != "000000" {
		t.Errorf("un codigo equivocado debia dar 401, llego %v", err)
	}

	// Paso 2: codigo correcto -> token con el rol
	sesion, err := VerificarCodigo(models.DatosCodigo{IdUsuario: id, Codigo: codigo})

	if err != nil {
		t.Fatalf("paso 2 fallo: %v", err)
	}

	datosToken, err := helpers.LeerToken(sesion["token"].(string))

	if err != nil || datosToken.IdUsuario != id || datosToken.Rol != "user" || sesion["rol"] != "user" {
		t.Errorf("el token no quedo bien: %+v %v %v", datosToken, err, sesion)
	}

	// El mismo codigo no sirve dos veces
	if _, err := VerificarCodigo(models.DatosCodigo{IdUsuario: id, Codigo: codigo}); err == nil {
		t.Error("el codigo se pudo usar dos veces")
	}
}

func TestAuthLoginMismoMensajeSiNoExisteOClaveMal(t *testing.T) {

	entorno(t, true)
	registrar(t, "sharith@correo.com")

	_, errNoExiste := IniciarSesion(models.DatosLogin{Email: "nadie@correo.com", Contrasena: "Clave1234"})
	_, errClaveMal := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Otra1234"})

	if errNoExiste == nil || errClaveMal == nil || errNoExiste.Error() != errClaveMal.Error() || codigoHttp(errNoExiste) != http.StatusUnauthorized {
		t.Errorf("los mensajes deben ser iguales y 401: %v / %v", errNoExiste, errClaveMal)
	}
}

func TestAuthLoginSeBloqueaA5Intentos(t *testing.T) {

	entorno(t, true)
	registrar(t, "sharith@correo.com")

	for i := 0; i < 5; i++ {
		IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Mala1234"})
	}

	// Aunque ahora ponga la clave correcta, sigue bloqueada
	if _, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"}); codigoHttp(err) != http.StatusLocked {
		t.Errorf("esperaba 423 (bloqueada), llego %v", err)
	}
}

func TestAuthEstadoSoloSeRevelaConLaClaveCorrecta(t *testing.T) {

	crud, _ := entorno(t, true)
	id := registrar(t, "sharith@correo.com")

	crud.mu.Lock()
	crud.tablas["usuario"][id]["Estado"] = "suspendido"
	crud.mu.Unlock()

	if _, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Mala1234"}); codigoHttp(err) != http.StatusUnauthorized {
		t.Errorf("con clave mala debe ser el 401 de siempre, llego %v", err)
	}

	if _, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"}); codigoHttp(err) != http.StatusForbidden || !strings.Contains(err.Error(), "suspendida") {
		t.Errorf("con clave correcta debe decir que esta suspendida (403), llego %v", err)
	}
}

func TestAuthUnCodigoPorMinutoYReenviar(t *testing.T) {

	crud, correos := entorno(t, true)
	id := registrar(t, "sharith@correo.com")

	// Reenviar sin haber hecho el paso 1 -> debe iniciar sesion
	if _, err := ReenviarCodigo(models.DatosReenviar{IdUsuario: id}); codigoHttp(err) != http.StatusUnauthorized {
		t.Errorf("reenviar sin paso 1 debia dar 401, llego %v", err)
	}

	IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"})

	// Un segundo login enseguida no manda otro correo, pero responde bien
	if resultado, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"}); err != nil || resultado["requiere_codigo"] != true || len(*correos) != 1 {
		t.Errorf("el segundo login debia responder bien sin otro correo: %v %v (%d correos)", resultado, err, len(*correos))
	}

	// Reenviar antes de un minuto -> 429
	if _, err := ReenviarCodigo(models.DatosReenviar{IdUsuario: id}); codigoHttp(err) != http.StatusTooManyRequests {
		t.Errorf("reenviar antes de un minuto debia dar 429, llego %v", err)
	}

	// Pasado el minuto -> llega un codigo nuevo y el viejo deja de servir
	viejo := codigoDelCorreo(t, (*correos)[0])
	sinEsperaEntreCodigos(crud)

	if _, err := ReenviarCodigo(models.DatosReenviar{IdUsuario: id}); err != nil || len(*correos) != 2 {
		t.Fatalf("reenviar fallo: %v (%d correos)", err, len(*correos))
	}

	nuevo := codigoDelCorreo(t, (*correos)[1])

	if viejo != nuevo {
		if _, err := VerificarCodigo(models.DatosCodigo{IdUsuario: id, Codigo: viejo}); err == nil {
			t.Error("el codigo viejo no debia servir despues de reenviar")
		}
	}

	if _, err := VerificarCodigo(models.DatosCodigo{IdUsuario: id, Codigo: nuevo}); err != nil {
		t.Errorf("el codigo nuevo debia servir: %v", err)
	}
}

func TestAuthSiElCorreoFallaElCodigoNoQueda(t *testing.T) {

	crud, _ := entorno(t, true)
	registrar(t, "sharith@correo.com")

	helpers.EnviarCorreo = func(para, asunto, mensaje string) error {
		return helpers.NuevoError(http.StatusBadGateway, "No fue posible enviar el correo")
	}

	if _, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Clave1234"}); codigoHttp(err) != http.StatusBadGateway {
		t.Errorf("esperaba 502 si el correo falla, llego %v", err)
	}

	activos := crud.contar("codigo_verificacion", func(r map[string]interface{}) bool { return r["Activo"] == true })

	if activos != 0 {
		t.Errorf("quedaron %d codigos activos que nunca llegaron al correo", activos)
	}
}

// googleDePrueba responde como https://oauth2.googleapis.com/tokeninfo
func googleDePrueba(t *testing.T) {

	tokens := map[string]map[string]string{
		"token-ana":           {"aud": "mi-app", "iss": "accounts.google.com", "sub": "g-111", "email": "ana@gmail.com", "email_verified": "true", "given_name": "Ana", "family_name": "Perez"},
		"token-otra-app":      {"aud": "otra-app", "iss": "accounts.google.com", "sub": "g-222", "email": "ana@gmail.com", "email_verified": "true"},
		"token-sin-verificar": {"aud": "mi-app", "iss": "accounts.google.com", "sub": "g-333", "email": "x@gmail.com", "email_verified": "false"},
		"token-sharith":       {"aud": "mi-app", "iss": "https://accounts.google.com", "sub": "g-444", "email": "sharith@correo.com", "email_verified": "true"},
	}

	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		datos, existe := tokens[r.URL.Query().Get("id_token")]
		if !existe {
			w.WriteHeader(400)
			w.Write([]byte(`{"error":"invalid_token"}`))
			return
		}
		json.NewEncoder(w).Encode(datos)
	}))
	t.Cleanup(servidor.Close)

	helpers.UrlGoogleTokenInfo = servidor.URL
	os.Setenv("GOOGLE_CLIENT_ID", "mi-app")
	t.Cleanup(func() { os.Unsetenv("GOOGLE_CLIENT_ID") })
}

func TestAuthGoogle(t *testing.T) {

	crud, _ := entorno(t, true)
	googleDePrueba(t)

	// Usuario nuevo: se crea, se vincula y entra
	sesion, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "google", IdToken: "token-ana"})

	if err != nil || sesion["email"] != "ana@gmail.com" || sesion["token"] == "" || sesion["rol"] != "user" {
		t.Fatalf("login con Google fallo: %v %v", sesion, err)
	}

	// La segunda vez entra con el mismo usuario (no crea otro)
	otraVez, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "google", IdToken: "token-ana"})

	if err != nil || otraVez["id_usuario"] != sesion["id_usuario"] || crud.contar("usuario", func(r map[string]interface{}) bool { return true }) != 1 {
		t.Errorf("la segunda vez debia entrar con el mismo usuario: %v %v", otraVez, err)
	}

	// Tokens que NO deben dejar entrar
	malos := map[string]int{
		"token-otra-app":      http.StatusUnauthorized, // hecho para otra app
		"token-sin-verificar": http.StatusUnauthorized, // correo sin verificar
		"token-inventado":     http.StatusUnauthorized, // Google no lo reconoce
		"":                    http.StatusBadRequest,   // no mando token
	}

	for token, esperado := range malos {
		if _, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "google", IdToken: token}); codigoHttp(err) != esperado {
			t.Errorf("token %q: esperaba %d, llego %v", token, esperado, err)
		}
	}

	if _, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "apple", IdToken: "x"}); codigoHttp(err) != http.StatusNotImplemented {
		t.Errorf("apple debia dar 501, llego %v", err)
	}
}

func TestAuthGoogleVinculaUnaCuentaExistente(t *testing.T) {

	entorno(t, true)
	googleDePrueba(t)
	id := registrar(t, "sharith@correo.com")

	// Google confirma que es dueña de ese correo: entra a SU cuenta
	sesion, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "google", IdToken: "token-sharith"})

	if err != nil || sesion["id_usuario"] != id {
		t.Errorf("debia entrar a la cuenta %d, llego %v %v", id, sesion, err)
	}
}

func TestAuthGoogleSinConfigurar(t *testing.T) {

	entorno(t, true)
	os.Unsetenv("GOOGLE_CLIENT_ID")

	if _, err := IniciarSesionSocial(models.DatosLoginSocial{Proveedor: "google", IdToken: "token-ana"}); codigoHttp(err) != http.StatusServiceUnavailable {
		t.Errorf("sin GOOGLE_CLIENT_ID debia dar 503, llego %v", err)
	}
}

func TestAuthRecuperarContrasena(t *testing.T) {

	crud, correos := entorno(t, true)
	id := registrar(t, "sharith@correo.com")

	// Bloquear la cuenta con 5 intentos malos
	for i := 0; i < 5; i++ {
		IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Mala1234"})
	}

	// Un correo que no existe: misma respuesta y ningun correo enviado
	respuestaNoExiste, _ := PedirCodigoRecuperacion("nadie@correo.com")
	respuesta, err := PedirCodigoRecuperacion("sharith@correo.com")

	if err != nil || respuesta["mensaje"] != respuestaNoExiste["mensaje"] || len(*correos) != 1 {
		t.Fatalf("pedir codigo fallo: %v %v (%d correos)", respuesta, err, len(*correos))
	}

	codigo := codigoDelCorreo(t, (*correos)[0])

	// La clave nueva tambien debe cumplir las reglas
	if _, err := RestablecerContrasena(models.DatosContrasena{Email: "sharith@correo.com", Codigo: codigo, ContrasenaNueva: "debil"}); codigoHttp(err) != http.StatusBadRequest {
		t.Errorf("una clave debil debia dar 400, llego %v", err)
	}

	if _, err := RestablecerContrasena(models.DatosContrasena{Email: "sharith@correo.com", Codigo: codigo, ContrasenaNueva: "Nueva1234"}); err != nil {
		t.Fatalf("restablecer fallo: %v", err)
	}

	// Ya no esta bloqueada y entra con la clave nueva
	if _, err := IniciarSesion(models.DatosLogin{Email: "sharith@correo.com", Contrasena: "Nueva1234"}); err != nil {
		t.Errorf("con la clave nueva debia entrar (y sin bloqueo), llego %v", err)
	}

	cambios := crud.contar("auditoria_login", func(r map[string]interface{}) bool {
		return r["TipoEvento"] == "cambio_contrasena" && valorComoTexto(r["IdUsuario"]) == strconv.Itoa(id)
	})

	if cambios != 1 {
		t.Errorf("el cambio de contrasena debia quedar en la auditoria, hay %d", cambios)
	}
}

func TestAuthCambiarContrasenaSeBloqueaA5Intentos(t *testing.T) {

	entorno(t, true)
	id := registrar(t, "sharith@correo.com")

	for i := 0; i < 5; i++ {
		if _, err := CambiarContrasena(id, models.DatosContrasena{ContrasenaActual: "Mala1234", ContrasenaNueva: "Nueva1234"}); codigoHttp(err) != http.StatusUnauthorized {
			t.Fatalf("intento %d: esperaba 401, llego %v", i+1, err)
		}
	}

	if _, err := CambiarContrasena(id, models.DatosContrasena{ContrasenaActual: "Clave1234", ContrasenaNueva: "Nueva1234"}); codigoHttp(err) != http.StatusLocked {
		t.Errorf("despues de 5 intentos debia dar 423, llego %v", err)
	}
}

func TestAuthCerrarSesionInvalidaElToken(t *testing.T) {

	entorno(t, true)
	sesion, _ := Registrar(models.DatosRegistro{Nombre: "Sharith", Apellido: "Rodriguez", Email: "sharith@correo.com", Contrasena: "Clave1234", AceptaTerminos: true})

	datos, err := helpers.LeerToken(sesion["token"].(string))

	if err != nil {
		t.Fatalf("el token del registro no sirve: %v", err)
	}

	CerrarSesion(datos, "127.0.0.1", "prueba")

	if _, err := helpers.LeerToken(sesion["token"].(string)); err == nil {
		t.Error("despues del logout el token no debia servir")
	}
}
