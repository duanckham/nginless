package nginless

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"go.uber.org/zap"
)

// doCall ...
func (n *Nginless) doCall(d *D, parameters []interface{}) *D {
	actionFile, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.tengo", n.actions, parameters[0].(string)))
	if err != nil {
		d.returnInternalServerError()
		return d
	}

	actionFile = append(actionFile, []byte("\nhandle(BUILDIN_REQ, BUILDIN_RES)")...)
	script := tengo.NewScript(actionFile)

	script.Add("BUILDIN_REQ", createReqModule(d))
	script.Add("BUILDIN_RES", createResModule(d))

	script.Add("fetch", fetchFunc)

	script.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	_, err = script.RunContext(context.Background())
	if err != nil {
		n.logger.Error(".doCall got error", zap.Error(err))
		d.returnInternalServerError()
	}

	return d
}

// createReqModule ...
func createReqModule(d *D) map[string]tengo.Object {
	// Process Query.
	queries := map[string]tengo.Object{}

	for k, q := range d.req.URL.Query() {
		for _, v := range q {
			queries[k] = &tengo.String{Value: v}
			break
		}
	}

	// Process headers.
	headers := map[string]tengo.Object{}

	for k, h := range d.req.Header {
		for _, v := range h {
			headers[k] = &tengo.String{Value: v}
			break
		}
	}

	return map[string]tengo.Object{
		"method":  &tengo.String{Value: d.req.Method},
		"host":    &tengo.String{Value: d.req.Host},
		"path":    &tengo.String{Value: d.req.URL.Path},
		"queries": &tengo.Map{Value: queries},
		"headers": &tengo.Map{Value: headers},
	}
}

// createResModule ...
func createResModule(d *D) map[string]tengo.Object {
	return map[string]tengo.Object{
		"html": &tengo.UserFunction{
			Name: "html",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					d.returnInternalServerError()
					d.done = true
					return nil, errors.New("Parameter cannot be empty")
				}

				arg := args[0]

				d.res.Header().Set("Content-Type", "text/html")
				d.done = true

				if arg.TypeName() == "string" {
					d.res.Write([]byte(arg.(*tengo.String).Value))
				} else {
					d.res.Write([]byte(arg.String()))
				}

				return nil, nil
			}},
		"text": &tengo.UserFunction{
			Name: "text",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					d.returnInternalServerError()
					d.done = true
					return nil, errors.New("Parameter cannot be empty")
				}

				arg := args[0]

				d.res.Header().Set("Content-Type", "text/plain")
				d.done = true

				if arg.TypeName() == "string" {
					d.res.Write([]byte(arg.(*tengo.String).Value))
				} else {
					d.res.Write([]byte(arg.String()))
				}

				return nil, nil
			}},
		"json": &tengo.UserFunction{
			Name: "json",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					d.returnInternalServerError()
					d.done = true
					return nil, errors.New("Parameter cannot be empty")
				}

				arg := args[0]

				if arg.TypeName() != "map" {
					d.returnInternalServerError()
					d.done = true
					return nil, errors.New("Parameter should be a map")
				}

				v, _ := tengo.NewVariable("", arg)
				b, _ := json.Marshal(v.Map())

				d.res.Header().Set("Content-Type", "application/json")
				d.res.Write(b)
				d.done = true

				return nil, nil
			}},
	}
}

var fetchFunc = &tengo.UserFunction{
	Name: "fetch",
	Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) == 0 {
			return nil, nil
		}

		var (
			reader  interface{} = nil
			arg                 = args[0]
			uri                 = ""
			method              = "GET"
			headers             = map[string]string{}
		)

		switch arg.TypeName() {
		case "string":
			uri = arg.(*tengo.String).Value
		case "map":
			m := arg.(*tengo.Map).Value

			if o, ok := m["url"]; ok {
				uri = o.(*tengo.String).Value
			}

			if o, ok := m["method"]; ok {
				method = strings.ToUpper(o.(*tengo.String).Value)
			}

			if o, ok := m["data"]; ok {
				switch o.TypeName() {
				case "string":
					reader = strings.NewReader(o.(*tengo.String).Value)
				case "map":
					v, _ := tengo.NewVariable("", o)
					b, _ := json.Marshal(v.Map())
					reader = bytes.NewReader(b)
				}
			}

			if o, ok := m["headers"]; ok {
				if o.TypeName() == "map" {
					for k, item := range o.(*tengo.Map).Value {
						switch item.TypeName() {
						case "string":
							headers[k] = item.(*tengo.String).Value
						default:
							headers[k] = item.String()
						}
					}
				}
			}

			if v, ok := m["json"]; ok {
				if v.TypeName() == "bool" && !v.(*tengo.Bool).IsFalsy() {
					headers["Content-Type"] = "application/json"
				}
			}
		}

		req, err := http.NewRequest(method, uri, reader.(io.Reader))
		if err != nil {
			return nil, err
		}

		// Copy request headers.
		for k, item := range headers {
			req.Header.Add(k, item)
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, _ := ioutil.ReadAll(res.Body)

		return &tengo.Map{
			Value: map[string]tengo.Object{
				"body": &tengo.String{
					Value: string(body),
				},
			},
		}, nil
	}}
