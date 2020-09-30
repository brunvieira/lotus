package echo_service

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func echo(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, string(ctx.RequestURI()))
}
