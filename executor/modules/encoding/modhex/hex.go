package modhex

import (
	"encoding/hex"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "hex"
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
				"encode": starlark.NewBuiltin("encode", hexEncode),
				"decode": starlark.NewBuiltin("decode", hexDecode),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func hexEncode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
		return starlark.None, err
	}

	return starlark.String(hex.EncodeToString([]byte(data))), nil
}

func hexDecode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
		return starlark.None, err
	}

	byteSlice, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return starlark.String(byteSlice), nil
}
