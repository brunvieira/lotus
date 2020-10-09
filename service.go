package lotus

import (
	"errors"
	"fmt"
	"github.com/brunvieira/fastalice"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"strings"
)

type protocol string

const (
	// Default Values

	// Default Host is the host used if not initialized.
	DefaultHost string = "localhost"
	// Default Port is the port used if not initialized.
	DefaultPort = "8080"
	// Default Namespace is the namespace used  if not initialized.
	DefaultNamespace = ""
	// Default Version is the version used if not initialized.
	DefaultVersion = "v0"

	// Protocols

	// HTTP protocol
	HTTP protocol = "http"
	// HTTPS protocol
	HTTPS = "https"

	// Errors

	// RouteNotFoundError error when the service can't find a route
	RouteNotFoundError string = "route not found"
)

type ServiceStatus struct {
	IsRunning        bool
	Address          net.Addr
	RegisteredRoutes int
}

// Service providers are constructs able to start, stop and show it's current health (heartbeat)
type ServiceProvider interface {
	// Start inits the main process executed by the service
	Start()
	// Stop terminates all processes related to the service
	Stop() error
	// Status returns information about the health of the service
	Status() *ServiceStatus
}

// ServiceContract holds the Contract description of a service
type ServiceContract struct {
	// Name of the service. Used as a identifier for the service
	Label string
	// Description of the service. Intended to be a Short text describing the functionalities of the service
	Description string
	// Protocol is the protocol which the serve will use. The current options are: HTTP, HTTPS or TCP. Defaults to HTTP
	Protocol protocol
	// Host of the service. Can be an IPV4 or IPV6 address. Defaults to "localhost"
	Host string
	// Namespace of the service. The unique namespace which the service will be delivered upon. Defaults to "/"
	Namespace string
	// Port of the service listener
	Port int
	// A version identifier for the service. Defaults to "v0"
	Version string
	// RoutesContracts is an array of RouteContract used to define the Routes on the contract
	RoutesContracts []RouteContract
	// Subscriptions is an array of other ServiceContract which this service will need to communicate. It's used to create
	// ServiceClient
	Subscriptions []ServiceContract
}

// RouteContractByLabel returns the route contract for the given label
func (sc *ServiceContract) RouteContractByLabel(label string) *RouteContract {
	for _, r := range sc.RoutesContracts {
		if r.Label == label {
			return &r
		}
	}
	return nil
}

func (sc *ServiceContract) host() string {
	if sc.Host != "" {
		return sc.Host
	}
	return DefaultHost
}

func (sc *ServiceContract) port() string {
	if sc.Port != 0 {
		return fmt.Sprintf("%d", sc.Port)
	}
	return DefaultPort
}

func (sc *ServiceContract) address() string {
	var builder strings.Builder

	builder.WriteString(sc.host())
	if sc.port()[0] != ':' {
		builder.WriteByte(':')
	}
	builder.WriteString(sc.port())
	return builder.String()
}

func (sc *ServiceContract) Suffix() string {
	var builder strings.Builder
	if len(sc.namespace()) > 0 && sc.namespace()[0] != '/' {
		builder.WriteByte('/')
	}
	builder.WriteString(sc.namespace())
	if sc.version()[0] != '/' {
		builder.WriteByte('/')
	}
	builder.WriteString(sc.version())
	return builder.String()
}

func (sc *ServiceContract) protocol() string {
	if sc.Protocol != "" {
		return string(sc.Protocol)
	}
	return string(HTTP)
}

func (sc *ServiceContract) namespace() string {
	if sc.Namespace != "" {
		return sc.Namespace
	}
	return DefaultNamespace
}

func (sc *ServiceContract) version() string {
	if sc.Version != "" {
		return sc.Version
	}
	return DefaultVersion
}

func (sc *ServiceContract) routeByLabel(label string) (*RouteContract, error) {
	for _, r := range sc.RoutesContracts {
		if r.Label == label {
			return &r, nil
		}
	}
	return nil, errors.New(RouteNotFoundError)
}

func (sc *ServiceContract) RouteUrl(label string) (string, error) {
	r, err := sc.routeByLabel(label)
	if err != nil {
		return "", err
	}
	w := strings.Builder{}
	w.Grow(len(sc.protocol()) + 3 + len(sc.address()) + len(sc.Suffix()) + len(r.Path))
	w.WriteString(sc.protocol())
	w.WriteString("://")
	w.WriteString(sc.address())
	w.WriteString(sc.Suffix())
	w.WriteString(r.Path)
	return w.String(), nil
}

