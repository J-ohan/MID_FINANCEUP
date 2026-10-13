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

// ErrNoEncontrado se devuelve cuando el registro pedido no existe
// (o está eliminado). El controlador lo usa para responder 404.
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

// ============================================================
// USUARIOS (crud_auth)
// ============================================================

func (s *CrudFinanzasService) ObtenerUsuario(idUsuario int) (models.Usuario, error) {
	var usuario models.Usuario

	url := fmt.Sprintf("%s/usuario/%d", s.URLUsuarios, idUsuario)

	err := consultar(url, &usuario)

	if err != nil {
		return usuario, err
	}

	return usuario, nil
}

// ObtenerNombresUsuarios devuelve id -> "Nombre Apellido" de todos los usuarios.
func (s *CrudFinanzasService) ObtenerNombresUsuarios() (map[int]string, error) {
	var usuarios []models.Usuario

	url := fmt.Sprintf("%s/usuario?limit=%d", s.URLUsuarios, limiteCrud)

	err := consultar(url, &usuarios)

	if err != nil {
		return nil, err
	}

	nombres := map[int]string{}

	for _, usuario := range usuarios {
		nombres[usuario.Id] = NombreCompleto(usuario)
	}

	return nombres, nil
}

// ============================================================
// CONSULTAS (crud_finanzas)
// idUsuario = 0 trae los registros de todos los usuarios (admin)
// ============================================================

func (s *CrudFinanzasService) ObtenerCategorias() ([]models.Categoria, error) {
	var categorias []models.Categoria

	err := consultar(s.URLFinanzas+"/categoria"+filtroActivos(0), &categorias)

	return categorias, err
}

func (s *CrudFinanzasService) ObtenerMovimientos(idUsuario int) ([]models.MovimientoIngresoEgreso, error) {
	var movimientos []models.MovimientoIngresoEgreso

	err := consultar(s.URLFinanzas+"/movimiento_ingreso_egreso"+filtroActivos(idUsuario), &movimientos)

	return movimientos, err
}

func (s *CrudFinanzasService) ObtenerMovimiento(id int) (models.MovimientoIngresoEgreso, error) {
	var movimiento models.MovimientoIngresoEgreso

	err := consultar(fmt.Sprintf("%s/movimiento_ingreso_egreso/%d", s.URLFinanzas, id), &movimiento)

	if err == nil && !movimiento.Activo {
		return movimiento, ErrNoEncontrado
	}

	return movimiento, err
}

func (s *CrudFinanzasService) ObtenerMetas(idUsuario int) ([]models.Meta, error) {
	var metas []models.Meta

	err := consultar(s.URLFinanzas+"/meta"+filtroActivos(idUsuario), &metas)

	return metas, err
}

func (s *CrudFinanzasService) ObtenerMeta(id int) (models.Meta, error) {
	var meta models.Meta

	err := consultar(fmt.Sprintf("%s/meta/%d", s.URLFinanzas, id), &meta)

	if err == nil && !meta.Activo {
		return meta, ErrNoEncontrado
	}

	return meta, err
}

func (s *CrudFinanzasService) ObtenerInversiones(idUsuario int) ([]models.Inversion, error) {
	var inversiones []models.Inversion

	err := consultar(s.URLFinanzas+"/inversion"+filtroActivos(idUsuario), &inversiones)

	return inversiones, err
}

func (s *CrudFinanzasService) ObtenerInversion(id int) (models.Inversion, error) {
	var inversion models.Inversion

	err := consultar(fmt.Sprintf("%s/inversion/%d", s.URLFinanzas, id), &inversion)

	if err == nil && !inversion.Activo {
		return inversion, ErrNoEncontrado
	}

	return inversion, err
}

func (s *CrudFinanzasService) ObtenerNivelesRiesgo() ([]models.NivelRiesgo, error) {
	var niveles []models.NivelRiesgo

	err := consultar(s.URLFinanzas+"/nivel_riesgo"+filtroActivos(0), &niveles)

	return niveles, err
}

