package main

import (
	_ "api_mid_financeup/routers" // Carga las rutas (routers/router.go)

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func main() {
	// Permite que el frontend (Angular) llame al MID desde otro puerto,
	// por ejemplo http://localhost:4200. Sin esto el navegador bloquea las peticiones.
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
