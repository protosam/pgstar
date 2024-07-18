package router

import (
	"fmt"
	"net/http/pprof"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/protosam/pgstar/executor"
	"go.starlark.net/starlark"
)

type route struct {
	Methods []string
	Path    string
	Script  string
}

type Config struct {
	rootdir    string
	routes     []route
	globals    map[string]starlark.Value
	pprofRoute string
	options    []WithOption
}

type WithOption interface {
	Apply(*executor.ManagedThread) error
}

func ConfigureAndBuildRouter(starscript string, opts ...WithOption) (*mux.Router, error) {
	// script might be temporarily unavailable due to how some editors handle writes
	if err := waitForFile(starscript, configFileTimeout); err != nil {
		return nil, err
	}

	cfg := &Config{
		rootdir: filepath.Dir(starscript),
		options: opts,
	}
	thread := executor.NewManagedThread(cfg.rootdir, filepath.Base(starscript))
	thread.Predeclare("getEnv", starlark.NewBuiltin("getEnv", cfg.GetEnv))
	thread.Predeclare("setGlobal", starlark.NewBuiltin("setGlobal", cfg.SetGlobal))
	thread.Predeclare("addRoute", starlark.NewBuiltin("addRoute", cfg.AddRoute))
	thread.Predeclare("enableProfilerRoute", starlark.NewBuiltin("enableProfilerRoute", cfg.EnableProfilerRoute))
	thread.SetModuleLoader(executor.NewModuleLoader(thread, thread.GetRootdir(), thread.GetStarfile()))

	for i := range opts {
		opts[i].Apply(thread)
	}

	_, err := thread.Exec()
	if err != nil {
		return nil, fmt.Errorf("configuration failed to run: %w", err)
	}

	return cfg.BuildRouter(), nil
}

func (cfg *Config) BuildRouter() *mux.Router {
	router := mux.NewRouter()
	for _, route := range cfg.routes {
		router.HandleFunc(route.Path, WithStarlarkHandler(cfg.rootdir, route.Script, cfg.globals, cfg.options...)).Methods(route.Methods...)
	}

	// enable pprof for Go debugging
	if cfg.pprofRoute != "" {
		pprofRouter := router.PathPrefix(cfg.pprofRoute).Subrouter()

		// Add pprof routes to the subrouter
		pprofRouter.HandleFunc("/", pprof.Index)
		pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
		pprofRouter.HandleFunc("/profile", pprof.Profile)
		pprofRouter.HandleFunc("/symbol", pprof.Symbol)
		pprofRouter.HandleFunc("/trace", pprof.Trace)

		// These three routes are automatically added by pprof.Index, but if you need them explicitly:
		pprofRouter.Handle("/goroutine", pprof.Handler("goroutine"))
		pprofRouter.Handle("/heap", pprof.Handler("heap"))
		pprofRouter.Handle("/threadcreate", pprof.Handler("threadcreate"))
		pprofRouter.Handle("/block", pprof.Handler("block"))
	}

	return router
}

func (cfg *Config) NullFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func (cfg *Config) GetEnv(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var defaultValue starlark.Value
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "default_value", &defaultValue); err != nil {
		return starlark.None, err
	}

	// ensure environment variables are prefixed
	name = "PGSTAR_ENV_" + name

	envval := os.Getenv(name)
	if envval == "" {
		return defaultValue, nil
	}

	return starlark.String(envval), nil
}

func (cfg *Config) SetGlobal(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var value starlark.Value
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "value", &value); err != nil {
		return starlark.None, err
	}

	if cfg.globals == nil {
		cfg.globals = make(map[string]starlark.Value)
	}

	cfg.globals[name] = value

	return starlark.None, nil
}

func (cfg *Config) AddRoute(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	sval_methods := starlark.NewList(nil)
	var path string
	var script string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "methods", &sval_methods, "path", &path, "script", &script); err != nil {
		return starlark.None, err
	}

	var methods []string
	for i := 0; i < sval_methods.Len(); i++ {
		if method, ok := starlark.AsString(sval_methods.Index(i)); ok {
			methods = append(methods, method)
		} else {
			return starlark.None, fmt.Errorf("method must be a list of strings")
		}
	}

	cfg.routes = append(cfg.routes, route{
		Methods: methods,
		Path:    path,
		Script:  script,
	})

	return starlark.None, nil
}

func (cfg *Config) EnableProfilerRoute(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "pprofRoute", &cfg.pprofRoute); err != nil {
		return starlark.None, err
	}
	return starlark.None, nil
}
