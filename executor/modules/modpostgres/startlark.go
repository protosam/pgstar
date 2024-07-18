package modpostgres

import (
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
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

func parseRow(rows pgx.Rows, fields []string) *starlark.Dict {
	values, _ := rows.Values()
	// slvals := starlark.NewList(nil)
	slvals := starlark.NewDict(0)
	for idx := range values {
		switch values[idx].(type) {
		case int:
			slvals.SetKey(starlark.String(fields[idx]), starlark.MakeInt(values[idx].(int)))
		case string:
			slvals.SetKey(starlark.String(fields[idx]), starlark.String(values[idx].(string)))
		case float32:
			slvals.SetKey(starlark.String(fields[idx]), starlark.Float(values[idx].(float32)))
		case float64:
			slvals.SetKey(starlark.String(fields[idx]), starlark.Float(values[idx].(float64)))
		case bool:
			slvals.SetKey(starlark.String(fields[idx]), starlark.Bool(values[idx].(bool)))
		case time.Time:
			val := values[idx].(time.Time)
			slvals.SetKey(starlark.String(fields[idx]), starlark.String(val.String()))
		case [16]uint8:
			val := values[idx].([16]uint8)
			id, _ := uuid.FromBytes(val[:])
			slvals.SetKey(starlark.String(fields[idx]), starlark.String(id.String()))
		case fmt.Stringer:
			slvals.SetKey(starlark.String(fields[idx]), starlark.String(fmt.Sprintf("%s", values[idx])))
			log.Printf("failed back to Sprintf for type in db.rows.next(): %#v", values[idx])
		case nil:
			slvals.SetKey(starlark.String(fields[idx]), starlark.None)
		default:
			// slvals.Append(starlark.None)
			slvals.SetKey(starlark.String(fields[idx]), starlark.None)
			log.Printf("This type is not handled yet: %#v", values[idx])
		}
	}
	return slvals
}
