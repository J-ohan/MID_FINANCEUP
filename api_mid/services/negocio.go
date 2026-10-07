package services

import (
	"fmt"
	"math"
	"strings"

	"api_mid_financeup/helpers"
	"api_mid_financeup/models"
)

// Ofertas devuelve los productos de credito activos con los datos de su banco.
func Ofertas() ([]models.Oferta, error) {
	ofertas := []models.Oferta{}

	var productos []models.ProductoCrediticio
	if err := helpers.Consultar("negocio", "producto_crediticio?"+helpers.Filtro("activo", true), &productos); err != nil {
		return ofertas, err
	}
	bancos, err := bancosPorId()
	if err != nil {
		return ofertas, err
	}

	for _, p := range productos {
		banco := bancos[p.IdBanco.Id]
		ofertas = append(ofertas, models.Oferta{
			IdProducto:  p.Id,
			Producto:    p.NombreProducto,
			Descripcion: p.Descripcion,
			Banco:       banco.NombreBanco,
			LogoBanco:   banco.UrlLogo,
			SitioWeb:    banco.SitioWeb,
			MontoMinimo: p.MontoMinimo,
			MontoMaximo: p.MontoMaximo,
			TasaMinima:  p.TasaMinima,
			TasaMaxima:  p.TasaMaxima,
			PlazoMinimo: p.PlazoMinimo,
			PlazoMaximo: p.PlazoMaximo,
			Requisitos:  separarRequisitos(p.Requisitos),
		})
	}
	return ofertas, nil
}

// CrearSolicitudCredito valida la solicitud contra el usuario (crud_auth) y el
// producto (crud_negocio), asigna un asesor del banco y la guarda como "lead".
func CrearSolicitudCredito(datos models.DatosSolicitudCredito) (models.SolicitudCredito, error) {
	if datos.IdUsuario <= 0 || datos.IdProducto <= 0 {
		return models.SolicitudCredito{}, helpers.NuevoError(400, "id_usuario e id_producto son obligatorios")
	}
	if err := ExisteUsuario(datos.IdUsuario); err != nil {
		return models.SolicitudCredito{}, err
	}

	var producto models.ProductoCrediticio
	if err := helpers.Consultar("negocio", fmt.Sprintf("producto_crediticio/%d", datos.IdProducto), &producto); err != nil {
		return models.SolicitudCredito{}, mensajeNoEncontrado(err, "El producto de credito no existe")
	}
	if !producto.Activo {
		return models.SolicitudCredito{}, helpers.NuevoError(400, "El producto de credito ya no esta disponible")
	}
	if datos.Monto < producto.MontoMinimo || datos.Monto > producto.MontoMaximo {
		return models.SolicitudCredito{}, helpers.NuevoError(400, fmt.Sprintf(
			"El monto debe estar entre %.0f y %.0f", producto.MontoMinimo, producto.MontoMaximo))
	}
	if datos.PlazoMeses < producto.PlazoMinimo || datos.PlazoMeses > producto.PlazoMaximo {
		return models.SolicitudCredito{}, helpers.NuevoError(400, fmt.Sprintf(
			"El plazo debe estar entre %d y %d meses", producto.PlazoMinimo, producto.PlazoMaximo))
	}

	// Se asigna el primer asesor activo del banco (si hay alguno).
	var asesores []models.AsesorBancario
	helpers.Consultar("negocio", "asesor_bancario?"+helpers.Filtro("id_banco", producto.IdBanco.Id, "activo", true), &asesores)

	lead := models.Lead{
		IdUsuario:     datos.IdUsuario,
		IdProducto:    &models.Ref{Id: producto.Id},
		TipoCredito:   producto.NombreProducto,
		MontoInteres:  datos.Monto,
		PlazoInteres:  datos.PlazoMeses,
		EstadoLead:    "nuevo",
		Observaciones: datos.Observaciones,
		Activo:        true,
	}
	if len(asesores) > 0 {
		lead.IdAsesor = &models.Ref{Id: asesores[0].Id}
	}
	if err := helpers.Enviar("POST", "negocio", "lead", lead, &lead); err != nil {
		return models.SolicitudCredito{}, err
	}

	RegistrarActividad(datos.IdUsuario, "CREAR_SOLICITUD",
		fmt.Sprintf("Solicitud de %s por %.0f a %d meses", producto.NombreProducto, datos.Monto, datos.PlazoMeses),
		"negocio.lead")

	solicitudes, err := armarSolicitudes([]models.Lead{lead})
	if err != nil || len(solicitudes) == 0 {
		return models.SolicitudCredito{IdSolicitud: lead.Id, Estado: lead.EstadoLead}, nil
	}
	return solicitudes[0], nil
}

