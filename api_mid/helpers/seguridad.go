package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LargoMaximoClave es el limite de bcrypt: no puede cifrar mas de 72 bytes.
const LargoMaximoClave = 72

// Cifrar convierte una clave o un codigo en un texto cifrado
// con bcrypt. Lo cifrado nunca se puede volver a leer: solo
// se puede comparar con CoincideCifrado.
// Si no se puede cifrar (por ejemplo, el texto es muy largo)
// devuelve un error en vez de guardar un hash vacio.
func Cifrar(texto string) (string, error) {

	if len(texto) > LargoMaximoClave {
		return "", NuevoError(http.StatusBadRequest, fmt.Sprintf("La contrasena no puede tener mas de %d caracteres", LargoMaximoClave))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(texto), bcrypt.DefaultCost)

	if err != nil {
		return "", NuevoError(http.StatusInternalServerError, "No fue posible cifrar el dato")
	}

	return string(hash), nil
}

// CoincideCifrado dice si "texto" es el mismo que se cifro en "hash".
func CoincideCifrado(hash string, texto string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(texto)) == nil
}

// NumeroAleatorio devuelve 6 numeros al azar, ej: "048213".
// Se usa para los codigos de verificacion y los radicados.
func NumeroAleatorio() string {
	numero, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", numero.Int64())
}
