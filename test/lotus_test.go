package test

import (
	"fmt"
	"github.com/brunvieira/lotus"
	"github.com/brunvieira/lotus/test/contract"
	"github.com/brunvieira/lotus/test/echo_service"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"testing"
)

func testRequestToHandler(
	t *testing.T,
	method lotus.Method,
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
	assert.NotNil(t, resp, "RequestHandler response must not be nil")
	assert.Equal(t, expectedStatus, resp.StatusCode(), fmt.Sprintf("%s test should return a %d status", testName, expectedStatus))
	if err != nil {
		panic(err)
	}
	return resp
}

func TestNoMiddlewares(t *testing.T) {
	service := echo_service.EchoService
	go service.Start()
	defer service.Stop()

	endpoint := service.Suffix() + contract.SimpleEchoRouteContract.Path
	url, err := service.RouteUrl(contract.SimpleEchoRouteContract.Label)
	assert.Nil(t, err, "Service must build a valid url for a route")

	if err != nil {
		t.Fatal(err)
	}

	resp := testRequestToHandler(t, lotus.GET, url, "No Middleware", fasthttp.StatusOK)
	defer fasthttp.ReleaseResponse(resp)

	body := resp.Body()
	assert.NotEmptyf(t, body, "Reading the body response should not return an error")
	assert.Equal(t, endpoint, string(body), "Body output should be the correct namespace format")
}
