package executor

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
)

var ErrStatePointerRequired = errors.New("state values must be pointers")
var ErrStateNotFound = errors.New("state not found")

type ModuleLoader struct {
	mt              *ManagedThread
	rootdir         string
	loaded          []string
	state           map[string]interface{}
	localizedStates map[string]modules.LocalizedModule
}

// NewModuleLoader creates a root level ModuleLoader
func NewModuleLoader(mt *ManagedThread, rootdir, starfile string) *ModuleLoader {
	return &ModuleLoader{
		mt:              mt,
		rootdir:         rootdir,
		loaded:          []string{starfile},
		state:           map[string]interface{}{},
		localizedStates: map[string]modules.LocalizedModule{},
	}
}

// NewChild returns a ModuleLoader child that shares the parent state and tracks
// the parent hierarchy to prevent recursive loading
func (loader *ModuleLoader) NewChild(module string) *ModuleLoader {
	return &ModuleLoader{
		mt:              loader.mt,
		rootdir:         loader.rootdir,
		loaded:          append(loader.loaded, module),
		state:           loader.state,
		localizedStates: loader.localizedStates,
	}
}

// Load handles loading script modules
func (loader *ModuleLoader) Load(thread *starlark.Thread, modulePath string) (starlark.StringDict, error) {
	if _, ok := Modules[modulePath]; ok {
		if _, ok := loader.localizedStates[modulePath]; !ok {
			module, err := Modules[modulePath](loader)
			loader.localizedStates[modulePath] = module
			if err != nil {
				return nil, err
			}

		}

		return loader.localizedStates[modulePath].Exports(), nil
	}

	// fallback to using another starlark script as the module
	for i := range loader.loaded {
		if loader.loaded[i] == modulePath {
			return nil, fmt.Errorf("recusive loading of module %s is not allowed", modulePath)
		}
	}
	return loader.mt.NewChild(modulePath).Exec()
}

// SetState can be called by modules to update a state
func (loader *ModuleLoader) SetState(name string, ref any) error {
	value := reflect.ValueOf(ref)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return ErrStatePointerRequired
	}
	loader.state[name] = ref
	return nil
}

// GetState can be called by modules to acquire the current state value
func (loader *ModuleLoader) GetState(name string, dest any) error {
	if _, ok := loader.state[name]; !ok {
		return fmt.Errorf("%w: %s", ErrStateNotFound, name)
	}

	source := reflect.ValueOf(loader.state[name])
	destPtr := reflect.ValueOf(dest)

	// ensure the destination is a pointer
	if destPtr.Kind() != reflect.Ptr {
		return ErrStatePointerRequired
	}

	// unfurl the top level pointer
	destPtr = destPtr.Elem()

	// compare data types
	if destPtr.Type() != source.Type() {
		return fmt.Errorf("type mismatch when getting state: %s", name)
	}

	// update the destination
	destPtr.Set(source)

	return nil
}

func (loader *ModuleLoader) GetThreadName() string {
	return loader.mt.Name
}

func (loader *ModuleLoader) Destroy() {
	for name := range loader.localizedStates {
		if err := loader.localizedStates[name].Destroy(loader); err != nil {
			log.Printf("%s: error destroying module %s: %s", loader.mt.Name, name, err)
		}
	}
}
