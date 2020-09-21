package lotus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"testing"
)

func echo(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, string(ctx.RequestURI()))
}
var simpleEchoRoute = Route {
	Label: "SimpleEcho",
	Description: "A Simple route that outputs the Request URI",
	Endpoint: echo,
	Path: "/echo",
}

func testRequestToHandler(
	t *testing.T,
	method Method,
	url string,
	testName string,
	expectedStatus int,
) *http.Response {
	req, err := http.NewRequest(string(method), url, nil)
	assert.Nil(t, err, fmt.Sprintf("%s test should be able to create a request", testName))

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err, "Sending the request must not return an error")
	assert.NotNil(t, resp, "Request response must not be nil")
	assert.Equal(t, expectedStatus, resp.StatusCode, fmt.Sprintf("%s test should return a %d status", testName, expectedStatus))
	if err != nil {
		panic(err)
	}
	return resp
}


func TestNoMiddlewares(t *testing.T) {
	service := Service {
		Host:      "localhost",
		Port:      8080,
		Namespace: "nomiddlewaretest",
		Routes: []*Route {
			&simpleEchoRoute,
		},
	}
	go service.Start()
	defer service.Stop()

	resp := testRequestToHandler(t, GET, "http://localhost:8080/nomiddlewaretest/v0/echo", "No Middleware", fasthttp.StatusOK)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err, "Reading the body response should not return an error")
	assert.Equal(t, service.suffix() + simpleEchoRoute.Path, string(body), "Body output should be the correct namespace format")
}