// Service is a Service Provider that starts itself and serves declared routes over a self created router
type Service struct {
	*ServiceContract
	// private addr field. Holds a reference to the service addr
	addr string
	// private router field. Holds a reference to the router
	router *fasthttprouter.Router
	// private listener field. Holds a reference to the listener
	listener net.Listener
	// routes is an array of Route from the Service. They are validate against the RoutesContracts from the ServiceContract
	routes []*Route
	// serviceClients holds references to ServiceClients this service subscribe to
	serviceClients []ServiceClient
}

/** Start inits the main process executed by the service. It first creates the internal router and then start a listener
for those routes and serve then on the specified Host, Port, Namespace and Version. For a Service created with the following
values:

		service := &Service {
			Label: "ExampleService",
			Description: "Just a server to cover route url building",
			Protocol: lotus.HTTP,
			Host: "myhost.com",
			Namespace: "example",
		}
*/
func (service *Service) Start() {
	service.validateRoutes()
	service.createRouter()
	service.startServiceClients()
	service.startRoutes()
	service.startListening()
}

func (service *Service) Stop() error {
	if service.listener == nil {
		return errors.New("service connection not found")
	}
	service.listener.Close()
	log.Println("Service", service.Label, "stopped...")
	return nil
}

func (service *Service) Status() *ServiceStatus {
	if service.listener == nil {
		return &ServiceStatus{
			IsRunning:        false,
			Address:          nil,
			RegisteredRoutes: len(service.routes),
		}
	}
	return &ServiceStatus{
		IsRunning:        true,
		Address:          service.listener.Addr(),
		RegisteredRoutes: len(service.routes),
	}
}

// SetupRoute searches for a route contract identified by label, creates a Route, add to it the endpoint and the
// middlewares and put it on the Routes array returning the newly created routes
func (service *Service) SetupRoute(
	label string,
	endpoint RequestHandler,
	middlewares []fastalice.Constructor,
	dataHandler fastalice.Constructor,
) *Route {
	routeContract := service.routeContract(label)
	service.testRouteContractExists(routeContract, label)
	route := Route{
		routeContract,
		endpoint,
		middlewares,
		dataHandler,
		[]ServiceClient{},
	}
	service.AddRoute(&route)
	return &route
}

// AddRoute adds an already created route to the service routes array
func (service *Service) AddRoute(route *Route) {
	routes := append(service.routes, route)
	service.routes = routes
}

func (service *Service) validateRoutes() {
	routeDescriptions := service.RoutesContracts
	for _, desc := range routeDescriptions {
		var routeContract *RouteContract
		for _, c := range service.routes {
			if c.Label == desc.Label {
				routeContract = c.RouteContract
			}
		}
		service.testRouteContractExists(routeContract, desc.Label)
	}
}

func (service *Service) testRouteContractExists(routeContract *RouteContract, label string) {
	if routeContract == nil {
		panic("Route for " + label + " not found")
	}
}

func (service *Service) routeContract(label string) *RouteContract {
	for _, routeContract := range service.RoutesContracts {
		if routeContract.Label == label {
			return &routeContract
		}
	}
	return nil
}

func (service *Service) createRouter() {
	if service.router == nil {
		service.router = fasthttprouter.New()
	}
}

func (service *Service) startServiceClients() {
	if service.ServiceContract.Subscriptions == nil || len(service.ServiceContract.Subscriptions) == 0 {
		return
	}
	if service.serviceClients == nil || len(service.serviceClients) == 0 {
		service.serviceClients = []ServiceClient{}
	}
	for _, sub := range service.ServiceContract.Subscriptions {
		client := ServiceClient{&sub}
		service.serviceClients = append(service.serviceClients, client)
	}
}

func (service *Service) startRoutes() {
	for _, route := range service.routes {
		for _, client := range service.serviceClients {
			route.addServiceClient(client)
		}
		route.startRoute(service.router, service.Suffix())
	}
}

func (service *Service) startListening() {
	ln, err := net.Listen("tcp", service.address())
	if err != nil {
		panic(err)
	}
	service.listener = ln
	log.Printf("Serving at: %s", service.address())
	fasthttp.Serve(ln, service.router.Handler)
}
