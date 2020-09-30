package contract

import (
	"github.com/brunvieira/lotus"
	"github.com/valyala/fasthttp"
)

var SimpleEchoRouteContract = lotus.RouteContract{
	Label: "SimpleEcho",
	Description: "A Simple route that outputs the Request URI",
	Path: "/echo",
}

var PostEchoRouteContract = lotus.RouteContract{
	Label: "PostEcho",
	Description: "A route that outputs the contents of it's body",
	Path: "/echo",
	Method: fasthttp.MethodPost,
}

var EchoServiceContract = lotus.ServiceContract{
	Label: "EchoService",
	Host:      "localhost",
	Port:      8080,
	Namespace: "nomiddlewaretest",
	RoutesContracts: []lotus.RouteContract{
		SimpleEchoRouteContract,
		PostEchoRouteContract,
	},
}