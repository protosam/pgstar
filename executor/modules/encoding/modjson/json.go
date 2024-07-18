package modjson

import (
	"fmt"

	"github.com/protosam/pgstar/executor/modules"
	"github.com/protosam/pgstar/executor/modules/starutils"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "json"
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
				"encode": starlark.NewBuiltin("json.encode", encode),
				"decode": starlark.NewBuiltin("json.decode", decode),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func encode(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &x); err != nil {
		return nil, err
	}

	jsonString, err := starutils.StarlarkJsonEncoder(x)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	return starlark.Tuple{
		starlark.String(jsonString),
		starlark.None,
	}, nil
}

func decode(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (v starlark.Value, err error) {
	var s string
	var d starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &s, "default?", &d); err != nil {
		return nil, err
	}
	if len(args) < 1 {
		// "x" parameter is positional only; UnpackArgs does not allow us to
		// directly express "def decode(x, *, default)"
		return nil, fmt.Errorf("%s: unexpected keyword argument x", b.Name())
	}

	decoded, err := starutils.StarlarkJsonDecoder(s, d)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	return starlark.Tuple{
		decoded,
		starlark.None,
	}, nil
}
