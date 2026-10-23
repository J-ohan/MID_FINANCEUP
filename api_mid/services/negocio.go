package services

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// ============================================================
// SERVICE DE ALIANZAS (crud_negocio)
// Aqui esta la logica: unir ofertas con su aliado y detalles,
// validar la solicitud y calcular la cuota.
// ============================================================

// datosOferta es como la pagina muestra una oferta.
func datosOferta(producto models.ProductoCrediticio, banco models.Banco, textos map[string][]string) map[string]interface{} {
	return map[string]interface{}{
		"id":             producto.Id,
		"slug":           producto.Slug,
		"aliado":         banco.NombreBanco,
		"tipo":           banco.TipoAliado,
		"logo_texto":     banco.LogoTexto,
		"logo_fondo":     banco.LogoFondo,
		"categoria":      producto.NombreProducto,
		"descripcion":    producto.Descripcion,
		"promesa":        producto.Promesa,
		"meta":           producto.Vigencia,
		"familia":        producto.Familia,
		"beneficio":      producto.Beneficio,
		"tasa_desde":     producto.TasaMinima,
		"monto_maximo":   producto.MontoMaximo,
		"plazo_maximo":   producto.PlazoMaximo,
		"aprobacion":     producto.TiempoAprobacion,
		"calificacion":   producto.Calificacion,
		"usuarios":       producto.Usuarios,
		"compatibilidad": producto.Compatibilidad,
		"destacada":      producto.Destacada,
		"cta_primaria":   producto.CtaPrimaria,
		"cta_secundaria": producto.CtaSecundaria,
		"etiquetas":      textos["etiqueta"],
		"perfiles":       textos["perfil"],
		"requisitos":     textos["requisito"],
	}
}

// Ofertas: productos + su aliado + etiquetas, perfiles y requisitos.
func Ofertas() ([]map[string]interface{}, error) {

	// 1. Productos activos, bancos y detalles
	var productos []models.ProductoCrediticio
	var bancos []models.Banco
	var detalles []models.ProductoDetalle

	if err := helpers.ConsultarLista(helpers.UrlCrudNegocio+"/producto_crediticio?query=activo:true&limit=1000", &productos); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudNegocio+"/banco?limit=1000", &bancos); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(helpers.UrlCrudNegocio+"/producto_detalle?query=activo:true&sortby=Orden&order=asc&limit=10000", &detalles); err != nil {
		return nil, err
	}

	// 2. Guardar bancos por id y detalles por producto y tipo
	bancoPorId := map[int]models.Banco{}
	for _, banco := range bancos {
		bancoPorId[banco.Id] = banco
	}

	// textos[idProducto][tipo] = lista de textos
	textos := map[int]map[string][]string{}

	for _, detalle := range detalles {
		idProducto := detalle.IdProducto.Id
		if textos[idProducto] == nil {
			textos[idProducto] = map[string][]string{}
		}
		textos[idProducto][detalle.Tipo] = append(textos[idProducto][detalle.Tipo], detalle.Texto)
	}

	// 3. Unir todo
	ofertas := []map[string]interface{}{}

	for _, producto := range productos {
		ofertas = append(ofertas, datosOferta(producto, bancoPorId[producto.IdBanco.Id], textos[producto.Id]))
	}

	return ofertas, nil
}

