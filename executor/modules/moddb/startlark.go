package moddb

import (
	"log"

	"go.starlark.net/starlark"
)

func pgargsToStarlarkValue(_ *starlark.Thread, fn *starlark.Builtin, pgargs *starlark.List) []interface{} {
	params := []interface{}{}
	for i := 0; i < pgargs.Len(); i++ {
		param := pgargs.Index(i)
		switch param.Type() {
		case "bool":
			params = append(params, param.Truth())
		case "int":
			var paramToGoVal int
			starlark.AsInt(param, &paramToGoVal)
			params = append(params, paramToGoVal)
		case "float":
			paramToGoVal, _ := starlark.AsFloat(param)
			params = append(params, paramToGoVal)
		case "string":
			paramToGoVal, _ := starlark.AsString(param)
			params = append(params, paramToGoVal)
		default:
			paramToGoVal, _ := starlark.AsString(param)
			params = append(params, paramToGoVal)
			log.Printf("failed back to Sprintf for type in %s(): %#v", fn.Name(), param)
		}
	}

	return params
}
