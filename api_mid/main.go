package main

import (
	"api_mid_financeup/helpers"
	_ "api_mid_financeup/routers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Las direcciones de los CRUD salen del archivo .env
	// (en la carpeta MID_FINANCEUP, no se sube a git).
	godotenv.Load("../.env")
	helpers.CargarUrls()

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// Permite que el frontend (Angular) llame al MID.
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	beego.Run()
}
