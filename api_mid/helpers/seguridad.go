package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// Cifrar convierte una clave o un codigo en un texto cifrado
// con bcrypt. Lo cifrado nunca se puede volver a leer: solo
// se puede comparar con CoincideCifrado.
func Cifrar(texto string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(texto), bcrypt.DefaultCost)
	return string(hash)
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
