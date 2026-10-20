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
// TOKEN DE SESION (JWT firmado con HS256)
// Cuando el usuario termina el login, el MID le entrega un token.
// La pagina lo guarda y lo manda en cada peticion en el encabezado:
//   Authorization: Bearer <token>
// El filtro (filtro.go) lo revisa antes de dejar pasar la peticion.
//
// El token tiene 3 partes separadas por puntos:
//   encabezado . datos (id del usuario, rol, vencimiento) . firma
// La firma se hace con JWT_SECRETO (del .env). Si alguien cambia
// los datos, la firma ya no coincide y el token se rechaza.
// ============================================================

// DuracionToken: cuanto dura una sesion antes de pedir login otra vez.
var DuracionToken = 8 * time.Hour

// DatosToken: lo que va guardado dentro del token.
type DatosToken struct {
	IdUsuario int    `json:"sub"`
	Rol       string `json:"rol"`
	Emitido   int64  `json:"iat"`
	Expira    int64  `json:"exp"`
	Id        string `json:"jti"` // id unico del token (para cerrar sesion)
}

var (
	secretoToken     []byte
	cargarSecretoUna sync.Once

	// Tokens cerrados con logout (id -> vencimiento). Viven en memoria:
	// si el MID se reinicia la lista se borra, pero esos tokens igual
	// vencen solos en DuracionToken.
	tokensRevocados      = map[string]int64{}
	tokensRevocadosMutex sync.Mutex
)

// CargarSeguridad lee JWT_SECRETO del .env. Se llama una vez desde main.go.
// Si no existe, se inventa uno al azar y se avisa en la consola
// (las sesiones se pierden cada vez que el MID se reinicia).
func CargarSeguridad() {
	cargarSecretoUna.Do(func() {

		if valor := os.Getenv("JWT_SECRETO"); len(valor) >= 32 {
			secretoToken = []byte(valor)
			return
		} else if valor != "" {
			fmt.Println("[AVISO] JWT_SECRETO es muy corto (minimo 32 caracteres); se usa uno temporal.")
		} else {
			fmt.Println("[AVISO] JWT_SECRETO no esta en el .env; se usa uno temporal (las sesiones se pierden al reiniciar).")
		}

		aleatorio := make([]byte, 32)
		rand.Read(aleatorio)
		secretoToken = aleatorio
	})
}

// CrearToken arma y firma un token para el usuario.
func CrearToken(idUsuario int, rol string) (string, error) {

	CargarSeguridad()

	idToken := make([]byte, 16)
	rand.Read(idToken)

	ahora := time.Now().UTC()

	datos := DatosToken{
		IdUsuario: idUsuario,
		Rol:       rol,
		Emitido:   ahora.Unix(),
		Expira:    ahora.Add(DuracionToken).Unix(),
		Id:        hex.EncodeToString(idToken),
	}

	encabezado := codificar([]byte(`{"alg":"HS256","typ":"JWT"}`))

	jsonDatos, err := json.Marshal(datos)

	if err != nil {
		return "", NuevoError(http.StatusInternalServerError, "No fue posible crear la sesion")
	}

	contenido := encabezado + "." + codificar(jsonDatos)

	return contenido + "." + firmar(contenido), nil
}

// LeerToken revisa la firma, que no este vencido y que no se haya
// cerrado la sesion. Si todo esta bien devuelve los datos del token.
func LeerToken(token string) (DatosToken, error) {

	CargarSeguridad()

	invalido := NuevoError(http.StatusUnauthorized, "La sesion no es valida. Inicia sesion de nuevo.")

	partes := strings.Split(strings.TrimSpace(token), ".")

	if len(partes) != 3 {
		return DatosToken{}, invalido
	}

	// 1. La firma debe coincidir (se compara de forma segura)
	firmaRecibida, err := base64.RawURLEncoding.DecodeString(partes[2])

	if err != nil || !hmac.Equal(firmaRecibida, calcularFirma(partes[0]+"."+partes[1])) {
		return DatosToken{}, invalido
	}

	// 2. Solo se aceptan tokens HS256 (los que crea este MID)
	var encabezado struct {
		Alg string `json:"alg"`
	}

	if jsonEncabezado, err := base64.RawURLEncoding.DecodeString(partes[0]); err != nil || json.Unmarshal(jsonEncabezado, &encabezado) != nil || encabezado.Alg != "HS256" {
		return DatosToken{}, invalido
	}

	// 3. Leer los datos
	var datos DatosToken

	jsonDatos, err := base64.RawURLEncoding.DecodeString(partes[1])

	if err != nil || json.Unmarshal(jsonDatos, &datos) != nil || datos.IdUsuario <= 0 {
		return DatosToken{}, invalido
	}

	// 4. Que no este vencido ni cerrado
	if time.Now().UTC().Unix() >= datos.Expira {
		return DatosToken{}, NuevoError(http.StatusUnauthorized, "La sesion vencio. Inicia sesion de nuevo.")
	}

	tokensRevocadosMutex.Lock()
	_, revocado := tokensRevocados[datos.Id]
	tokensRevocadosMutex.Unlock()

	if revocado {
		return DatosToken{}, invalido
	}

	return datos, nil
}

// RevocarToken cierra la sesion de ese token (logout).
func RevocarToken(datos DatosToken) {

	tokensRevocadosMutex.Lock()
	defer tokensRevocadosMutex.Unlock()

	// Se aprovecha para borrar los que ya vencieron solos
	ahora := time.Now().UTC().Unix()
	for id, expira := range tokensRevocados {
		if expira < ahora {
			delete(tokensRevocados, id)
		}
	}

	tokensRevocados[datos.Id] = datos.Expira
}

func codificar(datos []byte) string {
	return base64.RawURLEncoding.EncodeToString(datos)
}

func calcularFirma(contenido string) []byte {
	mac := hmac.New(sha256.New, secretoToken)
	mac.Write([]byte(contenido))
	return mac.Sum(nil)
}

func firmar(contenido string) string {
	return codificar(calcularFirma(contenido))
}
