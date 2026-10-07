package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:AsesorBancarioController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:BancoController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ContactoAsesorController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ConversacionUsuarioAsesorController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:CreditoDesembolsadoController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:LeadController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:ProductoCrediticioController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:id`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"] = append(beego.GlobalControllerRouter["crud_negocio/controllers:TransaccionComisionController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:id`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
