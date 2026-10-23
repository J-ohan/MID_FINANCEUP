package helpers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ============================================================
// TOKEN DE SESION (JWT HS256)
// La pagina lo manda en cada peticion: Authorization: Bearer <token>
// Se firma con JWT_SECRETO del .env; si alguien lo cambia, la firma no coincide.
// ============================================================

// DuracionToken es cuanto dura una sesion.
const DuracionToken = 8 * time.Hour

// Token temporal del login: se entrega despues de correo y clave, y solo
// sirve para verificar o reenviar el codigo (el filtro no lo deja pasar a mas).
const (
	RolVerificacion           = "verificacion_2fa"
	DuracionTokenVerificacion = 10 * time.Minute
)

// DatosToken es lo que va guardado dentro del token.
type DatosToken struct {
	IdUsuario int    `json:"sub"`
	Rol       string `json:"rol"`
	Emitido   int64  `json:"iat"`
	Expira    int64  `json:"exp"`
	Id        string `json:"jti"` // id unico, para poder cerrar esa sesion
}

var (
	secretoToken []byte
	cargarUnaVez sync.Once

	// Sesiones cerradas con logout (se borran al reiniciar; igual vencen solas)
	tokensRevocados = map[string]int64{}
	mutexRevocados  sync.Mutex

	// Usuarios a los que un admin les cambio el rol o el estado:
	// sus tokens creados hasta esa hora ya no sirven
	sesionesCerradasDesde = map[int]int64{}
)

// CargarSeguridad lee JWT_SECRETO (minimo 32 caracteres). Se llama en main.go.
// Si no existe, usa uno temporal: las sesiones se pierden al reiniciar.
func CargarSeguridad() {
	cargarUnaVez.Do(func() {

		if valor := os.Getenv("JWT_SECRETO"); len(valor) >= 32 {
			secretoToken = []byte(valor)
			return
		}

		fmt.Println("[AVISO] JWT_SECRETO falta o es muy corto en el .env; se usa uno temporal.")
		secretoToken = make([]byte, 32)
		rand.Read(secretoToken)
	})
}

// CrearToken arma el token de la sesion (dura 8 horas).
func CrearToken(idUsuario int, rol string) (string, error) {
	return crearTokenCon(idUsuario, rol, DuracionToken)
}

// CrearTokenVerificacion arma el token temporal del paso 2 del login.
func CrearTokenVerificacion(idUsuario int) (string, error) {
	return crearTokenCon(idUsuario, RolVerificacion, DuracionTokenVerificacion)
}

// LeerTokenVerificacion revisa el token temporal y devuelve el id del usuario.
func LeerTokenVerificacion(token string) (int, error) {

	datos, err := LeerToken(token)
	if err != nil || datos.Rol != RolVerificacion {
		return 0, NuevoError(http.StatusUnauthorized, "Inicia sesion de nuevo para recibir un codigo")
	}

	return datos.IdUsuario, nil
}

// crearTokenCon arma y firma un token con el rol y la duracion que se pidan.
func crearTokenCon(idUsuario int, rol string, duracion time.Duration) (string, error) {

	CargarSeguridad()

	idUnico := make([]byte, 16)
	rand.Read(idUnico)

	ahora := time.Now().UTC()

	datos, err := json.Marshal(DatosToken{
		IdUsuario: idUsuario,
		Rol:       rol,
		Emitido:   ahora.Unix(),
		Expira:    ahora.Add(duracion).Unix(),
		Id:        hex.EncodeToString(idUnico),
	})
	if err != nil {
		return "", NuevoError(http.StatusInternalServerError, "No fue posible crear la sesion")
	}

	contenido := codificar([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + codificar(datos)

	return contenido + "." + codificar(calcularFirma(contenido)), nil
}

// LeerToken revisa firma, vencimiento y logout; si todo esta bien devuelve sus datos.
func LeerToken(token string) (DatosToken, error) {

	CargarSeguridad()

	invalido := NuevoError(http.StatusUnauthorized, "La sesion no es valida. Inicia sesion de nuevo.")

	partes := strings.Split(strings.TrimSpace(token), ".")
	if len(partes) != 3 {
		return DatosToken{}, invalido
	}

	// 1. La firma debe coincidir
	firma, err := base64.RawURLEncoding.DecodeString(partes[2])
	if err != nil || !hmac.Equal(firma, calcularFirma(partes[0]+"."+partes[1])) {
		return DatosToken{}, invalido
	}

	// 2. Solo tokens HS256 (los que crea este MID)
	var encabezado struct {
		Alg string `json:"alg"`
	}
	if texto, err := base64.RawURLEncoding.DecodeString(partes[0]); err != nil || json.Unmarshal(texto, &encabezado) != nil || encabezado.Alg != "HS256" {
		return DatosToken{}, invalido
	}

	// 3. Los datos
	var datos DatosToken
	texto, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil || json.Unmarshal(texto, &datos) != nil || datos.IdUsuario <= 0 {
		return DatosToken{}, invalido
	}

	// 4. No vencido ni cerrado
	if time.Now().UTC().Unix() >= datos.Expira {
		return DatosToken{}, NuevoError(http.StatusUnauthorized, "La sesion vencio. Inicia sesion de nuevo.")
	}

	mutexRevocados.Lock()
	_, cerrado := tokensRevocados[datos.Id]
	desde, cerradoPorAdmin := sesionesCerradasDesde[datos.IdUsuario]
	mutexRevocados.Unlock()

	if cerrado || (cerradoPorAdmin && datos.Emitido <= desde) {
		return DatosToken{}, invalido
	}

	return datos, nil
}

// RevocarToken cierra esa sesion (logout).
func RevocarToken(datos DatosToken) {

	mutexRevocados.Lock()
	defer mutexRevocados.Unlock()

	// De paso se borran los que ya vencieron
	ahora := time.Now().UTC().Unix()
	for id, expira := range tokensRevocados {
		if expira < ahora {
			delete(tokensRevocados, id)
		}
	}

	tokensRevocados[datos.Id] = datos.Expira
}

// CerrarSesionesDeUsuario invalida todos los tokens que ya tiene ese usuario
// (cuando un admin le cambia el rol o lo suspende). Debe iniciar sesion de nuevo.
func CerrarSesionesDeUsuario(idUsuario int) {

	mutexRevocados.Lock()
	defer mutexRevocados.Unlock()

	// De paso se borran los de hace mas de 8 horas (esos tokens ya vencieron)
	ahora := time.Now().UTC().Unix()
	for id, desde := range sesionesCerradasDesde {
		if ahora-desde > int64(DuracionToken.Seconds()) {
			delete(sesionesCerradasDesde, id)
		}
	}

	sesionesCerradasDesde[idUsuario] = ahora
}

func codificar(datos []byte) string {
	return base64.RawURLEncoding.EncodeToString(datos)
}

func calcularFirma(contenido string) []byte {
	mac := hmac.New(sha256.New, secretoToken)
	mac.Write([]byte(contenido))
	return mac.Sum(nil)
}
