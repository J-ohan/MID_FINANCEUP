package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type CreditoDesembolsado struct {
	Id                int                 `orm:"column(id_credito);pk"`
	IdLead            *Lead               `orm:"column(id_lead);rel(fk)"`
	IdUsuario         *Usuario            `orm:"column(id_usuario);rel(fk)"`
	IdProducto        *ProductoCrediticio `orm:"column(id_producto);rel(fk)"`
	IdBanco           *Banco              `orm:"column(id_banco);rel(fk)"`
	NumeroCredito     string              `orm:"column(numero_credito);null"`
	MontoAprobado     float64             `orm:"column(monto_aprobado)"`
	TasaInteresFinal  float64             `orm:"column(tasa_interes_final);null"`
	PlazoMeses        int                 `orm:"column(plazo_meses);null"`
	FechaAprobacion   time.Time           `orm:"column(fecha_aprobacion);type(date);null"`
	FechaDesembolso   time.Time           `orm:"column(fecha_desembolso);type(date);null"`
	EstadoCredito     string              `orm:"column(estado_credito)"`
	SaldoActual       float64             `orm:"column(saldo_actual);null"`
	Activo            bool                `orm:"column(activo)"`
	FechaCreacion     time.Time           `orm:"column(fecha_creacion);type(timestamp without time zone);auto_now_add"`
	FechaModificacion time.Time           `orm:"column(fecha_modificacion);type(timestamp without time zone);auto_now_add"`
}

func (t *CreditoDesembolsado) TableName() string {
	return "credito_desembolsado"
}

func init() {
	orm.RegisterModel(new(CreditoDesembolsado))
}

// AddCreditoDesembolsado insert a new CreditoDesembolsado into database and returns
// last inserted Id on success.
func AddCreditoDesembolsado(m *CreditoDesembolsado) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetCreditoDesembolsadoById retrieves CreditoDesembolsado by Id. Returns error if
// Id doesn't exist
func GetCreditoDesembolsadoById(id int) (v *CreditoDesembolsado, err error) {
	o := orm.NewOrm()
	v = &CreditoDesembolsado{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllCreditoDesembolsado retrieves all CreditoDesembolsado matches certain condition. Returns empty list if
// no records exist
func GetAllCreditoDesembolsado(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(CreditoDesembolsado))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		if strings.Contains(k, "isnull") {
			qs = qs.Filter(k, (v == "true" || v == "1"))
		} else {
			qs = qs.Filter(k, v)
		}
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []CreditoDesembolsado
	qs = qs.OrderBy(sortFields...)
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateCreditoDesembolsado updates CreditoDesembolsado by Id and returns error if
// the record to be updated doesn't exist
func UpdateCreditoDesembolsadoById(m *CreditoDesembolsado) (err error) {
	o := orm.NewOrm()
	v := CreditoDesembolsado{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteCreditoDesembolsado deletes CreditoDesembolsado by Id and returns error if
// the record to be deleted doesn't exist
func DeleteCreditoDesembolsado(id int) (err error) {
	o := orm.NewOrm()
	v := CreditoDesembolsado{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&CreditoDesembolsado{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
