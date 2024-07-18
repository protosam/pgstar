package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/executor"
	"github.com/protosam/pgstar/executor/modules"
	"github.com/protosam/pgstar/executor/modules/modhttp"
	"github.com/protosam/pgstar/executor/modules/modpostgres"
	"go.starlark.net/starlark"
)

// SetDBPool updates the dbpool pointer
func SetDBPool(pool *pgxpool.Pool) {
	dbpool = pool
}

// WithStarlark returns an http handler function that runs a starlark script
func WithStarlark(rootdir, starfile string, globals map[string]starlark.Value) func(http.ResponseWriter, *http.Request) {
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

func waitForFile(path string, timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		_, err := os.Stat(path)
		select {
		case <-timeoutChan:
			return fmt.Errorf("failed to read file %s after waiting for %s: %w", path, timeout, err)
		case <-ticker.C:
			if err == nil {
				return nil
			}
		}
	}
}
