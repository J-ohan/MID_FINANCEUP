package services

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"

)

func RegistrarActividad (IdUsuario int, tipo string, descripcion string, entidad string) {
	actividad := models.RegistroActividad{
		IdUsuario: &idUsuario,
		TipoActividad: tipo,
		Descripcion: descripcion,
		EntidadAfectada: entidad,
	}
	helpers.Crear(helpers.urlCrudSoporte+"/registro_actividad", actividad, nil)
}

func estadosPorId() (map[int]string, error){
	var estados []models.EstadoPqr

	if err := helpers.ConsultarLista(helpers.urlCrudSoporte+"/estado_pqr?limit=1000", &estados); err !=nil{
		return nil, err
	}
}