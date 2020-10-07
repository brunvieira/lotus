package lotus

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
)

type DataType int

const (
	Binary DataType = 1
	JSON            = 2
	Form            = 3
)

type Payload map[string]interface{}

type ServiceRequest struct {
	RouteParams map[string]string
	QueryParams map[string]string
	Body        Payload
	DataType    DataType
}

type Context struct {
	*fasthttp.RequestCtx
	ServiceClients []*ServiceClient
}

func (ctx *Context) Payload() (Payload, error) {
	if payload, ok := ctx.UserValue(DefaultKey).(map[string]interface{}); ok {
		return payload, nil
	}
	return nil, errors.New("fail to convert payload")
}

type RequestHandler func(ctx *Context)

func payloadBodyToUrlValues(payload ServiceRequest) (form url.Values, err error) {
	b, err := json.Marshal(payload.Body)
	if err != nil {
		return form, err
	}
	form = map[string][]string{}
	var result Payload
	err = json.Unmarshal(b, &result)
	if err != nil {
		return form, err
	}
	for k, v := range result {
		if form[k] == nil {
			form[k] = []string{}
		}
		switch v := v.(type) {
		case []interface{}:
			for _, u := range v {
				if u, ok := u.(string); ok {
					form[k] = append(form[k], u)
				} else {
					err = errors.New(fmt.Sprintf("Could not convert: %+v of type %s to string", u, v))
				}
			}
		default:
			form[k] = append(form[k], fmt.Sprint(v))
		}
	}
	return form, err
}