// FichaOferta: oferta completa con caracteristicas, documentos, tarifas,
// preguntas y otras ofertas de la misma familia.
func FichaOferta(id int) (map[string]interface{}, error) {

	// 1. El producto y su banco
	var producto models.ProductoCrediticio

	if err := helpers.ConsultarUno(fmt.Sprintf("%s/producto_crediticio/%d", helpers.UrlCrudNegocio, id), &producto, "La oferta no existe"); err != nil {
		return nil, err
	}

	if !producto.Activo {
		return nil, helpers.NuevoError(http.StatusNotFound, "La oferta no existe")
	}

	banco := models.Banco{}
	helpers.ConsultarUno(fmt.Sprintf("%s/banco/%d", helpers.UrlCrudNegocio, producto.IdBanco.Id), &banco, "")

	// 2. Detalles separados por tipo
	var detalles []models.ProductoDetalle

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/producto_detalle?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", helpers.UrlCrudNegocio, id), &detalles); err != nil {
		return nil, err
	}

	textos := map[string][]string{}
	for _, detalle := range detalles {
		textos[detalle.Tipo] = append(textos[detalle.Tipo], detalle.Texto)
	}

	// 3. Tarifas y preguntas
	var tarifas []models.ProductoTarifa
	var preguntas []models.ProductoPregunta

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/producto_tarifa?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", helpers.UrlCrudNegocio, id), &tarifas); err != nil {
		return nil, err
	}

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/producto_pregunta?query=id_producto:%d,activo:true&sortby=Orden&order=asc&limit=1000", helpers.UrlCrudNegocio, id), &preguntas); err != nil {
		return nil, err
	}

	listaTarifas := []map[string]interface{}{}
	for _, tarifa := range tarifas {
		listaTarifas = append(listaTarifas, map[string]interface{}{"concepto": tarifa.Concepto, "valor": tarifa.Valor})
	}

	listaPreguntas := []map[string]interface{}{}
	for _, pregunta := range preguntas {
		listaPreguntas = append(listaPreguntas, map[string]interface{}{"pregunta": pregunta.Pregunta, "respuesta": pregunta.Respuesta})
	}

	// 4. Otras ofertas de la misma familia
	var familia []models.ProductoCrediticio
	helpers.ConsultarLista(helpers.UrlCrudNegocio+"/producto_crediticio?query=activo:true,familia:"+producto.Familia+"&limit=1000", &familia)

	relacionadas := []map[string]interface{}{}
	for _, otro := range familia {
		if otro.Id != producto.Id {
			relacionadas = append(relacionadas, map[string]interface{}{"id": otro.Id, "categoria": otro.NombreProducto, "tasa_desde": otro.TasaMinima})
		}
	}

	// 5. Armar la ficha
	ficha := datosOferta(producto, banco, textos)
	ficha["monto_minimo"] = producto.MontoMinimo
	ficha["plazo_minimo"] = producto.PlazoMinimo
	ficha["caracteristicas"] = textos["caracteristica"]
	ficha["documentos"] = textos["documento"]
	ficha["tarifas"] = listaTarifas
	ficha["preguntas"] = listaPreguntas
	ficha["relacionadas"] = relacionadas

	return ficha, nil
}

