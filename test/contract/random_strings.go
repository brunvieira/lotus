package contract

import (
	"github.com/brunvieira/lotus"
	"github.com/valyala/fasthttp"
)

var (
	RandomStringsRouteContract = lotus.RouteContract{
		Label:       "RandomStrings",
		Description: "A route that generate random strings",
		Method:      fasthttp.MethodPost,
		Path:        "/random",
	}
	RandomStringsServiceContract = lotus.ServiceContract{
		Label:       "RandomStringsService",
		Description: "Service that generates random strings",
		Host:        "localhost",
		Namespace:   "random_strings",
		Port:        8081,
		RoutesContracts: []lotus.RouteContract{
			RandomStringsRouteContract,
		},
	}
)
