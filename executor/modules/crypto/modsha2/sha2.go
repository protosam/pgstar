package modsha2

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "sha2"
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
				"sum256": starlark.NewBuiltin("sha2.sum256", hashsum(sha256.New())),
				"sum512": starlark.NewBuiltin("sha2.sum512", hashsum(sha512.New())),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func hashsum(hasher hash.Hash) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data string
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
			return starlark.None, err
		}

		if _, err := hasher.Write([]byte(data)); err != nil {
			return nil, err
		}

		return starlark.String(hasher.Sum(nil)), nil
	}
}
