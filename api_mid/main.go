package main

import (
	_ "api_mid_financeup/routers" // Carga las rutas (routers/router.go)

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
 
	// Permite que el frontend (Angular) llame al MID.
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"PUT",
			"PATCH",
			"GET",
			"POST",
			"OPTIONS",
			"DELETE",
		},
		AllowHeaders: []string{
			"Origin",
			"x-requested-with",
			"content-type",
			"accept",
			"origin",
			"authorization",
			"x-csrftoken",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Arranca el servidor en el puerto de conf/app.conf (8080)
	beego.Run()
}