// RadicarSolicitud: valida y guarda la solicitud de credito (lead)
// con radicado SOL-xxxxxx y calcula la cuota estimada.
func RadicarSolicitud(datos models.DatosSolicitudCredito) (map[string]interface{}, error) {

	// 1. Validar los datos del solicitante
	if datos.Nombre == "" || datos.Documento == "" || datos.Celular == "" || !strings.Contains(datos.Correo, "@") {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Faltan datos: nombre, documento, correo y celular son obligatorios")
	}

	if datos.Ingresos <= 0 {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Los ingresos mensuales deben ser mayores a 0")
	}

	if !datos.AceptaTerminos {
		return nil, helpers.NuevoError(http.StatusBadRequest, "Debes aceptar los terminos y la autorizacion de datos")
	}

	// 2. Revisar usuario y producto
	if _, err := BuscarUsuario(datos.IdUsuario); err != nil {
		return nil, err
	}

	var producto models.ProductoCrediticio

	if err := helpers.ConsultarUno(fmt.Sprintf("%s/producto_crediticio/%d", helpers.UrlCrudNegocio, datos.IdProducto), &producto, "El producto de credito no existe"); err != nil {
		return nil, err
	}

	if !producto.Activo {
		return nil, helpers.NuevoError(http.StatusNotFound, "El producto de credito no existe")
	}

	// 3. Revisar que monto y plazo esten en el rango del producto
	if datos.Monto < producto.MontoMinimo || datos.Monto > producto.MontoMaximo {
		return nil, helpers.NuevoError(http.StatusBadRequest, fmt.Sprintf("El monto debe estar entre %.0f y %.0f", producto.MontoMinimo, producto.MontoMaximo))
	}

	if datos.PlazoMeses < producto.PlazoMinimo || datos.PlazoMeses > producto.PlazoMaximo {
		return nil, helpers.NuevoError(http.StatusBadRequest, fmt.Sprintf("El plazo debe estar entre %d y %d meses", producto.PlazoMinimo, producto.PlazoMaximo))
	}

	// 4. Buscar un asesor del banco del producto
	var asesores []models.AsesorBancario
	helpers.ConsultarLista(fmt.Sprintf("%s/asesor_bancario?query=id_banco:%d,activo:true", helpers.UrlCrudNegocio, producto.IdBanco.Id), &asesores)

	// 5. Guardar la solicitud con su radicado
	solicitud := models.Lead{
		IdUsuario:            datos.IdUsuario,
		IdProducto:           &models.Ref{Id: producto.Id},
		Radicado:             "SOL-" + helpers.NumeroAleatorio(),
		TipoCredito:          producto.NombreProducto,
		MontoInteres:         datos.Monto,
		PlazoInteres:         datos.PlazoMeses,
		NombreSolicitante:    datos.Nombre,
		DocumentoSolicitante: datos.Documento,
		CorreoSolicitante:    datos.Correo,
		CelularSolicitante:   datos.Celular,
		IngresosMensuales:    datos.Ingresos,
		AceptaTerminos:       true,
		EstadoLead:           "nuevo",
		Activo:               true,
	}

	nombreAsesor := ""
	if len(asesores) > 0 {
		solicitud.IdAsesor = &models.Ref{Id: asesores[0].Id}
		nombreAsesor = asesores[0].Nombre + " " + asesores[0].Apellido
	}

	if err := helpers.Crear(helpers.UrlCrudNegocio+"/lead", solicitud, &solicitud); err != nil {
		return nil, err
	}

	// 6. Dejar registro de la actividad (crud_soporte)
	RegistrarActividad(datos.IdUsuario, "CREAR_SOLICITUD", fmt.Sprintf("Solicitud %s de %s por %.0f", solicitud.Radicado, producto.NombreProducto, datos.Monto), "negocio.lead")

	// 7. Cuota fija: cuota = monto * i / (1 - (1 + i)^-meses), con i = tasa anual / 12 / 100
	cuota := datos.Monto / float64(datos.PlazoMeses)

	i := producto.TasaMinima / 12 / 100
	if i > 0 {
		cuota = datos.Monto * i / (1 - math.Pow(1+i, float64(-datos.PlazoMeses)))
	}

	return map[string]interface{}{
		"id_solicitud":   solicitud.Id,
		"radicado":       solicitud.Radicado,
		"producto":       producto.NombreProducto,
		"monto":          datos.Monto,
		"plazo_meses":    datos.PlazoMeses,
		"cuota_estimada": math.Round(cuota*100) / 100,
		"estado":         "En estudio",
		"asesor":         nombreAsesor,
	}, nil
}

// SolicitudesUsuario: solicitudes de credito de un usuario.
func SolicitudesUsuario(idUsuario int) ([]map[string]interface{}, error) {

	var solicitudes []models.Lead

	if err := helpers.ConsultarLista(fmt.Sprintf("%s/lead?query=id_usuario:%d,activo:true&limit=1000", helpers.UrlCrudNegocio, idUsuario), &solicitudes); err != nil {
		return nil, err
	}

	lista := []map[string]interface{}{}

	for _, solicitud := range solicitudes {
		lista = append(lista, map[string]interface{}{
			"id_solicitud":    solicitud.Id,
			"radicado":        solicitud.Radicado,
			"producto":        solicitud.TipoCredito,
			"monto":           solicitud.MontoInteres,
			"plazo_meses":     solicitud.PlazoInteres,
			"estado":          solicitud.EstadoLead,
			"fecha_solicitud": solicitud.FechaGeneracion,
		})
	}

	return lista, nil
}
