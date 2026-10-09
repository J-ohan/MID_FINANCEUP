package main

import (
	_ "crud_auth/routers"

	"net/url"
	"os"
	"time"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	sqlConn, err := beego.AppConfig.String("sqlconn")
	if err != nil {
		panic(err)
	}
	// El usuario y la contrasena de la base salen del archivo .env
	// (en la carpeta MID_FINANCEUP, no se sube a git).
	godotenv.Load("../.env")
	conexion, err := url.Parse(sqlConn)
	if err != nil {
		panic(err)
	}
	usuario := conexion.User.Username()
	if valor := os.Getenv("DB_USER"); valor != "" {
		usuario = valor
	}
	conexion.User = url.UserPassword(usuario, os.Getenv("DB_PASSWORD"))
	sqlConn = conexion.String()
	// Las fechas se manejan en UTC, igual que la base de datos.
	orm.DefaultTimeLoc = time.UTC
	orm.RegisterDataBase("default", "postgres", sqlConn)
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
