package executor

import (
	"fmt"
	"log"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func PrintLog(thread *starlark.Thread, msg string) {
	log.Printf("%s: %s", thread.Name, msg)
}

func MakeStruct(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	argCount := args.Len()
	if argCount < 3 {
		return starlark.None, fmt.Errorf("%s expects at least 3 arguments", fn.Name())
	}

	if (argCount-1)%2 != 0 {
		return starlark.None, fmt.Errorf("%s expects matching pairs to build a struct", fn.Name())
	}

	provider := args.Index(0)

	var fields []starlark.Tuple
	for i := 1; i < argCount; i = i + 2 {
		fieldName := args.Index(i)
		fieldValue := args.Index(i + 1)
		fields = append(fields, starlark.Tuple{fieldName, fieldValue})
	}

	return starlarkstruct.FromKeywords(provider, fields), nil
}
