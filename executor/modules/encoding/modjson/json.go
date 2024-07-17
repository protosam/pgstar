package modjson

import (
	"github.com/protosam/pgstar/executor/modules"
	starlarkjson "go.starlark.net/lib/json"
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
				"encode": starlarkjson.Module.Members["encode"],
				"decode": starlarkjson.Module.Members["decode"],
				// "savepoints": starlark.NewBuiltin("db.savepoints", module.savepoints),
				// "query":      starlark.NewBuiltin("db.query", module.query),
				// "exec":       starlark.NewBuiltin("db.exec", module.exec),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}
