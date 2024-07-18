package modyaml

import (
	"encoding/json"
	"fmt"

	"github.com/protosam/pgstar/executor/modules"
	"github.com/protosam/pgstar/executor/modules/starutils"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	yaml "gopkg.in/yaml.v3"
)

const (
	ModuleName = "yaml"
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
				"encode": starlark.NewBuiltin("yaml.encode", encode),
				"decode": starlark.NewBuiltin("yaml.decode", decode),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

// TODO: Build a yaml parser. The current solution is just a hack.
// For every call to encode/decode, the json package is used as a middle-ware
// to bridge the starlark json library with the yaml marshallers.

func encode(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &x); err != nil {
		return nil, err
	}

	jsonStr, err := starutils.StarlarkJsonEncoder(x)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	// convert jsonStr to data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	// convert data to yamlStr
	yamlStr, err := yaml.Marshal(data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	return starlark.Tuple{
		starlark.String(yamlStr),
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

	var data map[string]interface{}
	// decode yaml string into Go data
	if err := yaml.Unmarshal([]byte(s), &data); err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	// convert Go data to json
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	// decode json string with starlark json decoder
	decoded, err := starutils.StarlarkJsonDecoder(string(jsonStr), d)
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
