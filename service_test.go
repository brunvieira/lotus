package lotus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"testing"
)


var contract = Contract{
	Services: []ServiceContract{
		{
			Label: "EchoService",
			Host:      "localhost",
			Port:      8080,
			Namespace: "nomiddlewaretest",
			RoutesContracts: []RouteContract{
				{
					Label: "SimpleEcho",
					Description: "A Simple route that outputs the Request URI",
					Path: "/echo",
				},
			},
		},
	},
}

func echo(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, string(ctx.RequestURI()))
}

func testRequestToHandler(
	t *testing.T,
	method Method,
	url string,
	testName string,
	expectedStatus int,
) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(string(method))
	req.SetRequestURI(url)

	err := fasthttp.Do(req, resp)
	assert.Nil(t, err, "Sending the request must not return an error")
	assert.NotNil(t, resp, "Request response must not be nil")
	assert.Equal(t, expectedStatus, resp.StatusCode(), fmt.Sprintf("%s test should return a %d status", testName, expectedStatus))
	if err != nil {
		panic(err)
	}
	return resp
}


func TestNoMiddlewares(t *testing.T) {
	service := Service{
		ServiceContract: contract.serviceContract("EchoService"),
	}
	simpleEchoRoute := service.SetupRoute("SimpleEcho", echo)
	go service.Start()
	defer service.Stop()

	endpoint := service.suffix() + simpleEchoRoute.Path
	url, err := service.RouteUrl(simpleEchoRoute.Label)
	assert.Nil(t, err, "Service must build a valid url for a route")

	if err != nil {
		t.Fatal(err)
	}

	resp := testRequestToHandler(t, GET, url, "No Middleware", fasthttp.StatusOK)
	defer fasthttp.ReleaseResponse(resp)

	body := resp.Body()
	assert.NotEmptyf(t, body, "Reading the body response should not return an error")
	assert.Equal(t, endpoint, string(body), "Body output should be the correct namespace format")
}
