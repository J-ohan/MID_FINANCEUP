package helpers

// ErrorApi es un error que ademas trae el codigo HTTP que se
// le debe responder a la pagina (400, 401, 404, 502...).
//
// Los services devuelven este error y el controlador solo lo
// pasa a ResponderError.
type ErrorApi struct {
	Codigo  int
	Mensaje string
}

// Error hace que ErrorApi se pueda usar como un error normal de Go.
func (e *ErrorApi) Error() string {
	return e.Mensaje
}

// NuevoError crea un ErrorApi.
// Ejemplo: helpers.NuevoError(404, "El usuario no existe")
func NuevoError(codigo int, mensaje string) *ErrorApi {
	return &ErrorApi{Codigo: codigo, Mensaje: mensaje}
}
