package modbase64

import (
	"encoding/base64"
	"fmt"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "base64"
)

type Module struct{}

func Constructor(loader modules.ModuleLoader) (modules.LocalizedModule, error) {
	return &Module{}, nil
}

func (module *Module) Exports() starlark.StringDict {
	return starlark.StringDict{
		"exports": starlarkstruct.FromStringDict(
			starlark.String(ModuleName),
			starlark.StringDict{
				"encode": starlark.NewBuiltin("encode", base64Encode),
				"decode": starlark.NewBuiltin("decode", base64Decode),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func base64Encode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
		return starlark.None, err
	}

	return starlark.String(base64.StdEncoding.EncodeToString([]byte(data))), nil
}

func base64Decode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
		return starlark.None, err
	}

	byteSlice, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	return starlark.Tuple{
		starlark.String(byteSlice),
		starlark.None,
	}, nil
}
