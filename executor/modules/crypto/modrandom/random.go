package modrandom

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "random"
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
				"bytes": starlark.NewBuiltin("random.bytes", randBytes),
				"int":   starlark.NewBuiltin("random.int", randomInt),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func randBytes(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var length int
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "length", &length); err != nil {
		return starlark.None, err
	}

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	return starlark.String(randomBytes), nil
}

func randomInt(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var min, max int64
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "min", &min, "max", &max); err != nil {
		return starlark.None, err
	}

	// Check that min is less than max
	if min > max {
		return nil, fmt.Errorf("min must be less than max")
	}

	// Calculate the range
	rangeSize := max - min + 1

	// Generate a random number in the range [0, rangeSize)
	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return nil, err
	}

	return starlark.MakeInt64(n.Int64() + min), nil
}
