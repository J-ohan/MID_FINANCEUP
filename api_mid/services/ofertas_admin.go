package services

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// ADMINISTRAR OFERTAS (solo admin: el filtro ya lo revisa)
// Crear, editar y retirar productos de los aliados.
// Retirar no borra: deja el producto con activo = false para que
// las solicitudes viejas sigan apuntando a el.
// ============================================================

// Familias que acepta la tabla negocio.producto_crediticio
var familiasOferta = map[string]bool{"creditos": true, "tarjetas": true, "ahorro": true, "comercios": true}

// AliadosAdmin devuelve los bancos activos para el selector del formulario.
func AliadosAdmin() ([]map[string]interface{}, error) {

	var bancos []models.Banco
	if err := helpers.ConsultarLista(helpers.UrlCrudNegocio+"/banco?query=activo:true&sortby=NombreBanco&order=asc&limit=1000", &bancos); err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}
	for _, banco := range bancos {
		lista = append(lista, map[string]interface{}{"id": banco.Id, "nombre": banco.NombreBanco, "tipo": banco.TipoAliado})
	}

	return lista, nil
}

// CrearOferta guarda un producto nuevo para un aliado.
func CrearOferta(datos models.DatosOfertaAdmin) (map[string]interface{}, error) {

	banco, err := validarOferta(&datos)
	if err != nil {
		return nil, err
	}

	producto := models.ProductoCrediticio{
		Slug:     slugDisponible(banco.NombreBanco + " " + datos.Categoria),
		Usuarios: 0,
		Activo:   true,
	}
	pasarDatos(&producto, datos)

	var creado models.ProductoCrediticio
	if err := helpers.Crear(helpers.UrlCrudNegocio+"/producto_crediticio", producto, &creado); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Oferta creada", "id": creado.Id}, nil
}

// EditarOferta cambia los datos de un producto que ya existe.
func EditarOferta(id int, datos models.DatosOfertaAdmin) (map[string]interface{}, error) {

	producto, err := buscarProducto(id)
	if err != nil {
		return nil, err
	}

	if _, err := validarOferta(&datos); err != nil {
		return nil, err
	}

	pasarDatos(&producto, datos)

	// El CRUD reemplaza el registro: se manda el producto completo
	if err := helpers.Actualizar(fmt.Sprintf("%s/producto_crediticio/%d", helpers.UrlCrudNegocio, id), producto); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Oferta actualizada", "id": id}, nil
}

// RetirarOferta quita el producto de la vitrina (activo = false).
func RetirarOferta(id int) (map[string]interface{}, error) {

	producto, err := buscarProducto(id)
	if err != nil {
		return nil, err
	}

	producto.Activo = false

	if err := helpers.Actualizar(fmt.Sprintf("%s/producto_crediticio/%d", helpers.UrlCrudNegocio, id), producto); err != nil {
		return nil, err
	}

	return map[string]interface{}{"mensaje": "Oferta retirada", "id": id}, nil
}

// buscarProducto trae un producto activo o responde 404.
func buscarProducto(id int) (models.ProductoCrediticio, error) {

	var producto models.ProductoCrediticio
	if err := helpers.ConsultarUno(fmt.Sprintf("%s/producto_crediticio/%d", helpers.UrlCrudNegocio, id), &producto, "La oferta no existe"); err != nil {
		return producto, err
	}

	if !producto.Activo {
		return producto, helpers.NuevoError(http.StatusNotFound, "La oferta no existe")
	}

	return producto, nil
}

