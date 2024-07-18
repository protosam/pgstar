package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/executor"
	"github.com/protosam/pgstar/executor/modules"
	"github.com/protosam/pgstar/executor/modules/modhttp"
	"github.com/protosam/pgstar/executor/modules/modpostgres"
	"go.starlark.net/starlark"
)

var dbpool *pgxpool.Pool

type route struct {
	Methods []string
	Path    string
	Script  string
}

type Config struct {
	rootdir string
	routes  []route
	globals map[string]starlark.Value
}

func SetDBPool(pool *pgxpool.Pool) {
	dbpool = pool
}

func ConfigureAndBuildRouter(starscript string) (*mux.Router, error) {
	cfg := &Config{
		rootdir: filepath.Dir(starscript),
	}
	thread := executor.NewManagedThread(cfg.rootdir, filepath.Base(starscript))
	thread.Predeclare("getEnv", starlark.NewBuiltin("getEnv", cfg.GetEnv))
	thread.Predeclare("setGlobal", starlark.NewBuiltin("setGlobal", cfg.SetGlobal))
	thread.Predeclare("addRoute", starlark.NewBuiltin("addRoute", cfg.AddRoute))
	thread.SetModuleLoader(executor.NewModuleLoader(thread, thread.GetRootdir(), thread.GetStarfile()))

	_, err := thread.Exec()
	if err != nil {
		return nil, fmt.Errorf("configuration failed to run: %w", err)
	}

	return cfg.BuildRouter(), nil
}

func withStarlark(rootdir, starfile string, globals map[string]starlark.Value) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		thread := executor.NewManagedThread(rootdir, starfile)
		moduleloader := executor.NewModuleLoader(thread, thread.GetRootdir(), thread.GetStarfile())
		moduleloader.SetState(modpostgres.StateNameDBPool, dbpool)
		moduleloader.SetState(modhttp.StateNameReader, r)
		moduleloader.SetState(modhttp.StateNameWriter, &w)
		thread.SetModuleLoader(moduleloader)

		for name, value := range globals {
			thread.Predeclare(name, value)
		}

		defer moduleloader.Destroy()
		if _, err := thread.Exec(); err != nil {
			if !errors.Is(err, modules.ErrEarlyExit) {
				log.Printf("%s: error: %s", thread.Name, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (cfg *Config) BuildRouter() *mux.Router {
	router := mux.NewRouter()
	for _, route := range cfg.routes {
		router.HandleFunc(route.Path, withStarlark(cfg.rootdir, route.Script, cfg.globals)).Methods(route.Methods...)
	}
	return router
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
