package helpers

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

// Pruebas del token de sesion. Se corren con: go test ./helpers

func TestTokenSeCreaYSeLee(t *testing.T) {

	token, err := CrearToken(7, "admin")

	if err != nil {
		t.Fatalf("no se pudo crear: %v", err)
	}

	datos, err := LeerToken(token)

	if err != nil || datos.IdUsuario != 7 || datos.Rol != "admin" {
		t.Errorf("el token no trae lo que se guardo: %+v %v", datos, err)
	}
}

func TestTokenCambiadoNoSirve(t *testing.T) {

	token, _ := CrearToken(7, "user")
	partes := strings.Split(token, ".")

	// Alguien cambia los datos para hacerse pasar por admin (usuario 1)
	falso := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":1,"rol":"admin","exp":9999999999,"jti":"x"}`))

	if _, err := LeerToken(partes[0] + "." + falso + "." + partes[2]); err == nil {
		t.Error("un token con los datos cambiados no debia servir")
	}

	// Token sin firma con alg "none"
	sinFirma := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))

	if _, err := LeerToken(sinFirma + "." + falso + "."); err == nil {
		t.Error("un token sin firma no debia servir")
	}

	for _, basura := range []string{"", "abc", "a.b", "a.b.c.d"} {
		if _, err := LeerToken(basura); err == nil {
			t.Errorf("%q no debia servir", basura)
		}
	}
}

func TestTokenVencidoNoSirve(t *testing.T) {

	original := DuracionToken
	DuracionToken = -time.Minute
	defer func() { DuracionToken = original }()

	token, _ := CrearToken(7, "user")

	if _, err := LeerToken(token); err == nil {
		t.Error("un token vencido no debia servir")
	}
}

func TestTokenRevocadoNoSirve(t *testing.T) {

	token, _ := CrearToken(7, "user")
	datos, _ := LeerToken(token)

	RevocarToken(datos)

	if _, err := LeerToken(token); err == nil {
		t.Error("despues del logout el token no debia servir")
	}

	// Otro token del mismo usuario sigue sirviendo (solo se cerro esa sesion)
	otro, _ := CrearToken(7, "user")

	if _, err := LeerToken(otro); err != nil {
		t.Errorf("la otra sesion debia seguir activa: %v", err)
	}
}
