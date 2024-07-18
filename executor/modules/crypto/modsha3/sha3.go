package modsha3

import (
	"hash"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"golang.org/x/crypto/sha3"
)

const (
	ModuleName = "sha3"
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
				"sum256": starlark.NewBuiltin("sha3.sum256", hashsum(sha3.New256)),
				"sum384": starlark.NewBuiltin("sha3.sum384", hashsum(sha3.New384)),
				"sum512": starlark.NewBuiltin("sha3.sum512", hashsum(sha3.New512)),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func hashsum(newHasherFn func() hash.Hash) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data string
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data); err != nil {
			return starlark.None, err
		}

		hasher := newHasherFn()

		if _, err := hasher.Write([]byte(data)); err != nil {
			return nil, err
		}

		return starlark.String(hasher.Sum(nil)), nil
	}
}
