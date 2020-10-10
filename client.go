package lotus

import (
	"github.com/valyala/fasthttp"
)

type ServiceClient struct {
	*ServiceContract
}

// Sends a request and returns a response and an error. The response must be released
func (sc *ServiceClient) SendRequest(routeContract RouteContract, payload ServiceRequest) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()

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
