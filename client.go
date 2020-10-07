package lotus

import (
	"errors"
	"github.com/valyala/fasthttp"
)

type ServiceClient struct {
	*ServiceContract
}

// Sends a request and returns a response and an error. The response must be released
func (sc *ServiceClient) SendRequest(routeLabel string, payload ServiceRequest) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()

	routeContract := sc.RouteContractByLabel(routeLabel)
	if routeContract == nil {
		return resp, errors.New("route not found")
	}

	url, err := sc.RouteUrl(routeContract.Label)
	if err != nil {
		return resp, err
	}

	req.SetRequestURI(url)
	err = routeContract.prepareRequest(req, payload)
	if err != nil {
		return resp, err
	}

	err = fasthttp.Do(req, resp)
	return resp, err
}
