package modregex

import (
	"regexp"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "regexp"
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
				"match": starlark.NewBuiltin("regex.match", match),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func match(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern, s starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pattern, "s", &s); err != nil {
		return nil, err
	}

	matched, err := regexp.MatchString(string(pattern), string(s))
	if err != nil {
		return nil, err
	}

	return starlark.Bool(matched), nil
}
