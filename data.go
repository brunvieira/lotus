package lotus

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/vmihailenco/msgpack"
	"net/url"
	"regexp"
	"strings"
)

type DataType int


const (
	Binary DataType = 1
	JSON = 2
	Form = 3
	RouteParams = 4
	QueryParams = 5
)

// DataHandler is a interface for handling the exchange of data among services
type DataHandler interface {
	// ReceiveRequest is a middleware for handling data receiving. It unpacks the data and maps it's values to a
	// *fasthttp.RequestCtx using fasthttp.SetUserValues
	ReceiveRequest(req fasthttp.RequestHandler) fasthttp.RequestHandler

	// PrepareRouteRequest prepares the data to be sent. The data will be sent using the DataType parameter in the
	// following manner:
	//
	// Binary: Default type. The data will be encoded using MsgPack in the body of the request
	// JSON: The data will be sent as a JSON string in the body of the request
	// Form: The data will be sent encoded as url.Values in the body of the request
	// RouteParams: The data will replace endpoint params with the equally named field. Example, :documentId will be
	// replaced by the field documentId
	// QueryParams: The data will be sent encoded as url.Values as a query parameter of the request
	PrepareRouteRequest(*fasthttp.Request, *Route) error
}

// DataContract is the data contract description
type DataContract struct {
	Type DataType
	// Key is the key used on SetUserValue. Defaults to data
	Key string
}

// DefaultDataHandler is the Default data handler
type DefaultDataHandler struct {
	*DataContract
	Payload map[string]interface{}
}

func NewDataContract(dataType DataType) DataContract {
		return DataContract{Type: dataType}
}

func (d *DefaultDataHandler) ReceiveRequest(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		err := receiveRequestFn(d, ctx)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteString(err.Error())
			return
		}
		next(ctx)
	}
}

func receiveRequestFn(d *DefaultDataHandler, ctx *fasthttp.RequestCtx) error {
	output := d
	body := ctx.PostBody()
	key := userValueKey(d)

	switch d.Type {
	case Binary:
		if len(body) == 0 {
			return nil
		}
		err := msgpack.Unmarshal(body, &output)
		ctx.SetUserValue(key, output.Payload)
		return err
	case JSON:
		if len(body) == 0 {
			return nil
		}
		err := json.Unmarshal(body, &output)
		ctx.SetUserValue(key, output.Payload)
		return err
	case RouteParams:
		// handled by fasthttp.Router
		return nil
	case QueryParams:
		// handled by fasthttp. Use ctx.QueryArgs().Peek or ctx.QueryArgs().PeekMulti()
		return nil
	case Form:
		// handled by fasthttp
		return nil
	}
	return nil
}

var routerParamReg = regexp.MustCompile(`:[a-zA-Z0-9]*`)

func (d *DefaultDataHandler) PrepareRouteRequest(req *fasthttp.Request, route *Route) error {
	output := d

	method := string(route.Method)
	req.Header.SetMethod(method)

	switch d.Type {
	case Binary:
		b, err := msgpack.Marshal(output)
		req.SetBody(b)
		return err
	case JSON:
		b, err := json.Marshal(output)
		req.SetBody(b)
		return err
	case Form:
		m, err := dataModelToMap(output)
		if err != nil {
			return err
		}
		req.SetBodyString(m.Encode())
		return nil
	case RouteParams:
		m, err := dataModelToMap(output)
		if err != nil {
			return err
		}
		path := routerParamReg.ReplaceAllFunc([]byte(route.Path), replaceRouteMatches(m))
		uriStr := req.URI().String()
		newUri := strings.Replace(uriStr, route.Path, string(path), 1)
		req.SetRequestURI(newUri)
		return nil
	case QueryParams:
		m, err := dataModelToMap(output)
		if err != nil {
			return err
		}
		query := "?" + m.Encode()
		uriStr := req.URI().String()
		newUri := uriStr + query
		req.SetRequestURI(newUri)
		return nil
	}
	return nil
}

func replaceRouteMatches(m map[string][]string) func([]byte) []byte {
	return func(match []byte) []byte {
		key := match[1:]
		value := m[string(key)]
		return []byte(value[0])
	}
}

func dataModelToMap(d *DefaultDataHandler) (url.Values, error) {
	var modelMap map[string][]string
	b, err := json.Marshal(d)
	if err != nil {
		return modelMap, err
	}
	err = json.Unmarshal(b, &modelMap)
	return modelMap, err
}

func userValueKey(d *DefaultDataHandler) string {
	if k := d.Key; len(k) == 0 {
		return "data"
	}
	return d.Key
}