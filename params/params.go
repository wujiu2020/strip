package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/teapots/teapot"
)

var DefaultMaxMemory int64 = 32 << 20 /* 32 MB */

// Params provides a unified view of the request params.
// Includes:
// - URL query string
// - Form values
// - File uploads
//
// Warning: param maps other than Values may be nil if there were none.
type Params struct {
	url.Values // A unified view of all the individual param maps below.

	// Set by the router
	Route teapot.RouteInfo // Parameters extracted from the route,  e.g. /customers/{id}

	// Set by the ParamsFilter
	Query url.Values // Parameters from the query string, e.g. /index?limit=10
	Form  url.Values // Parameters from the request body.

	Files    map[string][]*multipart.FileHeader // Files uploaded in a multipart form
	tmpFiles []*os.File                         // Temp files used during the request.

	req *http.Request
	log teapot.Logger
}

func ParseParams(params *Params, maxMemory int64) error {
	req := params.req

	params.Query = req.URL.Query()

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/html"
	} else {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}

	// Parse the body depending on the content type.
	switch contentType {
	case "application/x-www-form-urlencoded":
		// Typical form.
		if err := req.ParseForm(); err != nil {
			return err
		} else {
			params.Form = req.Form
		}

	case "multipart/form-data":
		// Multipart form.
		if maxMemory <= 0 {
			maxMemory = DefaultMaxMemory
		}
		if err := req.ParseMultipartForm(maxMemory); err != nil {
			return err
		} else {
			params.Form = req.MultipartForm.Value
			params.Files = req.MultipartForm.File
		}
	}

	params.Values = params.calcValues()
	return nil
}

// Bind looks for the named parameter, converts it to the requested type, and
// writes it into "dest", which must be settable.  If the value can not be
// parsed, "dest" is set to the zero value.
func (p *Params) Bind(dest interface{}, name string) {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		panic("non-pointer passed to Bind: " + name)
	}
	value = value.Elem()
	if !value.CanSet() {
		panic("non-settable variable passed to Bind: " + name)
	}
	value.Set(Bind(p, name, value.Type()))
}

func (p *Params) BindJsonBody(dest interface{}, userNumber ...bool) (err error) {
	body, err := ioutil.ReadAll(p.req.Body)
	if err != nil {
		return
	}

	if len(body) == 0 {
		return
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	if len(userNumber) > 0 && userNumber[0] {
		dec.UseNumber()
	}
	err = dec.Decode(dest)
	if err != nil {
		// forbid large request attack
		if len(body) > 512 {
			body = body[:512]
		}
		err = fmt.Errorf("decode body err: %v, %v", err, string(body[:len(body)]))
	}
	return
}

// can use url.Query / url.Values, parse to json
// can use url.Query / url.Values, parse to json
func (p *Params) BindValuesToStruct(dest interface{}) {
	pointerMap := make(map[uintptr]bool)
	val := reflect.ValueOf(dest)
	elm := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr && elm.Kind() != reflect.Struct {
		panic("need ptr of struct")
	}
	p.bindValuesToStruct(elm, pointerMap)
}

func (p *Params) bindValuesToStruct(elm reflect.Value, pointerMap map[uintptr]bool) (exits bool) {
	typ := elm.Type()

	for i := 0; i < elm.NumField(); i++ {
		field := elm.Field(i)
		ftyp := typ.Field(i)

		if ftyp.PkgPath != "" && !ftyp.Anonymous { // skip unexport
			continue
		}

		tag := ftyp.Tag.Get("param")
		if tag == "-" {
			continue
		}

		name := tag
		if name == "" {
			name = ftyp.Name
		}

		if idx := strings.Index(tag, ","); idx != -1 {
			name = tag[:idx]
		}

		// struct recursion
		if ftyp.Anonymous {
			var inf reflect.Value

			if field.Kind() == reflect.Ptr {
				pointer := field.Pointer()
				if pointerMap[pointer] {
					continue
				}

				var newField reflect.Value
				if field.IsNil() {
					newField = reflect.New(ftyp.Type.Elem())
				} else {
					newField = field
				}

				inf = reflect.Indirect(newField)
				if inf.Kind() != reflect.Struct {
					continue
				}

				if pointer > 0 {
					// save pointer
					pointerMap[pointer] = true
				}
				if ok := p.bindValuesToStruct(inf, pointerMap); ok {
					field.Set(newField)
				}

			} else if field.Kind() == reflect.Struct {
				inf = field
				p.bindValuesToStruct(inf, pointerMap)
			} else {
				continue
			}

		} else {
			if !exits {
				if vals, ok := p.Values[name]; ok && len(vals) > 0 {
					exits = true
				}
			}
			paramValue := Bind(p, name, field.Type())
			if paramValue.Type().ConvertibleTo(field.Type()) {
				field.Set(paramValue.Convert(field.Type()))
			}
		}
	}
	return
}

// calcValues returns a unified view of the component param maps.
func (p *Params) calcValues() url.Values {
	numParams := len(p.Query) + len(p.Route.Values) + len(p.Form)

	// If there were no params, return an empty map.
	if numParams == 0 {
		return make(url.Values, 0)
	}

	// If only one of the param sources has anything, return that directly.
	switch numParams {
	case len(p.Query):
		return p.Query
	case len(p.Route.Values):
		return p.Route.Values
	case len(p.Form):
		return p.Form
	}

	// Copy everything into the same map.
	values := make(url.Values, numParams)
	for k, v := range p.Query {
		values[k] = append(values[k], v...)
	}
	for k, v := range p.Route.Values {
		values[k] = append(values[k], v...)
	}
	for k, v := range p.Form {
		values[k] = append(values[k], v...)
	}
	return values
}

func ParamsParser() interface{} {
	var maxMemory int64

	return func(req *http.Request, ctx teapot.Context, log teapot.Logger, config *teapot.Config) *Params {
		params := new(Params)
		params.req = req
		params.log = log

		var routeInfo *teapot.RouteInfo
		ctx.Find(&routeInfo, "")

		if routeInfo != nil {
			params.Route = *routeInfo
		}

		if maxMemory == 0 {
			config.Bind(&maxMemory, "max_memory")
			if maxMemory == 0 {
				maxMemory = DefaultMaxMemory
			}
		}

		err := ParseParams(params, maxMemory)
		if err != nil {
			log.Notice("parse params error", err)
		}

		// Clean up from the request.
		defer func() {
			// Delete temp files.
			if req.MultipartForm != nil {
				err := req.MultipartForm.RemoveAll()
				if err != nil {
					log.Notice("error removing temporary files:", err)
				}
			}

			for _, tmpFile := range params.tmpFiles {
				err := os.Remove(tmpFile.Name())
				if err != nil {
					log.Notice("could not remove upload temp file:", err)
				}
			}
		}()

		return params
	}
}

func (p *Params) RealIp() string {
	return RealIp(p.req)
}

func (p *Params) RealHost() string {
	return RealHost(p.req)
}

func (p *Params) RealProto() string {
	return RealProto(p.req)
}

func (p *Params) RealURI() string {
	return RealURI(p.req)
}

func (p *Params) RealURL() (*url.URL, error) {
	return RealURL(p.req)
}
