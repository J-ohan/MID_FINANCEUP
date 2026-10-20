package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ============================================================
// LOGIN CON GOOGLE
// La pagina usa el boton de Google (Google Identity Services) y
// recibe un "id_token" firmado por Google. Ese token es lo unico
// que la pagina manda al MID.
//
// Aqui el MID le pregunta a Google si el token es real y si fue
// hecho para nuestra app (GOOGLE_CLIENT_ID del .env). Solo si todo
// esta bien se usan el correo y el id que Google confirma.
// Nunca se confia en un correo que mande la pagina.
// ============================================================

// UrlGoogleTokenInfo es la direccion de Google que revisa los tokens.
// Es variable para que las pruebas usen un Google de mentira.
var UrlGoogleTokenInfo = "https://oauth2.googleapis.com/tokeninfo"

var clienteGoogle = &http.Client{Timeout: 10 * time.Second}

// DatosGoogle: lo que Google confirma del usuario.
type DatosGoogle struct {
	IdExterno string // "sub": id unico del usuario en Google
	Email     string
	Nombre    string
	Apellido  string
}

// VerificarTokenGoogle revisa el id_token con Google y devuelve los datos confirmados.
func VerificarTokenGoogle(idToken string) (DatosGoogle, error) {

	idCliente := os.Getenv("GOOGLE_CLIENT_ID")

	if idCliente == "" {
		return DatosGoogle{}, NuevoError(http.StatusServiceUnavailable, "El inicio de sesion con Google no esta configurado")
	}

	if idToken == "" {
		return DatosGoogle{}, NuevoError(http.StatusBadRequest, "Falta el id_token de Google")
	}

	tokenInvalido := NuevoError(http.StatusUnauthorized, "El token de Google no es valido")

	// 1. Preguntarle a Google
	response, err := clienteGoogle.Get(UrlGoogleTokenInfo + "?id_token=" + url.QueryEscape(idToken))

	if err != nil {
		return DatosGoogle{}, NuevoError(http.StatusBadGateway, "No fue posible comunicarse con Google")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return DatosGoogle{}, tokenInvalido
	}

	body, _ := io.ReadAll(response.Body)

	// Google responde todo como texto, incluso "email_verified": "true"
	var datos struct {
		Aud           string `json:"aud"`
		Iss           string `json:"iss"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}

	if err := json.Unmarshal(body, &datos); err != nil {
		return DatosGoogle{}, tokenInvalido
	}

	// 2. El token debe ser para NUESTRA app y venir de Google
	if datos.Aud != idCliente {
		return DatosGoogle{}, tokenInvalido
	}

	if datos.Iss != "accounts.google.com" && datos.Iss != "https://accounts.google.com" {
		return DatosGoogle{}, tokenInvalido
	}

	// 3. El correo debe estar verificado por Google
	if datos.Sub == "" || datos.Email == "" || datos.EmailVerified != "true" {
		return DatosGoogle{}, NuevoError(http.StatusUnauthorized, "El correo de Google no esta verificado")
	}

	return DatosGoogle{
		IdExterno: datos.Sub,
		Email:     datos.Email,
		Nombre:    datos.GivenName,
		Apellido:  datos.FamilyName,
	}, nil
}
