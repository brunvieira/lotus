package lotus

//
//import (
//	"errors"
//	"github.com/buaazp/fasthttprouter"
//	"github.com/stretchr/testify/assert"
//	"github.com/valyala/fasthttp"
//	"net"
//	"strconv"
//	"testing"
//	"time"
//)
//
//type SamplePayload struct {
//	Foo        string
//	FooBar     []string
//	IntValue   int     `json:",string"`
//	FloatValue float64 `json:",string"`
//	Time       time.Time
//	Boolean    bool
//}
//
//var defaultPayload = SamplePayload{
//	Foo:        "foo",
//	FooBar:     []string{"foo", "bar"},
//	IntValue:   1200,
//	FloatValue: 3.1412,
//	Time:       time.Now(),
//	Boolean:    true,
//}
//
//var getPayload = map[string]string{
//	"foo": "bar",
//	"bar": "foo",
//}
//
//func TestNewDataContract(t *testing.T) {
//	dataType := DataType(JSON)
//	dataContract := NewDataContract(dataType)
//	assert.NotNil(t, dataContract, "NewDataContract should return a valid DataContract")
//	assert.Equal(t, dataType, dataContract.Type, "NewDataContract should have the initialized type")
//	assert.Equal(t, DefaultKey, dataContract.Key, "NewDataContract should have the default key")
//}
//
//func prepareRequestForDataType(
//	t *testing.T,
//	dataType DataType,
//	method Method,
//) *fasthttp.RequestHandler {
//	path := "/datahandler"
//
//	var dataHandler DataConverter
//	dataContract := NewDataContract(dataType)
//
//	dataHandler = &DataHandler{
//		DataContract: &dataContract,
//		ServiceRequest:      defaultPayload,
//	}
//
//	contract := routeContractForMethod(method, path, []DataContract{dataContract})
//	route := routeForContract(contract, path, nil, []DataConverter{dataHandler})
//
//	req := fasthttp.AcquireRequest()
//	err := dataHandler.PrepareRouteRequest(req, route)
//	assert.Nil(t, err, "The data should be prepared without errors")
//
//	return req
//}
//
//func testReceiveRequest(
//	t *testing.T,
//	router *fasthttprouter.Router,
//	req *fasthttp.RequestHandler,
//	path string,
//	port string,
//) *fasthttp.Response {
//	url := "localhost:" + port
//	ln, _ := net.Listen("tcp", url)
//	go fasthttp.Serve(ln, router.Handler)
//	defer ln.Close()
//
//	req.SetRequestURI("http://" + url + path)
//
//	resp := fasthttp.AcquireResponse()
//	err := fasthttp.Do(req, resp)
//	assert.Nil(t, err, "Sending the request must not return an error")
//
//	return resp
//}
//
//func testReceivePostRequest(
//	t *testing.T,
//	req *fasthttp.RequestHandler,
//	endpoint fasthttp.RequestHandler,
//	dataType DataType,
//	port string,
//) {
//	path := "/data"
//	dataContract := NewDataContract(dataType)
//	dataHandler := &DataHandler{
//		DataContract: &dataContract,
//		ServiceRequest:      defaultPayload,
//	}
//
//	contract := routeContractForMethod(POST, path, []DataContract{dataContract})
//	route := routeForContract(contract, path, nil, []DataConverter{dataHandler})
//	route.RequestHandler = endpoint
//
//	router := fasthttprouter.New()
//	route.startRoute(router, "")
//
//	resp := testReceiveRequest(t, router, req, path, port)
//	defer fasthttp.ReleaseResponse(resp)
//
//	respBody := resp.Body()
//	assert.NotEmptyf(t, respBody, "Reading the body response should not return an error")
//}
//
//func TestPrepareBinaryRequest(t *testing.T) {
//	req := prepareRequestForDataType(t, Binary, POST)
//	defer fasthttp.ReleaseRequest(req)
//
//	body := req.Body()
//	assert.NotNil(t, body, "RequestHandler should have a valid body")
//	assert.NotEmpty(t, body, "Body must be not empty")
//
//	var concluded bool
//	var err error
//	postEcho := func(ctx *fasthttp.RequestCtx) {
//		data := ctx.UserValue(DefaultKey)
//		if obj, ok := data.(SamplePayload); ok {
//			assert.Equal(t, defaultPayload, obj, "Posted json payload must match")
//		} else {
//			err = errors.New("not able to convert to SamplePayload")
//		}
//
//		ctx.Write(ctx.Path())
//		concluded = true
//	}
//
//	testReceivePostRequest(t, req, postEcho, Binary, "9095")
//	for !concluded {
//		assert.Nil(t, err, "Should be able to convert to sample payload")
//	}
//}
//
//func TestPrepareJsonRequest(t *testing.T) {
//	req := prepareRequestForDataType(t, JSON, POST)
//	defer fasthttp.ReleaseRequest(req)
//
//	body := req.Body()
//	assert.NotNil(t, body, "RequestHandler should have a valid body")
//	assert.NotEmpty(t, body, "Body must be not empty")
//
//	var concluded bool
//	var err error
//	postEcho := func(ctx *fasthttp.RequestCtx) {
//		data := ctx.UserValue(DefaultKey)
//		if obj, ok := data.(SamplePayload); ok {
//			assert.Equal(t, defaultPayload, obj, "Posted json payload must match")
//		} else {
//			err = errors.New("not able to convert to SamplePayload")
//		}
//		ctx.Write(ctx.Path())
//		concluded = true
//
//	}
//
//	testReceivePostRequest(t, req, postEcho, JSON, "9091")
//	for !concluded {
//		assert.Nil(t, err, "Should be able to convert to sample payload")
//	}
//}
//
//func TestPrepareFormRequest(t *testing.T) {
//	req := prepareRequestForDataType(t, Form, POST)
//	defer fasthttp.ReleaseRequest(req)
//
//	body := req.Body()
//	assert.NotNil(t, body, "RequestHandler should have a valid body")
//	assert.NotEmpty(t, body, "Body must be not empty")
//
//	postEcho := func(ctx *fasthttp.RequestCtx) {
//		form := ctx.PostArgs()
//
//		foo := string(form.Peek("Foo"))
//		assert.NotEmpty(t, foo, "Must be able to find a posted string")
//		assert.Equal(t, defaultPayload.Foo, foo, "String value must match")
//
//		var fooBar []string
//		for _, v := range form.PeekMulti("FooBar") {
//			fooBar = append(fooBar, string(v))
//		}
//		assert.NotEmpty(t, fooBar, "Must be able to find a posted array of strings")
//		assert.Equal(t, defaultPayload.FooBar, fooBar, "String array value must match")
//
//		intValue, err := strconv.Atoi(string(form.Peek("IntValue")))
//		assert.Nil(t, err, "Must be able to convert form value to int")
//		assert.NotEmpty(t, intValue, "Must be able to find a posted int value")
//		assert.Equal(t, defaultPayload.IntValue, intValue, "Int value must match")
//
//		floatValue, err := strconv.ParseFloat(string(form.Peek("FloatValue")), 64)
//		assert.Nil(t, err, "Must be able to convert form value to float64")
//		assert.NotEmpty(t, floatValue, "Must be able to find a posted float value")
//		assert.Equal(t, defaultPayload.FloatValue, floatValue, "Float value must match")
//
//		tm, err := time.Parse(time.RFC3339, string(form.Peek("Time")))
//		assert.Nil(t, err, "Must be able to convert form value to time")
//		assert.NotEmpty(t, tm, "Must be able to find a posted time")
//		assert.Equal(t, defaultPayload.Time.Format(time.RFC3339), tm.Format(time.RFC3339), "Time value must match")
//
//		boolean, err := strconv.ParseBool(string(form.Peek("Boolean")))
//		assert.Nil(t, err, "Must be able to convert form value to bool")
//		assert.NotEmpty(t, boolean, "Must be able to find a posted boolean")
//		assert.Equal(t, defaultPayload.Boolean, boolean, "Boolean value must match")
//
//		ctx.WriteString(form.String())
//	}
//
//	testReceivePostRequest(t, req, postEcho, Form, "9092")
//}
//
//func TestPrepareRouteParams(t *testing.T) {
//	path := "/data/:foo/:bar"
//	port := "9093"
//
//	var dataHandler DataConverter
//	dataContract := NewDataContract(RouteParams)
//
//	dataHandler = &DataHandler{
//		DataContract: &dataContract,
//		ServiceRequest:      getPayload,
//	}
//
//	getEcho := func(ctx *fasthttp.RequestCtx) {
//		foo := ctx.UserValue("foo").(string)
//		assert.NotEmpty(t, foo, "Parameter foo should not be empty")
//		assert.Equal(t, getPayload["foo"], foo, "Parameter foo should match payload foo field")
//
//		bar := ctx.UserValue("bar").(string)
//		assert.NotEmpty(t, bar, "Parameter bar should not be empty")
//		assert.Equal(t, getPayload["bar"], bar, "Parameter bar should match payload bar field")
//
//		ctx.Write(ctx.Path())
//	}
//
//	contract := routeContractForMethod(GET, path, []DataContract{dataContract})
//	route := routeForContract(contract, path, nil, []DataConverter{dataHandler})
//	route.RequestHandler = getEcho
//
//	router := fasthttprouter.New()
//	route.startRoute(router, "")
//
//	url := "localhost:" + port
//	ln, _ := net.Listen("tcp", url)
//	go fasthttp.Serve(ln, router.Handler)
//	defer ln.Close()
//
//	req := fasthttp.AcquireRequest()
//	fasthttp.ReleaseRequest(req)
//
//	req.SetRequestURI("http://" + url + path)
//	err := dataHandler.PrepareRouteRequest(req, route)
//	assert.Nil(t, err, "Route data should be prepared without errors")
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	err = fasthttp.Do(req, resp)
//	assert.Nil(t, err, "Sending the request must not return an error")
//}
//
//func TestPrepareQueryParams(t *testing.T) {
//	path := "/data"
//	port := "9094"
//
//	var dataHandler DataConverter
//	dataContract := NewDataContract(QueryParams)
//
//	dataHandler = &DataHandler{
//		DataContract: &dataContract,
//		ServiceRequest:      getPayload,
//	}
//
//	getEcho := func(ctx *fasthttp.RequestCtx) {
//		args := ctx.QueryArgs()
//		foo := string(args.Peek("foo"))
//		assert.NotEmpty(t, foo, "Parameter foo should not be empty")
//		assert.Equal(t, getPayload["foo"], foo, "Parameter foo should match payload foo field")
//
//		bar := string(args.Peek("bar"))
//		assert.NotEmpty(t, bar, "Parameter bar should not be empty")
//		assert.Equal(t, getPayload["bar"], bar, "Parameter bar should match payload bar field")
//
//		ctx.Write(ctx.Path())
//	}
//
//	contract := routeContractForMethod(GET, path, []DataContract{dataContract})
//	route := routeForContract(contract, path, nil, []DataConverter{dataHandler})
//	route.RequestHandler = getEcho
//
//	router := fasthttprouter.New()
//	route.startRoute(router, "")
//
//	url := "localhost:" + port
//	ln, _ := net.Listen("tcp", url)
//	go fasthttp.Serve(ln, router.Handler)
//	defer ln.Close()
//
//	req := fasthttp.AcquireRequest()
//	fasthttp.ReleaseRequest(req)
//
//	req.SetRequestURI("http://" + url + path)
//	err := dataHandler.PrepareRouteRequest(req, route)
//	assert.Nil(t, err, "Route data should be prepared without errors")
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	err = fasthttp.Do(req, resp)
//	assert.Nil(t, err, "Sending the request must not return an error")
//}
