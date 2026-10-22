package services

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// VERIFICACION DE CORREO
// Comprueba que una persona es duena de un correo: se le envia
// un codigo y debe devolverlo. Los codigos viven en memoria.
// ============================================================

const (
	minutosCodigoCorreo     = 5 // vigencia del codigo
	maxIntentosCodigoCorreo = 5 // intentos antes de anularlo
)

// codigoCorreo es un codigo pendiente de un correo.
type codigoCorreo struct {
	codigo   string
	creado   time.Time
	expira   time.Time
	intentos int
}

var (
	codigosCorreo      = map[string]codigoCorreo{} // correo -> codigo pendiente
	correosVerificados = map[string]time.Time{}    // correo -> momento de verificacion, consumido al registrar
	mutexCodigosCorreo sync.Mutex                  // evita que dos peticiones lo cambien a la vez
)

// EnviarVerificacionCorreo genera un codigo de 6 numeros y lo envia al correo.
func EnviarVerificacionCorreo(email string) (map[string]interface{}, error) {

	email, err := normalizarEmail(email)
	if err != nil {
		return nil, err
	}

	ahora := time.Now()
	codigo := helpers.NumeroAleatorio()

	mutexCodigosCorreo.Lock()
	limpiarCodigosVencidos(ahora)

	// Maximo un codigo por minuto
	if anterior, existe := codigosCorreo[email]; existe && ahora.Sub(anterior.creado) < segundosEntreCodigos*time.Second {
		mutexCodigosCorreo.Unlock()
		return nil, helpers.NuevoError(http.StatusTooManyRequests, "Ya te enviamos un codigo hace menos de un minuto. Revisa tu correo o espera un momento.")
	}

	codigosCorreo[email] = codigoCorreo{codigo: codigo, creado: ahora, expira: ahora.Add(minutosCodigoCorreo * time.Minute)}
	mutexCodigosCorreo.Unlock()

	mensaje := fmt.Sprintf("Hola,\n\nTu codigo para verificar este correo es: %s\n\nVence en %d minutos.\nSi no pediste este codigo, ignora este mensaje.\n\nFinanceUp",
		codigo, minutosCodigoCorreo)

	// Si el correo no sale, el codigo no sirve
	if err := helpers.EnviarCorreo(email, "Verifica tu correo en FinanceUp", mensaje); err != nil {
		mutexCodigosCorreo.Lock()
		delete(codigosCorreo, email)
		mutexCodigosCorreo.Unlock()
		return nil, err
	}

	return map[string]interface{}{
		"mensaje":   "Te enviamos un codigo de verificacion a tu correo",
		"email":     email,
		"expira_en": minutosCodigoCorreo * 60, // segundos
	}, nil
}

// ConfirmarVerificacionCorreo revisa el codigo; si es correcto, el correo queda verificado.
func ConfirmarVerificacionCorreo(datos models.DatosVerificarCorreo) (map[string]interface{}, error) {

	email, err := normalizarEmail(datos.Email)
	if err != nil {
		return nil, err
	}

	codigoEscrito := strings.TrimSpace(datos.Codigo)
	if codigoEscrito == "" {
		return nil, helpers.NuevoError(http.StatusBadRequest, "El codigo es obligatorio")
	}

	mutexCodigosCorreo.Lock()
	defer mutexCodigosCorreo.Unlock()

	pendiente, existe := codigosCorreo[email]
	if !existe {
		return nil, helpers.NuevoError(http.StatusNotFound, "No hay un codigo pendiente para este correo. Pide uno nuevo.")
	}

	if time.Now().After(pendiente.expira) {
		delete(codigosCorreo, email)
		return nil, helpers.NuevoError(http.StatusUnauthorized, "El codigo vencio. Pide uno nuevo.")
	}

	// Comparacion segura (tarda lo mismo aunque se equivoque en el primer numero)
	if subtle.ConstantTimeCompare([]byte(codigoEscrito), []byte(pendiente.codigo)) != 1 {

		pendiente.intentos++

		if pendiente.intentos >= maxIntentosCodigoCorreo {
			delete(codigosCorreo, email)
			return nil, helpers.NuevoError(http.StatusUnauthorized, "Superaste los intentos. Pide un codigo nuevo.")
		}

		codigosCorreo[email] = pendiente

		return nil, helpers.NuevoError(http.StatusUnauthorized, fmt.Sprintf("El codigo no es correcto. Te quedan %d intentos.", maxIntentosCodigoCorreo-pendiente.intentos))
	}

	// Correcto: se borra para que no sirva dos veces
	delete(codigosCorreo, email)
	correosVerificados[email] = time.Now().Add(30 * time.Minute)

	return map[string]interface{}{"mensaje": "Correo verificado", "email": email, "verificado": true}, nil
}

// correoVerificado dice si el correo fue verificado y sigue vigente, sin consumirlo.
func correoVerificado(email string) bool {
	mutexCodigosCorreo.Lock()
	defer mutexCodigosCorreo.Unlock()

	expira, existe := correosVerificados[email]
	return existe && time.Now().Before(expira)
}

// consumirCorreoVerificado impide completar el registro si no se confirmo el codigo.
func consumirCorreoVerificado(email string) bool {
	mutexCodigosCorreo.Lock()
	defer mutexCodigosCorreo.Unlock()

	expira, existe := correosVerificados[email]
	delete(correosVerificados, email)
	return existe && time.Now().Before(expira)
}

// limpiarCodigosVencidos borra los codigos viejos para no llenar la memoria.
// Se llama con el mutex tomado.
func limpiarCodigosVencidos(ahora time.Time) {
	for email, pendiente := range codigosCorreo {
		if ahora.After(pendiente.expira) {
			delete(codigosCorreo, email)
		}
	}
}
