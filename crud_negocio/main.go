package main

import (
	_ "crud_negocio/routers"

	"net/url"
	"time"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

func main() {
	sqlConn, err := beego.AppConfig.String("sqlconn")
	if err != nil {
		panic(err)
	}
	// La contrasena se lee de conexion_bd.conf (archivo que no se sube a git).
	if clave, _ := beego.AppConfig.String("db_password"); clave != "" {
		conexion, err := url.Parse(sqlConn)
		if err != nil {
			panic(err)
		}
		conexion.User = url.UserPassword(conexion.User.Username(), clave)
		sqlConn = conexion.String()
	}
	// Las fechas se manejan en UTC, igual que la base de datos.
	orm.DefaultTimeLoc = time.UTC
	orm.RegisterDataBase("default", "postgres", sqlConn)
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
