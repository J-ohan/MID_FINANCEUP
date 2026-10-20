package helpers

import (
	"fmt"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
)

// ============================================================
// CORREO
// Envia correos por SMTP (por ejemplo Gmail con una "contrasena
// de aplicacion", o cualquier servicio SMTP).
//
// Los datos salen del archivo .env:
//   SMTP_HOST=smtp.gmail.com
//   SMTP_PORT=587
//   SMTP_USUARIO=tu_correo@gmail.com
//   SMTP_CLAVE=contrasena_de_aplicacion
//   CORREO_REMITENTE=FinanceUp <tu_correo@gmail.com>
//
// Si el .env no tiene SMTP_HOST:
//   - en modo "dev" el correo se muestra en la consola del MID
//     (solo para probar en tu PC, nunca en un servidor);
//   - en cualquier otro modo responde error 503.
// ============================================================

// EnviarCorreo es una variable para que las pruebas puedan
// reemplazarla por una que no envie nada de verdad.
var EnviarCorreo = enviarCorreoSMTP

// CorreoConfigurado dice si el .env tiene los datos del servidor SMTP.
func CorreoConfigurado() bool {
	return os.Getenv("SMTP_HOST") != ""
}

func enviarCorreoSMTP(para string, asunto string, mensaje string) error {

	if !CorreoConfigurado() {

		if beego.BConfig.RunMode == "dev" {
			fmt.Printf("[CORREO NO CONFIGURADO - SOLO DESARROLLO] Para: %s | %s\n%s\n", para, asunto, mensaje)
			return nil
		}

		return NuevoError(http.StatusServiceUnavailable, "El servicio de correo no esta configurado")
	}

	host := os.Getenv("SMTP_HOST")
	puerto := os.Getenv("SMTP_PORT")
	if puerto == "" {
		puerto = "587" // puerto con STARTTLS, el que usa smtp.SendMail
	}

	usuario := os.Getenv("SMTP_USUARIO")
	remitente := os.Getenv("CORREO_REMITENTE")
	if remitente == "" {
		remitente = usuario
	}

	// Direccion sola del remitente (sin el nombre), para el sobre del correo
	direccionRemitente := remitente
	if inicio, fin := strings.Index(remitente, "<"), strings.Index(remitente, ">"); inicio >= 0 && fin > inicio {
		direccionRemitente = remitente[inicio+1 : fin]
	}

	// El asunto se codifica para que las tildes se vean bien
	contenido := "From: " + remitente + "\r\n" +
		"To: " + para + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", asunto) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" + mensaje + "\r\n"

	autenticacion := smtp.PlainAuth("", usuario, os.Getenv("SMTP_CLAVE"), host)

	if err := smtp.SendMail(host+":"+puerto, autenticacion, direccionRemitente, []string{para}, []byte(contenido)); err != nil {
		return NuevoError(http.StatusBadGateway, "No fue posible enviar el correo")
	}

	return nil
}
