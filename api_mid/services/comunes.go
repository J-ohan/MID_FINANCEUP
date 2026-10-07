package services

import (
	"errors"
	"math"
	"time"

	"api_mid_financeup/helpers"
)

// mensajeNoEncontrado cambia el mensaje generico de "no encontrado" por uno
// mas claro para el usuario. Los demas errores pasan sin cambios.
func mensajeNoEncontrado(err error, mensaje string) error {
	var errorMid *helpers.ErrorMid
	if errors.As(err, &errorMid) && errorMid.Status == 404 {
		return helpers.NuevoError(404, mensaje)
	}
	return err
}

// fecha convierte una fecha a texto AAAA-MM-DD (vacio si no tiene fecha).
func fecha(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// redondear deja un numero con 2 decimales.
func redondear(valor float64) float64 {
	return math.Round(valor*100) / 100
}
