package models

// Ref es una llave foranea tal como llega de los CRUD: {"Id": 1}
// El CRUD manda un objeto con muchos campos vacios; solo
// nos importa el Id.
type Ref struct {
	Id int `json:"Id"`
}
