package helpers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ============================================================
// LOGIN CON GOOGLE
// La pagina manda el "id_token" que le dio Google y aqui se le
// pregunta a Google si es real y si es para nuestra app (GOOGLE_CLIENT_ID).
// Nunca se confia en un correo que mande la pagina.
// ============================================================

const urlGoogleTokenInfo = "https://oauth2.googleapis.com/tokeninfo"

var clienteGoogle = &http.Client{Timeout: 10 * time.Second}

// DatosGoogle es lo que Google confirma del usuario.
type DatosGoogle struct {
	IdExterno string // "sub": id unico en Google
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

	invalido := NuevoError(http.StatusUnauthorized, "El token de Google no es valido")

	respuesta, err := clienteGoogle.Get(urlGoogleTokenInfo + "?id_token=" + url.QueryEscape(idToken))
	if err != nil {
		return DatosGoogle{}, NuevoError(http.StatusBadGateway, "No fue posible comunicarse con Google")
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		return DatosGoogle{}, invalido
	}

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

	if err := json.NewDecoder(respuesta.Body).Decode(&datos); err != nil {
		return DatosGoogle{}, invalido
	}

	// Debe ser para NUESTRA app, venir de Google y tener el correo verificado
	if datos.Aud != idCliente || (datos.Iss != "accounts.google.com" && datos.Iss != "https://accounts.google.com") {
		return DatosGoogle{}, invalido
	}

	if datos.Sub == "" || datos.Email == "" || datos.EmailVerified != "true" {
		return DatosGoogle{}, NuevoError(http.StatusUnauthorized, "El correo de Google no esta verificado")
	}

	return DatosGoogle{IdExterno: datos.Sub, Email: datos.Email, Nombre: datos.GivenName, Apellido: datos.FamilyName}, nil
}
