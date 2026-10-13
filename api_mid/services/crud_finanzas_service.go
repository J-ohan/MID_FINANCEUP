package services

import (
	"api_mid_financeup/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

//ErrNoEncontrado se devuelve cuando el registro pedido no existe
//(o esta eliminado). el controlador lousa para resonder 404.
var ErrNoEncontrado = errors.New("El registro no existe.")

// Sin "limit" los CRUD hechos con bee devuelven solo 10 registros.
const limiteCrud = 10000

type CrudFinanzasService struct {
	URLFinanzas string
	URLUsuarios string
}

func NewCrudFinanzasService() *CrudFinanzasService {
	return &CrudFinanzasService{
		URLFinanzas: web.AppConfig.DefaultString("urlCrudFinanzas", "http://localhost:8082/v1"),
		URLUsuarios: web.AppConfig.DefaultString("urlCrudUsuarios", "http://localhost:8081/v1"),
	}
}