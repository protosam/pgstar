package modmath

import (
	"github.com/protosam/pgstar/executor/modules"
	starlarkmath "go.starlark.net/lib/math"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "math"
)

type Module struct{}

func Constructor(loader modules.ModuleLoader) (modules.LocalizedModule, error) {
	return &Module{}, nil
}

func (module *Module) Exports() starlark.StringDict {
	return starlark.StringDict{
		"exports": starlarkstruct.FromStringDict(
			starlark.String(ModuleName),
			starlarkmath.Module.Members,
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}
