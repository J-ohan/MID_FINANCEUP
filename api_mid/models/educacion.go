package models

// ------------------------------------------------------------
// Tablas de crud_educacion (como llegan del CRUD)
// ------------------------------------------------------------

type ModuloEducativo struct {
	Id     int    `json:"Id"`
	Titulo string `json:"Titulo"`
	Nivel  string `json:"Nivel"`
}

type ProgresoEducativo struct {
	IdModulo             *Ref `json:"IdModulo"`
	PorcentajeCompletado int  `json:"PorcentajeCompletado"`
	Calificacion         int  `json:"Calificacion"`
}
