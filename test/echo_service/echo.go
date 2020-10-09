package echo_service

import (
	"fmt"
	"github.com/brunvieira/lotus"
	"github.com/brunvieira/lotus/test/contract"
	"github.com/valyala/fasthttp"
)

func echo(ctx *lotus.Context) {
	fmt.Fprint(ctx, string(ctx.RequestURI()))
}
func randomString(ctx *lotus.Context) {
	c := ctx.ServiceClient(contract.RandomStringsServiceContract)
	resp, err := c.SendRequest(contract.RandomStringsRouteContract, lotus.ServiceRequest{})
	defer fasthttp.ReleaseResponse(resp)
	if err == nil {
		ctx.Write(resp.Body())
	}
}
