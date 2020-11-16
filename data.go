package lotus

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
	"reflect"
)

type DataType string

const (
	// JSON http header for json data
	JSON DataType = "application/json"

	// Binary http header for msgpack data. This is the default type used for intra-service communication
	Binary DataType = "application/msgpack"

	// Form http header for non-multipart form
	Form DataType = "application/x-www-form-urlencoded"

	// MultipartForm http header fo
	MultipartForm DataType = "multipart/form-data"
)

type Payload interface{} // This will be much better with generics

type ServiceRequest struct {
	RouteParams map[string]string
	QueryParams map[string]string
	Body        Payload
	DataType    DataType
}

type Context struct {
	*fasthttp.RequestCtx
	ServiceClients []ServiceClient
}

func (ctx *Context) Payload() (Payload, error) {
	if payload, ok := ctx.UserValue(DefaultKey).(interface{}); ok {
		return payload, nil
	}
	return nil, errors.New("fail to convert payload")
}

func (ctx *Context) ServiceClient(sub ServiceContract) *ServiceClient {
	for _, c := range ctx.ServiceClients {
		if c.Label == sub.Label {
			return &c
		}
	}
	return nil
}

type RequestHandler func(ctx *Context)

func dataToUrlValues(data interface{}) (form url.Values, err error) {
	form = map[string][]string{}
	iValue := reflect.ValueOf(data)
	for i := 0; i < iValue.NumField(); i++ {
		k := iValue.Type().Field(i).Name
		v := iValue.Field(i)

		if form[k] == nil {
			form[k] = []string{}
		}

		switch v.Kind() {
		case reflect.Slice:
			v2 := reflect.ValueOf(v.Interface())
			for j := 0; j < v2.Len(); j++ {
				f2 := v2.Index(j)
				form[k] = append(form[k], fmt.Sprint(f2))
			}
		default:
			form[k] = append(form[k], fmt.Sprint(v))
		}
	}
	return form, err
}
