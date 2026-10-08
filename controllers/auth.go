 package controllers

import (
	"strings"

	"api-usuarios/models"
	"api-usuarios/utils"

	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	web.Controller
}

type RegistroRequest struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Correo   string `json:"correo"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Correo   string `json:"correo"`
	Password string `json:"password"`
}