// ObtenerHistorial no devuelve error: si falla (por ejemplo, si crud_finanzas
// todavía no tiene la tabla historial_inversion), la gráfica sale vacía.
func (s *CrudFinanzasService) ObtenerHistorial() []models.HistorialInversion {
	var historial []models.HistorialInversion

	err := consultar(s.URLFinanzas+"/historial_inversion"+filtroActivos(0), &historial)

	if err != nil {
		return []models.HistorialInversion{}
	}

	return historial
}

// ============================================================
// GUARDAR, ACTUALIZAR Y BORRAR (crud_finanzas)
// ============================================================

// Crear guarda un registro nuevo. "registro" debe pasarse con &
// para que quede con el Id que le asignó la base de datos.
func (s *CrudFinanzasService) Crear(tabla string, registro interface{}) error {
	return enviar(http.MethodPost, s.URLFinanzas+"/"+tabla, registro)
}

// Actualizar reemplaza el registro completo.
func (s *CrudFinanzasService) Actualizar(tabla string, id int, registro interface{}) error {
	return enviar(http.MethodPut, fmt.Sprintf("%s/%s/%d", s.URLFinanzas, tabla, id), registro)
}

// Borrar elimina el registro de verdad. Solo se usa para deshacer
// una importación de Excel que falló a la mitad.
func (s *CrudFinanzasService) Borrar(tabla string, id int) error {
	return enviar(http.MethodDelete, fmt.Sprintf("%s/%s/%d", s.URLFinanzas, tabla, id), nil)
}

// ============================================================
// FUNCIONES INTERNAS
// ============================================================

// filtroActivos arma el "?query=" para pedir solo los registros activos.
func filtroActivos(idUsuario int) string {
	if idUsuario > 0 {
		return fmt.Sprintf("?query=id_usuario:%d,activo:true&limit=%d", idUsuario, limiteCrud)
	}

	return fmt.Sprintf("?query=activo:true&limit=%d", limiteCrud)
}

// consultar hace un GET y guarda el JSON de la respuesta en "destino".
func consultar(url string, destino interface{}) error {
	respuesta, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("No fue posible comunicarse con el CRUD: %w", err)
	}

	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		return fmt.Errorf("El CRUD retorna con código %d", respuesta.StatusCode)
	}

	body, _ := io.ReadAll(respuesta.Body)

	err = json.Unmarshal(body, destino)

	if err != nil {
		// Los CRUD de bee responden los errores como texto y con código 200
		if strings.Contains(string(body), "no row found") {
			return ErrNoEncontrado
		}

		return fmt.Errorf("Error al convertir la respuesta del CRUD: %s", string(body))
	}

	return nil
}

// enviar hace un POST, PUT o DELETE con "registro" en JSON.
// POST: el CRUD responde 201 con el registro creado (se copia en "registro").
// PUT y DELETE: el CRUD responde "OK".
func enviar(metodo string, url string, registro interface{}) error {
	var cuerpo io.Reader

	if registro != nil {
		jsonRegistro, _ := json.Marshal(registro)
		cuerpo = bytes.NewBuffer(jsonRegistro)
	}

	peticion, _ := http.NewRequest(metodo, url, cuerpo)
	peticion.Header.Set("Content-Type", "application/json")

	respuesta, err := http.DefaultClient.Do(peticion)

	if err != nil {
		return fmt.Errorf("No fue posible comunicarse con el CRUD: %w", err)
	}

	defer respuesta.Body.Close()

	body, _ := io.ReadAll(respuesta.Body)

	if metodo == http.MethodPost {
		if respuesta.StatusCode != http.StatusCreated {
			return fmt.Errorf("El CRUD no guardó el registro: %s", string(body))
		}

		json.Unmarshal(body, registro)

		return nil
	}

	if strings.TrimSpace(string(body)) != `"OK"` {
		return fmt.Errorf("El CRUD no actualizó el registro: %s", string(body))
	}

	return nil
}
