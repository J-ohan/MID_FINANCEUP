package models

import (
	"reflect"
	"time"
)

// ConservarFechas copia en "nuevo" las fechas que llegan vacias,
// tomandolas del registro "anterior" que ya esta en la base de datos.
// Asi un PUT que no envia fecha_creacion (u otra fecha) no la borra.
func ConservarFechas(nuevo, anterior interface{}) {
	n := reflect.ValueOf(nuevo).Elem()
	a := reflect.ValueOf(anterior).Elem()
	for i := 0; i < n.NumField(); i++ {
		if t, ok := n.Field(i).Interface().(time.Time); ok && t.IsZero() {
			n.Field(i).Set(a.Field(i))
		}
	}
}
