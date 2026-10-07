package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:CategoriaController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:EditarMetaController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:FinanzasController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:InversionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MetaController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoIngresoEgresoController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoInversionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:MovimientoMetaController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:NivelRiesgoController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:SolicitudConsolidacionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoInversionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoIngresoMetaController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"] = append(beego.GlobalControllerRouter["crud_finanzas/controllers:TipoInversionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
