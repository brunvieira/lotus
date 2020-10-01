package lotus

import (
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"log"
	"testing"
	"time"
)

var SimpleEchoRouteContract = RouteContract{
	Label: "SimpleEcho",
	Description: "A Simple route that outputs the Request URI",
	Path: "/echo",
}

var PostEchoRouteContract = RouteContract{
	Label: "PostEcho",
	Description: "A route that outputs the contents of it's body",
	Path: "/echo",
	Method: fasthttp.MethodPost,
}

var EmptyRoute = Route{}

var EchoServiceContract = ServiceContract{
	Label: "EchoService",
	Host:      "localhost",
	Namespace: "servicetest",
	Version: "v1",
	Protocol: HTTP,
	RoutesContracts: []RouteContract{
		SimpleEchoRouteContract,
		PostEchoRouteContract,
	},
}

var ServiceContractWithDefaultValues = ServiceContract{
	Label: "DefaultEchoService",
	Host:      "",
	Namespace: "",
	Port: 8090,
	RoutesContracts: []RouteContract{
		SimpleEchoRouteContract,
		PostEchoRouteContract,
	},
}

var contract = &Contract{
	Services: []ServiceContract{
		EchoServiceContract,
		ServiceContractWithDefaultValues,
	},
}


func echo(ctx *fasthttp.RequestCtx) {
	ctx.Write(ctx.Method())
	ctx.WriteString(":")
	ctx.Write(ctx.RequestURI())
}

func TestContractFindServiceByLabel(t *testing.T) {
	foundServiceContract := contract.serviceContract("EchoService")
	assert.NotNil(t, foundServiceContract, "serviceContract must return a non nil found service")
	assert.Equal(t, EchoServiceContract, *foundServiceContract, "Found service contract must match EchoServiceContract")

	notFoundServiceContract := contract.serviceContract("NotFound")
	assert.Nil(t, notFoundServiceContract, "serviceContract must return a nil for not found service")
}

func TestFindLabelOnContract(t *testing.T) {
	foundRouteContract := EchoServiceContract.RouteContractByLabel("SimpleEcho")

	assert.NotNil(t, foundRouteContract, "RouteContractByLabel should return a non nil route contract for 'SimpleEcho' value")
	assert.Equal(t, SimpleEchoRouteContract, *foundRouteContract, "RouteContractByLabel should return the correct SimpleEcho route contract")

	notFoundRouteContract := EchoServiceContract.RouteContractByLabel("Foo")
	assert.Nil(t, notFoundRouteContract, "RouteContractByLabel should return a nil contract for 'Foo' value")
}

func TestSetupRoute(t *testing.T) {
	service := Service{ServiceContract: &EchoServiceContract}

	shouldPanic(t, func () { service.SetupRoute("EmptyRoute", echo, nil, nil) })
	notFoundRouteUrl, err := service.RouteUrl("NotFoundRoute")
	assert.Empty(t, notFoundRouteUrl, "A non found route should return an empty url")

	simpleEchoRoute := service.SetupRoute("SimpleEcho", echo, nil, nil)
	assert.NotNil(t, simpleEchoRoute, "SetupRoute should return a non nil route for 'SimpleEcho' value")
	assert.NotNil(t, err, "RouteUrl should return a non nil route error for a not found route")

	simpleEchoRouteUrl, err := service.RouteUrl("SimpleEcho")
	assert.NotEmpty(t, simpleEchoRouteUrl, "A non empty route should return a valid url")

	status := service.Status()
	assert.False(t, status.IsRunning, "Service must be running")
	assert.Equal(t, len(service.routes), status.RegisteredRoutes, "Service must have same number of routes registered")
	err = service.Stop()
	assert.NotNil(t, err, "Stopping a non started service should return an error")

}

func TestPanicWhenServicePortIsTaken(t *testing.T) {
	service := Service{ServiceContract: &EchoServiceContract}
	simpleEchoRoute := service.SetupRoute("SimpleEcho", echo, nil, nil)
	assert.NotNil(t, simpleEchoRoute, "SetupRoute should return a non nil route for 'SimpleEcho' value")

	postEchoRoute := Route{
		RouteContract: &PostEchoRouteContract,
		Endpoint: echo,
		Middlewares: nil,
		DataHandlers: nil,
	}
	service.AddRoute(&postEchoRoute)

	anotherService := Service{ServiceContract: &EchoServiceContract}
	anotherSimpleEchoRoute := anotherService.SetupRoute("SimpleEcho", echo, nil, nil)
	anotherPostEchoRoute := anotherService.SetupRoute("PostEcho", echo, nil, nil)
	assert.NotNil(t, anotherSimpleEchoRoute, "SetupRoute should return a non nil route for 'SimpleEcho' value")
	assert.NotNil(t, anotherPostEchoRoute, "SetupRoute should return a non nil route for 'PostEcho' value")

	go service.Start()
	defer service.Stop()

	shouldPanic(t, func() {
		anotherService.Start()
		anotherService.Stop()
	})
}

func TestStartService(t *testing.T) {
	service := Service{ServiceContract: &ServiceContractWithDefaultValues}
	simpleEchoRoute := service.SetupRoute("SimpleEcho", echo, nil, nil)
	assert.NotNil(t, simpleEchoRoute, "SetupRoute should return a non nil route for 'SimpleEcho' value")

	postEchoRoute := Route{
		RouteContract: &PostEchoRouteContract,
		Endpoint: echo,
		Middlewares: nil,
		DataHandlers: nil,
	}
	service.AddRoute(&postEchoRoute)

	url, err := service.RouteUrl(SimpleEchoRouteContract.Label)
	log.Println("url", url)
	assert.Nil(t, err, "Service must build a valid url for a route")

	go service.Start()
	defer service.Stop()

	time.Sleep(2 * time.Second)

	resp := testRequestToHandler(t, GET, url, nil, "Start", fasthttp.StatusOK)
	defer fasthttp.ReleaseResponse(resp)

	status := service.Status()
	assert.True(t, status.IsRunning, "Service must be running")
	assert.Equal(t, len(service.routes), status.RegisteredRoutes, "Service must have same number of routes registered")
}

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}