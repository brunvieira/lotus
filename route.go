package lotus

import (
	"encoding/json"
	"github.com/brunvieira/fastalice"
	"github.com/buaazp/fasthttprouter"
	"github.com/mitchellh/mapstructure"
	"github.com/valyala/fasthttp"
	"github.com/vmihailenco/msgpack"
	"regexp"
	"strings"
)

// Method defines usable methods for endpoints
type Method string

const (
	// DELETE defines the DELETE type of request
	DELETE Method = fasthttp.MethodDelete
	// GET defines the GET type of request
	GET = fasthttp.MethodGet
	// POST defines the POST type of request
	POST = fasthttp.MethodPost
	// PUT defines the PUT type of request
	PUT = fasthttp.MethodPut

	// Default values
	DefaultRouteMethod Method = GET

	DefaultBodyDataType = Binary

	// DefaultKey is the UserValue key used when no Key is passed to the initializer
	DefaultKey string = "data"
)

var routerParamReg = regexp.MustCompile(`:[a-zA-Z0-9]*`)

type DataHandlerConfig struct {
	BodyType     DataType
	UserValueKey string
}

// RouteContract is the Contract description of a Route
type RouteContract struct {
	// Label is an identification for the route. It's used on status and log information
	Label string
	// Description is a short text that helps to document the route
	Description string
	// Method is the fasthttp method the route will listen upon
	Method Method
	// Path is the location where the route will listen to
	Path string
	// DataHandlerConfig is the configuration for the route DataHandler. This is an optional field
	DataHandlerConfig DataHandlerConfig
	// Data is the Data used. Use an empty struct value
	Data interface{}
}

func (route *RouteContract) prepareRequest(req *fasthttp.Request, payload ServiceRequest) (err error) {

	method := string(route.Method)
	req.Header.SetMethod(method)

	err = route.prepareRouteParams(req, payload)
	if err != nil {
		return
	}

	err = route.prepareQueryParams(req, payload)
	if err != nil {
		return
	}

	dataType := payload.DataType
	if len(dataType) == 0 {
		dataType = route.DataType()
	}

	switch dataType {
	case JSON:
		b, err := json.Marshal(payload.Body)
		req.Header.Set("Content-Type", string(JSON))
		req.SetBody(b)
		return err
	case MultipartForm:
	case Form:
		m, err := dataToUrlValues(payload.Body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", string(Form))
		req.SetBodyString(m.Encode())
		return err
	default:
		b, err := msgpack.Marshal(payload.Body)
		req.Header.Set("Content-Type", string(Binary))
		req.SetBody(b)
		return err
	}
	return nil
}

func (route *RouteContract) prepareQueryParams(req *fasthttp.Request, payload ServiceRequest) error {
	if len(payload.QueryParams) == 0 {
		return nil
	}
	m, err := dataToUrlValues(payload.QueryParams)
	if err != nil {
		return err
	}
	query := "?" + m.Encode()
	uriStr := req.URI().String()

	newUri := uriStr + query
	req.SetRequestURI(newUri)
	return nil
}

func (route *RouteContract) prepareRouteParams(req *fasthttp.Request, payload ServiceRequest) error {
	if len(payload.RouteParams) == 0 {
		return nil
	}

	m, err := dataToUrlValues(payload.RouteParams)
	if err != nil {
		return err
	}
	path := routerParamReg.ReplaceAllFunc([]byte(route.Path), replaceRouteMatches(m))
	uriStr := req.URI().String()
	newUri := strings.Replace(uriStr, route.Path, string(path), 1)
	req.SetRequestURI(newUri)
	return nil
}

func (route *RouteContract) method() Method {
	if len(route.Method) > 0 {
		return route.Method
	}
	return DefaultRouteMethod
}

func (route *RouteContract) DataType() DataType {
	if route.DataHandlerConfig.BodyType != "" {
		return route.DataHandlerConfig.BodyType
	}
	return DefaultBodyDataType
}

func (route *RouteContract) userValueKey() string {
	if len(route.DataHandlerConfig.UserValueKey) > 0 {
		return route.DataHandlerConfig.UserValueKey
	}
	return DefaultKey
}

// Route is the definition of a route and what's used to initialize a route
type Route struct {
	*RouteContract
	// The endpoint is the path used to receive the request
	RequestHandler RequestHandler
	// Middlewares functions executed before the RequestHandler
	Middlewares []fastalice.Constructor
	// DataHandlers used to receive requests
	DataHandler fastalice.Constructor
	// serviceClients holds references to service clients
	serviceClients []ServiceClient
}

func (route *Route) startRoute(router *fasthttprouter.Router, prefix string) {
	handler := startMiddlewares(route)
	path := prefix + route.Path
	switch route.method() {
	case DELETE:
		router.DELETE(path, handler)
	case GET:
		router.GET(path, handler)
	case POST:
		router.POST(path, handler)
	case PUT:
		router.PUT(path, handler)
	}
}

func startMiddlewares(route *Route) fasthttp.RequestHandler {
	chain := fastalice.New(route.Middlewares...)

	dh := route.defaultDataHandler
	if route.DataHandler != nil {
		dh = route.DataHandler
	}
	chain = chain.Append(dh)
	return chain.Then(route.defaultRequestHandler)
}

func (route *Route) addServiceClient(client ServiceClient) {
	if route.serviceClients == nil || len(route.serviceClients) == 0 {
		route.serviceClients = []ServiceClient{}
	}
	route.serviceClients = append(route.serviceClients, client)
}

func (route *Route) defaultRequestHandler(ctx *fasthttp.RequestCtx) {
	lotusCtx := Context{ctx, route.serviceClients}
	route.RequestHandler(&lotusCtx)
}

func (route *Route) defaultDataHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var m map[string]interface{}
		var err error

		body := ctx.PostBody()
		key := route.userValueKey()

		dataType := DataType(ctx.Request.Header.Peek("Content-Type"))

		if len(body) > 0 {
			if dataType == JSON {
				err = json.Unmarshal(body, &m)
			}
			if dataType == Binary {
				err = msgpack.Unmarshal(body, &m)
			}
			if len(m) > 0 {
				data := route.Data
				mapstructure.Decode(m, &data) // @TODO get rid of this once we have Generics
				ctx.SetUserValue(key, data)
			}
		}
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteString(err.Error())
			return
		}
		next(ctx)
	}
}

func replaceRouteMatches(m map[string][]string) func([]byte) []byte {
	return func(match []byte) []byte {
		key := match[1:]
		value := m[string(key)]
		return []byte(value[0])
	}
}
