package modhttp

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName      = "http"
	StateNameReader = "http/reader"
	StateNameWriter = "http/writer"
)

type Module struct {
	w http.ResponseWriter
	r *http.Request
}

func Constructor(loader modules.ModuleLoader) (modules.LocalizedModule, error) {
	var w *http.ResponseWriter
	var r *http.Request
	if err := loader.GetState(StateNameWriter, &w); err != nil {
		return nil, err
	}
	if err := loader.GetState(StateNameReader, &r); err != nil {
		return nil, err
	}
	return &Module{
		w: *w,
		r: r,
	}, nil
}

func (module *Module) Exports() starlark.StringDict {
	return starlark.StringDict{
		"exports": starlarkstruct.FromStringDict(
			starlark.String(ModuleName),
			starlark.StringDict{
				"post":       starlark.NewBuiltin("http.post", module.HTTPPostFn),
				"query":      starlark.NewBuiltin("http.query", module.HTTPQueryFn),
				"vars":       starlark.NewBuiltin("http.vars", module.HTTPVarsFn),
				"write":      starlark.NewBuiltin("http.write", module.HTTPWriterFn),
				"setHeader":  starlark.NewBuiltin("http.setHeader", module.HTTPSetHeaderFn),
				"location":   starlark.NewBuiltin("http.location", module.HTTPLocationFn),
				"setCookie":  starlark.NewBuiltin("http.setCookie", module.HTTPSetCookieFn),
				"method":     starlark.NewBuiltin("http.method", module.HTTPMethodFn),
				"headers":    starlark.NewBuiltin("http.headers", module.HTTPHeadersFn),
				"cookies":    starlark.NewBuiltin("http.cookies", module.HTTPCookiesFn),
				"remoteAddr": starlark.NewBuiltin("http.remoteAddr", module.HTTPRemoteAddrFn),
				"host":       starlark.NewBuiltin("http.host", module.HTTPHostFn),
				"protocol":   starlark.NewBuiltin("http.protocol", module.HTTPProtocolFn),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func (module *Module) HTTPWriterFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var statuscode int
	var data starlark.Value
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "statuscode", &statuscode, "data", &data); err != nil {
		return starlark.None, err
	}

	module.w.WriteHeader(statuscode)

	sljson, err := MarshalValueToJson(data)
	if err != nil {
		return starlark.None, err
	}
	resp, _ := starlark.AsString(sljson)
	module.w.Write([]byte(resp))
	return starlark.None, modules.ErrEarlyExit

}

func (module *Module) HTTPHostFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	return starlark.String(module.r.Host), nil

}

func (module *Module) HTTPProtocolFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}
	protocol := "http"
	if module.r.TLS != nil {
		protocol = "https"
	}
	return starlark.String(protocol), nil
}

func (module *Module) HTTPLocationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var statuscode int
	var destination string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "statuscode", &statuscode, "destination", &destination); err != nil {
		return starlark.None, err
	}

	module.w.Header().Set("Location", destination)
	module.w.WriteHeader(statuscode)
	return starlark.None, modules.ErrEarlyExit

}

func (module *Module) HTTPSetCookieFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	cookie := http.Cookie{}
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 2, &cookie.Name, &cookie.Value, &cookie.Expires, &cookie.Path, &cookie.Domain, &cookie.Secure, &cookie.HttpOnly); err != nil {
		return starlark.None, err
	}
	http.SetCookie(module.w, &cookie)
	return starlark.None, nil
}

func (module *Module) HTTPCookiesFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}
	request_cookies := module.r.Cookies()
	cookies := starlark.NewDict(0)

	for i := range request_cookies {
		request_cookie := request_cookies[i]
		cookie := starlark.NewDict(0)
		name := starlark.String(request_cookie.Name)
		// build the cookie object
		cookie.SetKey(starlark.String("name"), name)
		cookie.SetKey(starlark.String("value"), starlark.String(request_cookie.Value))
		cookie.SetKey(starlark.String("expires"), starlark.String(request_cookie.Expires.String()))
		cookie.SetKey(starlark.String("path"), starlark.String(request_cookie.Path))
		cookie.SetKey(starlark.String("domain"), starlark.String(request_cookie.Domain))
		cookie.SetKey(starlark.String("secure"), starlark.Bool(request_cookie.Secure))
		cookie.SetKey(starlark.String("httponly"), starlark.Bool(request_cookie.HttpOnly))

		// append cookie to the dict of cookies
		cookies.SetKey(name, cookie)
	}
	return cookies, nil
}

func (module *Module) HTTPMethodFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	return starlark.String(module.r.Method), nil
}

func (module *Module) HTTPRemoteAddrFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}
	ip := module.r.RemoteAddr
	if strings.Contains(ip, ":") {
		// If the address contains a port (e.g., "192.168.1.1:12345"), strip it
		ip = ip[:strings.LastIndex(ip, ":")]
	}
	return starlark.String(ip), nil
}

func (module *Module) HTTPSetHeaderFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, value string
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 2, &key, &value); err != nil {
		return starlark.None, err
	}

	module.w.Header().Set(key, value)
	return starlark.None, nil
}

func (module *Module) HTTPHeadersFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	headers := starlark.NewDict(0)
	for key := range module.r.Header {
		values := starlark.NewList(nil)
		for i := range module.r.Header[key] {
			values.Append(starlark.String(module.r.Header[key][i]))
		}
		headers.SetKey(starlark.String(key), values)
	}
	return headers, nil
}

func (module *Module) HTTPPostFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	switch module.r.Header.Get("Content-Type") {
	case "application/json":
		rawpost, _ := io.ReadAll(module.r.Body)
		postdata, err := UnmarshalJsonToValue(string(rawpost))
		if err != nil {
			return starlark.None, errors.New("invalid json request")
		}
		return postdata, nil
	case "multipart/form-data":
		// TODO: Implement
		log.Printf("multipart form data is not supported")
		return starlark.None, errors.New("multipart form data is not supported")
	default:
		err := module.r.ParseForm()
		if err != nil {
			return starlark.None, errors.New("failed to parse form data")
		}

		postdata := starlark.NewDict(0)
		for key := range module.r.PostForm {
			values := starlark.NewList(nil)
			for i := range module.r.PostForm[key] {
				values.Append(starlark.String(module.r.PostForm[key][i]))
			}
			postdata.SetKey(starlark.String(key), values)
		}
		return postdata, nil
	}
}

func (module *Module) HTTPQueryFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}
	queryvalues := module.r.URL.Query()

	querydata := starlark.NewDict(0)
	for key := range queryvalues {
		values := starlark.NewList(nil)
		for i := range queryvalues[key] {
			values.Append(starlark.String(queryvalues[key][i]))
		}
		querydata.SetKey(starlark.String(key), values)
	}
	return querydata, nil
}

func (module *Module) HTTPVarsFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	vars := starlark.NewDict(0)
	for key, value := range mux.Vars(module.r) {
		vars.SetKey(starlark.String(key), starlark.String(value))
	}
	return vars, nil
}
