package modtime

import (
	"time"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "time"
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
				"now":      starlark.NewBuiltin("time.now", now),
				"format":   starlark.NewBuiltin("time.format", format),
				"epoch":    starlark.NewBuiltin("time.epoch", epoch),
				"timezone": starlark.NewBuiltin("time.timezone", timezone),

				// Just pass along the convenience strings
				"Layout":      starlark.String(time.Layout),
				"ANSIC":       starlark.String(time.ANSIC),
				"UnixDate":    starlark.String(time.UnixDate),
				"RubyDate":    starlark.String(time.RubyDate),
				"RFC822":      starlark.String(time.RFC822),
				"RFC822Z":     starlark.String(time.RFC822Z),
				"RFC850":      starlark.String(time.RFC850),
				"RFC1123":     starlark.String(time.RFC1123),
				"RFC1123Z":    starlark.String(time.RFC1123Z),
				"RFC3339":     starlark.String(time.RFC3339),
				"RFC3339Nano": starlark.String(time.RFC3339Nano),
				"Kitchen":     starlark.String(time.Kitchen),
				"Stamp":       starlark.String(time.Stamp),
				"StampMilli":  starlark.String(time.StampMilli),
				"StampMicro":  starlark.String(time.StampMicro),
				"StampNano":   starlark.String(time.StampNano),
				"DateTime":    starlark.String(time.DateTime),
				"DateOnly":    starlark.String(time.DateOnly),
				"TimeOnly":    starlark.String(time.TimeOnly),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func now(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	return starlark.MakeInt64(time.Now().UTC().Unix()), nil
}

func format(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var formatStr string
	var epoch int64
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 2, &epoch, &formatStr); err != nil {
		return starlark.None, err
	}

	return starlark.String(time.Unix(epoch, 0).UTC().Format(formatStr)), nil
}

func epoch(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var formatStr, timestampStr string
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 2, &timestampStr, &formatStr); err != nil {
		return starlark.None, err
	}

	timestamp, err := time.Parse(formatStr, timestampStr)
	if err != nil {
		return starlark.None, err
	}

	return starlark.MakeInt64(timestamp.UTC().Unix()), nil
}

func timezone(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var timestampStr, formatStr, timezone string
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 3, &timestampStr, &formatStr, &timezone); err != nil {
		return starlark.None, err
	}

	// Load a location (time zone)
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return starlark.None, err
	}

	timestamp, err := time.Parse(formatStr, timestampStr)
	if err != nil {
		return starlark.None, err
	}

	// return starlark.String(timestamp.In(location).Format(formatStr)), nil
	return starlark.String(timestamp.In(location).Format(formatStr)), nil
}
