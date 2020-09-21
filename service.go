package lotus

import (
	"bytes"
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
	"net"
)

type protocol string

const (
	// Default Values

	// Default Host is the host used if not initialized.
	DefaultHost string = "localhost"
	// Default Port is the port used if not initialized.
	DefaultPort = "8080"
	// Default Domain is the domain used  if not initialized.
	DefaultDomain = "/"
	// Default Version is the version used if not initialized.
	DefaultVersion = "v0"

	// Protocols

	// HTTP protocol
	HTTP protocol = "http"
	// HTTPS protocol
	HTTPS = "https"
)

// Service providers are constructs able to start, stop and show it's current health (heartbeat)
type ServiceProvider interface {
	// Start inits the main process executed by the service
	Start()
	// Stop terminates all processes related to the service
	Stop() error
	// Status returns information about the health of the service
	Status() error
}

// Service is a Service Provider that starts itself and serves declared routes over a self created router
type Service struct {
	// private addr field. Holds a reference to the service addr
	addr string
	// private router field. Holds a reference to the router
	router *fasthttprouter.Router
	// private listener field. Holds a reference to the listener
	listener net.Listener
	// Name of the service. Used as a identifier for the service
	Label string
	// Description of the service. Intended to be a Short text describing the functionalities of the service
	Description string
	// Protocol is the protocol which the serve will use. The current options are: HTTP, HTTPS or TCP. Defaults to HTTP
	Protocol  protocol
	// Host of the service. Can be an IPV4 or IPV6 address. Defaults to "localhost"
	Host string
	// Domain of the service. The unique namespace which the service will be delivered upon. Defaults to "/"
	Domain string
	// Port of the service listener
	Port int
	// A version identifier for the service. Defaults to "v0"
	Version string
	// Routes store the routes served by this service
	Routes []*Route
}

/** Start inits the main process executed by the service. It first creates the internal router and then start a listener
for those routes and serve then on the specified Host, Port, Domain and Version. For a Service created with the following
values:

		service := &Service {
			Label: "ExampleService",
			Description: "Just a server to cover route url building",
			Protocol: lotus.HTTP,+
			Host: "myhost.com",
 */
func (service *Service) Start() {
	service.createRouter()
	service.startRoutes()
	service.startListening()
}

func (service *Service) createRouter() {
	if service.router == nil {
		service.router = fasthttprouter.New()
	}
}

func (service *Service) startRoutes()  {
	for _, route := range service.Routes {
		route.startRoute(service.router, service.suffix())
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

func (service *Service) address() string {
	var buffer bytes.Buffer

	buffer.WriteString(service.host())
	if service.port()[0] != ':' {
		buffer.WriteByte(':')
	}
	buffer.WriteString(service.port())
	return buffer.String()
}

func (service *Service) suffix() string {
	var buffer bytes.Buffer
	if service.domain()[0] != '/' {
		buffer.WriteByte('/')
	}
	buffer.WriteString(service.domain())
	if service.version()[0] != '/' {
		buffer.WriteByte('/')
	}
	buffer.WriteString(service.version())
	return buffer.String()
}

func (service *Service) apiHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		service.router.Handler(ctx)
	}
}

func (service *Service) protocol() string {
	if service.Protocol != "" {
		return string(service.Protocol)
	}
	return string(HTTP)
}

func (service *Service) host() string {
	if service.Host != "" {
		return service.Host
	}
	return DefaultHost
}

func (service *Service) port() string {
	if service.Port != 0 {
		return fmt.Sprintf("%d", service.Port)
	}
	return DefaultPort
}

func (service *Service) domain() string {
	if service.Domain != "" {
		return service.Domain
	}
	return DefaultDomain
}

func (service *Service) version() string {
	if service.Version != "" {
		return service.Version
	}
	return DefaultVersion
}

func (service *Service) Stop() error {
	err := service.listener.Close()
	if err != nil {
		log.Printf("Failed to stop service: %s. Error: %s", service.Label, err)
	}
	log.Println("Service", service.Label, "stopped...")
	return err
}

func (service *Service) Status() error {
	return nil
}


