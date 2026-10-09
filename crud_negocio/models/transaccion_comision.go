package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type TransaccionComision struct {
	Id                 int                  `orm:"column(id_transaccion);pk"`
	IdCredito          *CreditoDesembolsado `orm:"column(id_credito);rel(fk)"`
	IdBanco            *Banco               `orm:"column(id_banco);rel(fk)"`
	MontoComision      float64              `orm:"column(monto_comision)"`
	PorcentajeAplicado float64              `orm:"column(porcentaje_aplicado);null"`
	FechaTransaccion   time.Time            `orm:"column(fecha_transaccion);type(timestamp without time zone);null;auto_now_add"`
	Estado             string               `orm:"column(estado)"`
	ReferenciaPago     string               `orm:"column(referencia_pago);null"`
	Activo             bool                 `orm:"column(activo)"`
	FechaCreacion      time.Time            `orm:"column(fecha_creacion);type(timestamp without time zone);auto_now_add"`
	FechaModificacion  time.Time            `orm:"column(fecha_modificacion);type(timestamp without time zone);auto_now_add"`
}

func (t *TransaccionComision) TableName() string {
	return "transaccion_comision"
}

func init() {
	orm.RegisterModel(new(TransaccionComision))
}

// AddTransaccionComision insert a new TransaccionComision into database and returns
// last inserted Id on success.
func AddTransaccionComision(m *TransaccionComision) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetTransaccionComisionById retrieves TransaccionComision by Id. Returns error if
// Id doesn't exist
func GetTransaccionComisionById(id int) (v *TransaccionComision, err error) {
	o := orm.NewOrm()
	v = &TransaccionComision{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllTransaccionComision retrieves all TransaccionComision matches certain condition. Returns empty list if
// no records exist
func GetAllTransaccionComision(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(TransaccionComision))
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

	var l []TransaccionComision
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

// UpdateTransaccionComision updates TransaccionComision by Id and returns error if
// the record to be updated doesn't exist
func UpdateTransaccionComisionById(m *TransaccionComision) (err error) {
	o := orm.NewOrm()
	v := TransaccionComision{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteTransaccionComision deletes TransaccionComision by Id and returns error if
// the record to be deleted doesn't exist
func DeleteTransaccionComision(id int) (err error) {
	o := orm.NewOrm()
	v := TransaccionComision{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&TransaccionComision{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
