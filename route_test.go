package lotus

import (
	"fmt"
	"github.com/brunvieira/fastalice"
	"github.com/buaazp/fasthttprouter"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

func tagMiddleware(tag string) fastalice.Constructor {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString(tag)
			next(ctx)
		}
	}
}

func testMethodHandler(path string) RequestHandler {
	return func(ctx *Context) {
		ctx.WriteString(path)
		if data := ctx.UserValue("data"); data != nil {
			if dataMap, ok := data.(map[string]interface{}); ok {
				ctx.WriteString(fmt.Sprintf("[%s]=%v", "Foo", dataMap["Foo"]))
				ctx.WriteString(fmt.Sprintf("[%s]=%v", "Bar", dataMap["Bar"]))
				if dataMap["FooBar"] != nil {
					ctx.WriteString(fmt.Sprintf("[%s]=%v", "FooBar", dataMap["FooBar"]))
				}
			}
		}
	}
}

func routeContractForMethod(method Method, path string) *RouteContract {
	return &RouteContract{
		"Test" + string(method),
		"Test " + string(method) + " Method",
		method,
		path,
		DataHandlerConfig{},
		nil,
	}
}

func routeForContract(contract *RouteContract, path string, middlewares []fastalice.Constructor, dataHandler fastalice.Constructor) *Route {
	route := Route{
		RouteContract:  contract,
		RequestHandler: testMethodHandler(path),
		Middlewares:    middlewares,
		DataHandler:    dataHandler,
	}
	return &route
}

func testRequestToHandler(
	t *testing.T,
	method Method,
	url string,
	body []byte,
	testName string,
	expectedStatus int,
) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	if body != nil {
		req.SetBody(body)
	}

	resp := fasthttp.AcquireResponse()

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

func testMethod(
	t *testing.T,
	method Method,
	port string,
) {
	path := "/echo"
	contract := routeContractForMethod(method, path)
	route := routeForContract(contract, path, nil, nil)

	router := fasthttprouter.New()
	route.startRoute(router, "")

	url := "localhost" + ":" + port
	ln, _ := net.Listen("tcp", url)
	go fasthttp.Serve(ln, router.Handler)
	defer ln.Close()

	resp := testRequestToHandler(t, method, "http://"+url+path, nil, string(method), fasthttp.StatusOK)
	defer fasthttp.ReleaseResponse(resp)

	respBody := resp.Body()
	assert.NotEmptyf(t, respBody, "Reading the body response should not return an error")
	assert.Equal(t, path, string(respBody), "Body output should be the correct method")
}

func TestGet(t *testing.T) {
	testMethod(t, "", "8080")
}

func TestDelete(t *testing.T) {
	testMethod(t, DELETE, "8081")
}

func TestPost(t *testing.T) {
	testMethod(t, POST, "8082")
}

func TestPut(t *testing.T) {
	testMethod(t, PUT, "8083")
}

func TestMiddlewareDataHandlerOrder(t *testing.T) {
	path := "/middlewares"
	method := Method(POST)
	payload := ServiceRequest{
		Body: map[string]interface{}{
			"Foo": "foo",
			"Bar": "bar",
		},
	}
	middlewares := []fastalice.Constructor{
		tagMiddleware("/t1"),
		tagMiddleware("/t2"),
		tagMiddleware("/t3"),
	}

	contract := routeContractForMethod(method, path)
	route := routeForContract(contract, path, middlewares, nil)

	router := fasthttprouter.New()
	route.startRoute(router, "")

	url := "localhost:8084"
	ln, _ := net.Listen("tcp", url)
	go fasthttp.Serve(ln, router.Handler)
	defer ln.Close()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://" + url + path)
	route.prepareRequest(req, payload)

	fasthttp.Do(req, resp)

	body := resp.Body()
	assert.NotEmptyf(t, body, "Reading the body response should not return an error")
	assert.Equal(t, "/t1/t2/t3"+path+"[Foo]=foo[Bar]=bar", string(body), "Body output should be the correct write order")

}

func TestJsonData(t *testing.T) {
	path := "/echo"
	method := Method(POST)
	contract := routeContractForMethod(method, path)
	route := routeForContract(contract, path, nil, nil)
	payload := ServiceRequest{
		DataType: JSON,
		Body: map[string]interface{}{
			"Foo": "foo",
			"Bar": "bar",
		},
	}

	router := fasthttprouter.New()
	route.startRoute(router, "")

	url := "localhost:8085"
	ln, _ := net.Listen("tcp", url)
	go fasthttp.Serve(ln, router.Handler)
	defer ln.Close()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://" + url + path)
	route.prepareRequest(req, payload)

	fasthttp.Do(req, resp)

	body := resp.Body()
	assert.NotEmptyf(t, body, "Reading the body response should not return an error")
	assert.Equal(t, path+"[Foo]=foo[Bar]=bar", string(body), "Body output should be the correct path and data")
}

func TestFormData(t *testing.T) {
	path := "/echo"
	method := Method(POST)
	contract := routeContractForMethod(method, path)
	route := routeForContract(contract, path, nil, nil)
	payload := ServiceRequest{
		DataType: Form,
		Body: map[string]interface{}{
			"Foo": "foo",
			"Bar": "bar",
			"FooBar": []string{"foo", "bar"},
		},
	}
	route.DataHandler = func (next fasthttp.RequestHandler) fasthttp.RequestHandler{
		return func (ctx *fasthttp.RequestCtx){
			data := map[string]interface{}{}
			key := route.userValueKey()
			args := ctx.PostArgs()

			if args.Has("Foo") {
				data["Foo"] = string(args.Peek("Foo"))
			}
			if args.Has("Bar") {
				data["Bar"] = string(args.Peek("Bar"))
			}
			if args.Has("FooBar") {
				fooBar := args.PeekMulti("FooBar")
				fooBarSlice := make([]string, len(fooBar))
				for i, v := range fooBar {
					fooBarSlice[i] = string(v)
				}
				data["FooBar"] = fooBarSlice
			}
			ctx.SetUserValue(key, data)
			next(ctx)
		}
	}

	router := fasthttprouter.New()
	route.startRoute(router, "")

	url := "localhost:8087"
	ln, _ := net.Listen("tcp", url)
	go fasthttp.Serve(ln, router.Handler)
	defer ln.Close()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://" + url + path)
	route.prepareRequest(req, payload)

	fasthttp.Do(req, resp)

	body := resp.Body()
	assert.NotEmptyf(t, body, "Reading the body response should not return an error")
	assert.Equal(t, path+"[Foo]=foo[Bar]=bar[FooBar]=[foo bar]", string(body), "Body output should be the correct path and data")
}

