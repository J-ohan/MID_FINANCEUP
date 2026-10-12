package controllers

import "time"

// ============================================================================
// COMPARTIDOS.GO — Plantillas que usa finanzas y que vienen de crud_auth
// ============================================================================
//
// En la rama "fix" estas dos plantillas están en auth.go (el archivo del
// módulo de usuarios). Como en esta rama no existe auth.go, se dejan aquí.
//
// OJO: si algún día se une esta rama con una que sí tenga auth.go, hay que
// borrar este archivo, porque Go no deja declarar dos veces "Ref" ni "Usuario".
// ============================================================================

// Ref es una llave foránea tal como la manda un CRUD, por ejemplo:
// "IdCategoria": {"Id": 3}
type Ref struct {
	Id int `json:"Id"`
}

// Usuario es la tabla auth.usuario tal como la devuelve crud_auth.
// Finanzas la usa para revisar que el usuario exista y para saber su nombre.
type Usuario struct {
	Id                int       `json:"Id"`
	TipoDocumento     *Ref      `json:"TipoDocumento"`
	Nombre            string    `json:"Nombre"`
	Apellido          string    `json:"Apellido"`
	Email             string    `json:"Email"`
	Telefono          string    `json:"Telefono"`
	Cedula            string    `json:"Cedula"`
	Ciudad            string    `json:"Ciudad"`
	Direccion         string    `json:"Direccion"`
	FechaNacimiento   time.Time `json:"FechaNacimiento"`
	Estado            string    `json:"Estado"`
	FechaRegistro     time.Time `json:"FechaRegistro"`
	FechaUltimaSesion time.Time `json:"FechaUltimaSesion"`
	Activo            bool      `json:"Activo"`
	FechaCreacion     time.Time `json:"FechaCreacion"`
	FechaModificacion time.Time `json:"FechaModificacion"`
}
