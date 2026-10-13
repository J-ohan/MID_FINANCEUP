package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

// ============================================================
// Direcciones de los 5 CRUD
// Tienen un valor por defecto (localhost) y se pueden cambiar
// en el archivo .env sin tocar el codigo.
// ============================================================

var (
	urlCrudAuth      = "http://localhost:8081/v1"
	urlCrudFinanzas  = "http://localhost:8082/v1"
	urlCrudEducacion = "http://localhost:8083/v1"
	urlCrudNegocio   = "http://localhost:8084/v1"
	urlCrudSoporte   = "http://localhost:8085/v1"
)

// CargarUrls lee las direcciones del .env (si existen).
// Se llama una sola vez desde main.go al arrancar.
func CargarUrls() {
	if valor := os.Getenv("URL_CRUD_AUTH"); valor != "" {
		urlCrudAuth = valor
	}
	if valor := os.Getenv("URL_CRUD_FINANZAS"); valor != "" {
		urlCrudFinanzas = valor
	}
	if valor := os.Getenv("URL_CRUD_EDUCACION"); valor != "" {
		urlCrudEducacion = valor
	}
	if valor := os.Getenv("URL_CRUD_NEGOCIO"); valor != "" {
		urlCrudNegocio = valor
	}
	if valor := os.Getenv("URL_CRUD_SOPORTE"); valor != "" {
		urlCrudSoporte = valor
	}
}

// Ref es una llave foranea tal como llega de los CRUD: {"Id": 1}
type Ref struct {
	Id int `json:"Id"`
}

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
		"crud_auth":      urlCrudAuth,
		"crud_finanzas":  urlCrudFinanzas,
		"crud_educacion": urlCrudEducacion,
		"crud_negocio":   urlCrudNegocio,
		"crud_soporte":   urlCrudSoporte,
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

	puerto, _ := beego.AppConfig.String("httpport")
	mid := "http://localhost:" + puerto + "/v1"

	secciones := map[string]string{
		"perfil":              fmt.Sprintf("%s/perfil/%d", mid, id),
		"finanzas":            fmt.Sprintf("%s/resumen-financiero/%d", mid, id),
		"educacion":           fmt.Sprintf("%s/progreso-educativo/%d", mid, id),
		"solicitudes_credito": fmt.Sprintf("%s/solicitud-credito/usuario/%d", mid, id),
		"pqr":                 fmt.Sprintf("%s/pqr/usuario/%d", mid, id),
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
