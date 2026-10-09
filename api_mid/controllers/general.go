package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

// GeneralController operations for General
type GeneralController struct {
	beego.Controller
}

// ============================================================
// GET /v1/estado
// Dice que CRUD estan encendidos
// ============================================================

func (c *GeneralController) Estado() {

	// --------------------------------------------------------
	// 1. Direcciones de los 5 CRUD
	// --------------------------------------------------------

	cruds := map[string]string{
		"crud_auth":      "http://localhost:8081/v1",
		"crud_finanzas":  "http://localhost:8082/v1",
		"crud_educacion": "http://localhost:8083/v1",
		"crud_negocio":   "http://localhost:8084/v1",
		"crud_soporte":   "http://localhost:8085/v1",
	}

	// --------------------------------------------------------
	// 2. Intentar llamar a cada uno
	// --------------------------------------------------------

	estado := map[string]string{"mid": "encendido"}

	for nombre, direccion := range cruds {

		response, err := http.Get(direccion)

		if err != nil {
			estado[nombre] = "apagado"
			continue
		}

		response.Body.Close()
		estado[nombre] = "encendido"
	}

	// --------------------------------------------------------
	// 3. Devolver el JSON
	// --------------------------------------------------------

	c.Data["json"] = estado
	c.ServeJSON()
}

// ============================================================
// GET /v1/dashboard/:id
// Junta en una sola respuesta las rutas del MID:
// perfil, resumen financiero, progreso, solicitudes y PQR
// ============================================================

func (c *GeneralController) GetDashboard() {

	// --------------------------------------------------------
	// 1. Obtener el ID enviado en la URL
	// --------------------------------------------------------

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))

	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"error": "El ID debe ser numerico"}
		c.ServeJSON()
		return
	}

	// --------------------------------------------------------
	// 2. Rutas del MID que se van a juntar
	// --------------------------------------------------------

	secciones := map[string]string{
		"perfil":              fmt.Sprintf("http://localhost:8080/v1/perfil/%d", id),
		"finanzas":            fmt.Sprintf("http://localhost:8080/v1/resumen-financiero/%d", id),
		"educacion":           fmt.Sprintf("http://localhost:8080/v1/progreso-educativo/%d", id),
		"solicitudes_credito": fmt.Sprintf("http://localhost:8080/v1/solicitud-credito/usuario/%d", id),
		"pqr":                 fmt.Sprintf("http://localhost:8080/v1/pqr/usuario/%d", id),
	}

	// --------------------------------------------------------
	// 3. Llamar a cada ruta y guardar su respuesta
	//    Si una falla (por ejemplo, un CRUD apagado),
	//    se agrega un aviso y se sigue con las demas.
	// --------------------------------------------------------

	dashboard := map[string]interface{}{}
	avisos := []string{}

	for nombre, direccion := range secciones {

		response, err := http.Get(direccion)

		if err != nil {
			avisos = append(avisos, nombre+": no respondio")
			continue
		}

		body, _ := io.ReadAll(response.Body)
		response.Body.Close()

		if response.StatusCode != http.StatusOK {
			avisos = append(avisos, nombre+": "+string(body))
			continue
		}

		var datos interface{}
		json.Unmarshal(body, &datos)
		dashboard[nombre] = datos
	}

	// --------------------------------------------------------
	// 4. Si no hay perfil, el usuario no existe
	// --------------------------------------------------------

	if dashboard["perfil"] == nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"error": "El usuario no existe", "avisos": avisos}
		c.ServeJSON()
		return
	}

	if len(avisos) > 0 {
		dashboard["avisos"] = avisos
	}

	// --------------------------------------------------------
	// 5. Devolver el JSON
	// --------------------------------------------------------

	c.Data["json"] = dashboard
	c.ServeJSON()
}