// validarOferta limpia los textos, revisa los numeros y que el aliado exista.
func validarOferta(datos *models.DatosOfertaAdmin) (models.Banco, error) {

	banco := models.Banco{}

	datos.Categoria = strings.TrimSpace(datos.Categoria)
	datos.Descripcion = strings.TrimSpace(datos.Descripcion)
	datos.Aprobacion = strings.TrimSpace(datos.Aprobacion)
	datos.Meta = strings.TrimSpace(datos.Meta)
	datos.CtaPrimaria = strings.TrimSpace(datos.CtaPrimaria)
	datos.CtaSecundaria = strings.TrimSpace(datos.CtaSecundaria)
	datos.Familia = strings.ToLower(strings.TrimSpace(datos.Familia))

	if utf8.RuneCountInString(datos.Categoria) < 3 {
		return banco, helpers.NuevoError(http.StatusBadRequest, "El nombre del producto debe tener minimo 3 letras")
	}

	if !familiasOferta[datos.Familia] {
		return banco, helpers.NuevoError(http.StatusBadRequest, "La familia debe ser creditos, tarjetas, ahorro o comercios")
	}

	if datos.TasaDesde < 0 || datos.TasaDesde > 100 {
		return banco, helpers.NuevoError(http.StatusBadRequest, "La tasa debe estar entre 0 y 100")
	}

	if datos.MontoMinimo < 0 || datos.MontoMaximo <= 0 || datos.MontoMinimo > datos.MontoMaximo {
		return banco, helpers.NuevoError(http.StatusBadRequest, "El monto maximo debe ser mayor a 0 y no menor que el minimo")
	}

	// Los productos de ahorro pueden no tener plazo (0 y 0)
	if datos.PlazoMinimo < 0 || datos.PlazoMinimo > datos.PlazoMaximo {
		return banco, helpers.NuevoError(http.StatusBadRequest, "El plazo maximo no puede ser menor que el minimo")
	}

	if datos.CtaPrimaria == "" {
		datos.CtaPrimaria = "Solicitar"
	}

	if err := helpers.ConsultarUno(fmt.Sprintf("%s/banco/%d", helpers.UrlCrudNegocio, datos.IdAliado), &banco, "El aliado no existe"); err != nil {
		return banco, err
	}

	if banco.Id == 0 {
		return banco, helpers.NuevoError(http.StatusNotFound, "El aliado no existe")
	}

	return banco, nil
}

// pasarDatos copia lo que llena el admin en el producto.
func pasarDatos(producto *models.ProductoCrediticio, datos models.DatosOfertaAdmin) {
	producto.IdBanco = &models.Ref{Id: datos.IdAliado}
	producto.NombreProducto = datos.Categoria
	producto.Descripcion = datos.Descripcion
	producto.Familia = datos.Familia
	producto.TasaMinima = datos.TasaDesde
	producto.MontoMinimo = datos.MontoMinimo
	producto.MontoMaximo = datos.MontoMaximo
	producto.PlazoMinimo = datos.PlazoMinimo
	producto.PlazoMaximo = datos.PlazoMaximo
	producto.TiempoAprobacion = datos.Aprobacion
	producto.Vigencia = datos.Meta
	producto.CtaPrimaria = datos.CtaPrimaria
	producto.CtaSecundaria = datos.CtaSecundaria
	producto.Destacada = datos.Destacada

	// La tasa maxima no puede quedar por debajo de la minima
	if producto.TasaMaxima < producto.TasaMinima {
		producto.TasaMaxima = producto.TasaMinima
	}
}

// slugDisponible arma el texto de la URL ("Banco Andino Libre" -> "banco-andino-libre").
// El slug es unico en la tabla: si ya existe se le agrega un numero.
func slugDisponible(texto string) string {

	quitarTildes := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")
	texto = quitarTildes.Replace(strings.ToLower(texto))

	slug := ""
	for _, letra := range texto {
		if (letra >= 'a' && letra <= 'z') || (letra >= '0' && letra <= '9') {
			slug += string(letra)
		} else if !strings.HasSuffix(slug, "-") {
			slug += "-"
		}
	}
	slug = strings.Trim(slug, "-")

	var iguales []models.ProductoCrediticio
	helpers.ConsultarLista(helpers.UrlCrudNegocio+"/producto_crediticio?query=slug:"+slug+"&limit=1", &iguales)

	if len(iguales) > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	return slug
}
