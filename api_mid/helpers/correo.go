package helpers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/oauth2"
)

// ============================================================
// CORREO CON GMAIL
// Archivos (dentro de api_mid, NO se suben a git):
//   credentials/credentials.json -> se descarga de Google Cloud Console
//   token/token.json             -> se crea solo la primera vez que se autoriza
// ============================================================

const (
	archivoCredenciales = "credentials/credentials.json"
	archivoToken        = "token/token.json"
	urlEnviarGmail      = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
	permisoEnviarGmail  = "https://www.googleapis.com/auth/gmail.send"
)

// clienteGmail se crea una sola vez al arrancar (CargarCorreo) y se reusa.
var clienteGmail *http.Client

// CargarCorreo prepara Gmail al arrancar el MID (se llama en main.go).
// Si todavia no existe token.json, pide la autorizacion en la consola.
func CargarCorreo() {

	configuracion, err := leerCredenciales()
	if err != nil {
		fmt.Println("[AVISO] Gmail no configurado:", err)
		return
	}

	token, err := leerToken()
	if err != nil {
		if token, err = autorizarEnConsola(configuracion); err != nil {
			fmt.Println("[AVISO] Gmail no autorizado:", err)
			return
		}
	}

	// El cliente renueva el permiso solo cuando vence
	clienteGmail = configuracion.Client(context.Background(), token)
}

// EnviarCorreo manda un correo de texto con Gmail.
// Sin Gmail configurado: en modo dev lo muestra en la consola; si no, error 503.
func EnviarCorreo(para string, asunto string, cuerpo string) error {

	if clienteGmail == nil {
		if beego.BConfig.RunMode == "dev" {
			fmt.Printf("[CORREO - SOLO DESARROLLO] Para: %s | %s\n%s\n", para, asunto, cuerpo)
			return nil
		}
		return NuevoError(http.StatusServiceUnavailable, "El servicio de correo no esta configurado")
	}

	// Mensaje en formato MIME (el asunto se codifica para que salgan las tildes)
	mensaje := strings.Join([]string{
		"To: " + para,
		"Subject: " + mime.QEncoding.Encode("utf-8", asunto),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		cuerpo,
	}, "\r\n")

	cuerpoPeticion, _ := json.Marshal(map[string]string{
		"raw": base64.RawURLEncoding.EncodeToString([]byte(mensaje)),
	})

	respuesta, err := clienteGmail.Post(urlEnviarGmail, "application/json", bytes.NewReader(cuerpoPeticion))
	if err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible comunicarse con Gmail")
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		return NuevoError(http.StatusBadGateway, "Gmail no pudo enviar el correo")
	}

	return nil
}

// leerCredenciales arma la configuracion OAuth con credentials.json
// (sirve tanto para credenciales de "Aplicacion de escritorio" como "Web").
func leerCredenciales() (*oauth2.Config, error) {

	contenido, err := os.ReadFile(archivoCredenciales)
	if err != nil {
		return nil, fmt.Errorf("no se encontro %s", archivoCredenciales)
	}

	type datosCliente struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	var archivo struct {
		Installed *datosCliente `json:"installed"`
		Web       *datosCliente `json:"web"`
	}

	if err := json.Unmarshal(contenido, &archivo); err != nil {
		return nil, fmt.Errorf("%s no es valido", archivoCredenciales)
	}

	datos := archivo.Installed
	if datos == nil {
		datos = archivo.Web
	}

	if datos == nil || datos.ClientID == "" || len(datos.RedirectURIs) == 0 {
		return nil, fmt.Errorf("%s no tiene los datos del cliente", archivoCredenciales)
	}

	return &oauth2.Config{
		ClientID:     datos.ClientID,
		ClientSecret: datos.ClientSecret,
		RedirectURL:  datos.RedirectURIs[0],
		Scopes:       []string{permisoEnviarGmail},
		Endpoint:     oauth2.Endpoint{AuthURL: datos.AuthURI, TokenURL: datos.TokenURI},
	}, nil
}

// leerToken lee el permiso guardado en token.json.
func leerToken() (*oauth2.Token, error) {

	contenido, err := os.ReadFile(archivoToken)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(contenido, token); err != nil {
		return nil, err
	}

	return token, nil
}

// autorizarEnConsola pide el permiso de Gmail UNA vez y lo guarda en token.json:
// muestra un enlace, la persona lo abre, acepta y pega el codigo en la consola.
func autorizarEnConsola(configuracion *oauth2.Config) (*oauth2.Token, error) {

	fmt.Println("==================== AUTORIZAR GMAIL ====================")
	fmt.Println("1. Abre este enlace en el navegador y acepta los permisos:")
	fmt.Println(configuracion.AuthCodeURL("financeup", oauth2.AccessTypeOffline))
	fmt.Println("2. Copia el codigo (parametro \"code\" de la direccion) y pegalo aqui:")

	var codigo string
	if _, err := fmt.Scan(&codigo); err != nil {
		return nil, fmt.Errorf("no se pudo leer el codigo")
	}

	token, err := configuracion.Exchange(context.Background(), strings.TrimSpace(codigo))
	if err != nil {
		return nil, fmt.Errorf("Google no acepto el codigo")
	}

	// Guardar el permiso para no pedirlo otra vez
	if err := os.MkdirAll("token", 0700); err != nil {
		return nil, err
	}

	contenido, _ := json.Marshal(token)
	if err := os.WriteFile(archivoToken, contenido, 0600); err != nil {
		return nil, err
	}

	fmt.Println("Gmail autorizado. Permiso guardado en", archivoToken)

	return token, nil
}
