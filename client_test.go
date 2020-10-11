package lotus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"testing"
	"time"
)

type EchoPayload struct {
	Foo string
	Bar string
}

var (
	postEchoRouteContract = RouteContract{
		Label:       "PostEcho",
		Description: "A route that outputs the contents of it's body",
		Path:        "/echo",
		Method:      fasthttp.MethodPost,
		Data:        EchoPayload{},
	}
	serviceClientContract = ServiceContract{
		Label:     "DefaultEchoService",
		Host:      "",
		Namespace: "",
		Port:      10080,
		RoutesContracts: []RouteContract{
			SimpleEchoRouteContract,
			postEchoRouteContract,
		},
	}
	echoServiceClient = ServiceClient{
		&serviceClientContract,
	}
	defaultPayload = ServiceRequest{
		Body: EchoPayload{
			Foo: "foo",
			Bar: "bar",
		},
	}
	echoPayload = func(ctx *Context) {
		payload, err := ctx.Payload()
		if err != nil {
			panic(err)
		}
		ctx.WriteString(fmt.Sprint(payload))
	}
)

func TestSendRequest(t *testing.T) {
	service := Service{ServiceContract: &serviceClientContract}
	service.SetupRoute("SimpleEcho", echo, nil, nil)
	service.SetupRoute("PostEcho", echoPayload, nil, nil)

	go service.Start()
	defer service.Stop()

	time.Sleep(2 * time.Second)

	resp, err := echoServiceClient.SendRequest(postEchoRouteContract, defaultPayload)
	defer fasthttp.ReleaseResponse(resp)

	assert.Nil(t, err, "Sending a request must not return an error")
	assert.NotNil(t, resp, "Sending a request must return a non empty response")

	body := resp.Body()
	assert.NotNil(t, body, "Sending a request must return a non empty body")
	assert.Equal(t, fmt.Sprint(defaultPayload.Body), string(body), "Result payload must be equal sent payload")

}
