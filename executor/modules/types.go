package modules

import (
	"errors"

	"go.starlark.net/starlark"
)

var ErrEarlyExit = errors.New("EXIT CALLED BY SCRIPT")

// Used to expose a module loader to modules for consumption
type ModuleLoader interface {
	SetState(string, interface{}) error
	GetState(string, interface{}) error
	GetThreadName() string
}

// Used by module loaders to orchestrate the use of a module
type LocalizedModule interface {
	Name() string
	Exports() starlark.StringDict
	Destroy(ModuleLoader) error
}

type ModuleExporterFn func(loader ModuleLoader) (LocalizedModule, error)
