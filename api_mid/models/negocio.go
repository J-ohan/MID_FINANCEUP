package models

// ------------------------------------------------------------
// Tablas de crud_negocio (como llegan del CRUD)
// ------------------------------------------------------------

type Banco struct {
	Id          int    `json:"Id"`
	NombreBanco string `json:"NombreBanco"`
	UrlLogo     string `json:"UrlLogo"`
}

type ProductoCrediticio struct {
	Id             int     `json:"Id"`
	IdBanco        *Ref    `json:"IdBanco"`
	NombreProducto string  `json:"NombreProducto"`
	Descripcion    string  `json:"Descripcion"`
	MontoMinimo    float64 `json:"MontoMinimo"`
	MontoMaximo    float64 `json:"MontoMaximo"`
	TasaMinima     float64 `json:"TasaMinima"`
	TasaMaxima     float64 `json:"TasaMaxima"`
	PlazoMinimo    int     `json:"PlazoMinimo"`
	PlazoMaximo    int     `json:"PlazoMaximo"`
}