// SolicitudesDeUsuario lista las solicitudes de credito de un usuario.
func SolicitudesDeUsuario(idUsuario int) ([]models.SolicitudCredito, error) {
	if err := ExisteUsuario(idUsuario); err != nil {
		return []models.SolicitudCredito{}, err
	}
	var leads []models.Lead
	if err := helpers.Consultar("negocio", "lead?"+helpers.Filtro("id_usuario", idUsuario, "activo", true)+"&sortby=fecha_generacion&order=desc", &leads); err != nil {
		return []models.SolicitudCredito{}, err
	}
	return armarSolicitudes(leads)
}

// armarSolicitudes completa cada lead con el nombre del producto, el banco,
// el asesor y la cuota mensual estimada.
func armarSolicitudes(leads []models.Lead) ([]models.SolicitudCredito, error) {
	solicitudes := []models.SolicitudCredito{}
	bancos, err := bancosPorId()
	if err != nil {
		return solicitudes, err
	}

	for _, l := range leads {
		solicitud := models.SolicitudCredito{
			IdSolicitud:    l.Id,
			Producto:       l.TipoCredito,
			Monto:          l.MontoInteres,
			PlazoMeses:     l.PlazoInteres,
			Estado:         l.EstadoLead,
			FechaSolicitud: l.FechaGeneracion,
		}
		var producto models.ProductoCrediticio
		if l.IdProducto != nil && helpers.Consultar("negocio", fmt.Sprintf("producto_crediticio/%d", l.IdProducto.Id), &producto) == nil {
			solicitud.Producto = producto.NombreProducto
			solicitud.Banco = bancos[producto.IdBanco.Id].NombreBanco
			solicitud.CuotaEstimada = cuotaMensual(l.MontoInteres, producto.TasaMinima, l.PlazoInteres)
		}
		var asesor models.AsesorBancario
		if l.IdAsesor != nil && helpers.Consultar("negocio", fmt.Sprintf("asesor_bancario/%d", l.IdAsesor.Id), &asesor) == nil {
			solicitud.Asesor = asesor.Nombre + " " + asesor.Apellido
		}
		solicitudes = append(solicitudes, solicitud)
	}
	return solicitudes, nil
}

// cuotaMensual calcula la cuota fija de un credito (sistema frances).
// tasaAnual viene en porcentaje, por ejemplo 12 = 12% anual.
func cuotaMensual(monto, tasaAnual float64, meses int) float64 {
	if meses <= 0 {
		return 0
	}
	tasaMensual := tasaAnual / 100 / 12
	if tasaMensual == 0 {
		return redondear(monto / float64(meses))
	}
	cuota := monto * tasaMensual / (1 - math.Pow(1+tasaMensual, float64(-meses)))
	return redondear(cuota)
}

func bancosPorId() (map[int]models.Banco, error) {
	var bancos []models.Banco
	if err := helpers.Consultar("negocio", "banco?limit=1000", &bancos); err != nil {
		return nil, err
	}
	resultado := map[int]models.Banco{}
	for _, b := range bancos {
		resultado[b.Id] = b
	}
	return resultado, nil
}

// separarRequisitos convierte "Cedula, comprobante ingresos" en una lista.
func separarRequisitos(texto string) []string {
	lista := []string{}
	for _, r := range strings.Split(texto, ",") {
		if r = strings.TrimSpace(r); r != "" {
			lista = append(lista, r)
		}
	}
	return lista
}
