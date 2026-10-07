package controllers

import (
	"api_mid_financeup/models"
	"api_mid_financeup/services"
)

// FinanzasController GET /v1/resumen-financiero/:id (usa crud_auth + crud_finanzas)
type FinanzasController struct {
	BaseController
}

func (c *FinanzasController) ResumenFinanciero() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	resumen, err := services.ResumenFinanciero(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, resumen)
}

// EducacionController GET /v1/progreso-educativo/:id (usa crud_auth + crud_educacion)
type EducacionController struct {
	BaseController
}

func (c *EducacionController) ProgresoEducativo() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	progreso, err := services.ProgresoEducativo(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, progreso)
}

// NegocioController atiende ofertas y solicitudes de credito (usa crud_auth + crud_negocio + crud_soporte).
type NegocioController struct {
	BaseController
}

// Ofertas GET /v1/ofertas
func (c *NegocioController) Ofertas() {
	ofertas, err := services.Ofertas()
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, ofertas)
}

// CrearSolicitud POST /v1/solicitud-credito
func (c *NegocioController) CrearSolicitud() {
	var datos models.DatosSolicitudCredito
	if err := c.leerCuerpo(&datos); err != nil {
		c.fallo(err)
		return
	}
	solicitud, err := services.CrearSolicitudCredito(datos)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(201, solicitud)
}

// SolicitudesDeUsuario GET /v1/solicitud-credito/usuario/:id
func (c *NegocioController) SolicitudesDeUsuario() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	solicitudes, err := services.SolicitudesDeUsuario(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, solicitudes)
}

// SoporteController atiende las PQR (usa crud_auth + crud_soporte).
type SoporteController struct {
	BaseController
}

// PqrDeUsuario GET /v1/pqr/usuario/:id
func (c *SoporteController) PqrDeUsuario() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	pqrs, err := services.PqrDeUsuario(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, pqrs)
}

// CrearPqr POST /v1/pqr
func (c *SoporteController) CrearPqr() {
	var datos models.DatosPqr
	if err := c.leerCuerpo(&datos); err != nil {
		c.fallo(err)
		return
	}
	pqr, err := services.CrearPqr(datos)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(201, pqr)
}

// DashboardController GET /v1/dashboard/:id (usa los 5 CRUD) y GET /v1/estado
type DashboardController struct {
	BaseController
}

func (c *DashboardController) Dashboard() {
	id, err := c.idDeLaRuta("id")
	if err != nil {
		c.fallo(err)
		return
	}
	panel, err := services.Dashboard(id)
	if err != nil {
		c.fallo(err)
		return
	}
	c.ok(200, panel)
}

func (c *DashboardController) Estado() {
	c.ok(200, services.EstadoServicios())
}
