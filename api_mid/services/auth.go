package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api_mid/helpers"
	"api_mid/models"
)

// BuscarUsuario trae un usuario por su id (lo usan otros modulos).
func BuscarUsuario(id int) (models.Usuario, error) {
	var usuario models.Usuario
	err := helpers.ConsultarUno(fmt.Sprintf("%s/usuario/%d", helpers.UrlCrudAuth, id), &usuario, "El usuario no existe")
	return usuario, err
}

// buscarUsuarioPorEmail devuelve el usuario y si existe o no.
func buscarUsuarioPorEmail(email string) (models.Usuario, bool, error) {

	var usuarios []models.Usuario

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?query=email:"+url.QueryEscape(email), &usuarios); err != nil {
		return models.Usuario{}, false, err
	}

	if len(usuarios) == 0 {
		return models.Usuario{}, false, nil
	}

	return usuarios[0], true, nil
}

// NombresUsuarios devuelve un mapa id -> "Nombre Apellido" de todos los usuarios.
func NombresUsuarios() (map[int]string, error) {

	var usuarios []models.Usuario

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/usuario?limit=10000", &usuarios); err != nil {
		return nil, err
	}

	nombres := map[int]string{}
	for _, usuario := range usuarios {
		nombres[usuario.Id] = usuario.Nombre + " " + usuario.Apellido
	}

	return nombres, nil
}

// asignarRolUsuario le da el rol "Usuario" a un usuario nuevo

func asignarRolUsuario(idUsuario int) error {

	var roles []models.Rol

	if err := helpers.ConsultarLista(helpers.UrlCrudAuth+"/rol?query=nombre_rol:Usuario", &roles); err != nil {
		return err
	}

	if len(roles) == 0 {
		return helpers.NuevoError(http.StatusInternalServerError, "No existe el rol Usuario en la base de datos")
	}

	asignacion := models.UsuarioRol{IdUsuario: &models.Ref{Id: idUsuario}, IdRol: &models.Ref{Id: roles[0].Id}, Activo: true}

	return helpers.Crear(helpers.UrlCrudAuth+"/usuario_rol", asignacion, nil)
}

// Por ahora el envio es mostrarlo en la consola del MID.
func crearCodigo(idUsuario int, email string, proposito string, minutos int) error {

	codigo := helpers.NumeroAleatorio()

	nuevoCodigo := models.CodigoVerificacion{
		IdUsuario:       &models.Ref{Id: idUsuario},
		Proposito:       proposito,
		Canal:           "email",
		CodigoHash:      helpers.Cifrar(codigo),
		FechaExpiracion: time.Now().UTC().Add(time.Duration(minutos) * time.Minute),
		Activo:          true,
	}

	if err := helpers.Crear(helpers.UrlCrudAuth+"/codigo_verificacion", nuevoCodigo, nil); err != nil {
		return err
	}
	
// Cuando haya servicio de correo, aqui se envia al email
	fmt.Println("[CODIGO "+proposito+"]", email, codigo)

	return nil
}


// Si es correcto lo desactiva (no sirve dos veces); si no, suma un intento.
func validarCodigo(idUsuario int, proposito string, codigoEscrito string) error {

	var codigos []models.CodigoVerificacion

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/codigo_verificacion?query=id_usuario:%d,proposito:%s,activo:true&sortby=Id&order=desc&limit=1", helpers.UrlCrudAuth, idUsuario, proposito), &codigos); err != nil {
		return err
	}

	if len(codigos) == 0 {
		return helpers.NuevoError(http.StatusUnauthorized, "No hay un codigo pendiente. Pide uno nuevo.")
	}

	codigo := codigos[0]
	ahora := time.Now().UTC()

	if codigo.FechaExpiracion.Before(ahora) || codigo.Intentos >= 5 {
		return helpers.NuevoError(http.StatusUnauthorized, "El codigo vencio. Pide uno nuevo.")
	}

	correcto := helpers.CoincideCifrado(codigo.CodigoHash, codigoEscrito)

	if correcto {
		codigo.Activo = false
		codigo.FechaUso = ahora
	} else {
		codigo.Intentos++
	}

	helpers.Actualizar(fmt.Sprintf("%s/codigo_verificacion/%d", helpers.UrlCrudAuth, codigo.Id), codigo)

	if !correcto {
		return helpers.NuevoError(http.StatusUnauthorized, "El codigo no es correcto")
	}

	return nil
}

