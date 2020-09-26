package lotus

import (
	"github.com/brunvieira/fastalice"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// Method defines usable methods for endpoints
type Method string

const (
	// DELETE defines the DELETE type of request
	DELETE Method = fasthttp.MethodDelete
	// GET defines the GET type of request
	GET  = fasthttp.MethodGet
	// POST defines the POST type of request
	POST = fasthttp.MethodPost
	// PUT defines the PUT type of request
	PUT = fasthttp.MethodPut

	// Default values
	DefaultRouteMethod Method = GET
)

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
}

// Route is the definition of a route and what's used to initialize a route
type Route struct {
	*RouteContract
	// The endpoint is the path used to receive the request
	Endpoint fasthttp.RequestHandler
	// Middlewares functions executed before the Endpoint
	Middlewares []fastalice.Constructor
}

func (route *Route) startRoute(router *fasthttprouter.Router, prefix string) {
	handler := fastalice.New(route.Middlewares...).Then(route.Endpoint)
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

func (route *Route) method() Method {
	if len(route.Method) > 0  {
		return route.Method
	}
	return DefaultRouteMethod
}