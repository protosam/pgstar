package modpostgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName      = "db"
	StateNameDBPool = "postgres/dbpool"
)

type Module struct {
	tx             pgx.Tx
	ctx            context.Context
	autosavepoints bool
	savepointname  string
}

func Constructor(loader modules.ModuleLoader) (modules.LocalizedModule, error) {
	var dbpool *pgxpool.Pool
	err := loader.GetState(StateNameDBPool, &dbpool)
	if err != nil {
		return nil, err
	}

	module := &Module{}
	module.ctx = context.Background()
	module.tx, err = dbpool.Begin(module.ctx)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("%s: failed to start transaction: %w", loader.GetThreadName(), err)
	}
	module.savepointname = strings.Join([]string{"pgstar", strings.ReplaceAll(uuid.Must(uuid.NewRandom()).String(), "-", "")}, "_")

	return module, nil
}

func (module *Module) Exports() starlark.StringDict {
	return starlark.StringDict{
		"exports": starlarkstruct.FromStringDict(
			starlark.String(ModuleName),
			starlark.StringDict{
				"savepoints": starlark.NewBuiltin("db.savepoints", module.savepoints),
				"query":      starlark.NewBuiltin("db.query", module.query),
				"exec":       starlark.NewBuiltin("db.exec", module.exec),
			},
		),
	}
}

func (module *Module) Name() string {
	return ModuleName
}

func (module *Module) Destroy(loader modules.ModuleLoader) error {
	defer module.tx.Rollback(module.ctx)

	if err := module.tx.Commit(module.ctx); err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("%s: unable to commit transaction: %w", loader.GetThreadName(), err)
	}

	return nil
}

func (module *Module) savepoints(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "enable", &module.autosavepoints); err != nil {
		return starlark.None, err
	}
	return starlark.None, nil
}

func (module *Module) query(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var sql string
	pgargs := starlark.NewList(nil)
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "sql", &sql, "args", &pgargs); err != nil {
		return starlark.None, err
	}

	params := pgargsToStarlarkValue(thread, fn, pgargs)

	if module.autosavepoints {
		if _, err := module.tx.Exec(module.ctx, "SAVEPOINT "+module.savepointname); err != nil {
			// this is an unrecoverable system error
			return &starlark.Tuple{
				starlark.None,
				starlark.None,
			}, fmt.Errorf("%s(): %s", fn.Name(), err)
		}
	}

	rows, err := module.tx.Query(module.ctx, sql, params...)
	if err != nil {
		if module.autosavepoints {
			if _, err := module.tx.Exec(module.ctx, "ROLLBACK TO SAVEPOINT "+module.savepointname); err != nil {
				// this is an unrecoverable system error
				return &starlark.Tuple{
					starlark.None,
					starlark.None,
				}, fmt.Errorf("%s(): %s", fn.Name(), err)
			}
		}

		return &starlark.Tuple{
			starlark.None,
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	slfields := starlark.NewList(nil)
	var fields []string
	for _, field := range rows.FieldDescriptions() {
		fields = append(fields, field.Name)
		slfields.Append(starlark.String(field.Name))
	}

	slrows := &Rows{
		Struct: *starlarkstruct.FromStringDict(
			starlark.String("db.rows"),
			starlark.StringDict{
				"fields": slfields,
			},
		),
		fields: fields,
		rows:   rows,
		module: module,
		thread: thread,
	}

	return &starlark.Tuple{slrows, starlark.None}, nil
}

func (module *Module) exec(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var sql string
	pgargs := starlark.NewList(nil)
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "sql", &sql, "args", &pgargs); err != nil {
		return starlark.None, err
	}

	params := pgargsToStarlarkValue(thread, fn, pgargs)

	if module.autosavepoints {
		if _, err := module.tx.Exec(module.ctx, "SAVEPOINT "+module.savepointname); err != nil {
			// this is an unrecoverable system error
			return &starlark.Tuple{
				starlark.None,
				starlark.None,
			}, fmt.Errorf("%s(): %s", fn.Name(), err)
		}
	}

	cmdTag, err := module.tx.Exec(module.ctx, sql, params...)
	if err != nil {
		if module.autosavepoints {
			if _, err := module.tx.Exec(module.ctx, "ROLLBACK TO SAVEPOINT "+module.savepointname); err != nil {
				return &starlark.Tuple{
					starlark.None,
					starlark.String(fmt.Sprintf("%s", err)),
				}, fmt.Errorf("%s(): %s", fn.Name(), err)
			}
		}

		return &starlark.Tuple{
			starlark.MakeInt64(cmdTag.RowsAffected()),
			starlark.String(fmt.Sprintf("%s", err)),
		}, nil
	}

	if module.autosavepoints {
		if _, err := module.tx.Exec(module.ctx, "RELEASE SAVEPOINT "+module.savepointname); err != nil {
			return &starlark.Tuple{
				starlark.None,
				starlark.String(fmt.Sprintf("%s", err)),
			}, fmt.Errorf("%s(): %s", fn.Name(), err)
		}
	}

	return &starlark.Tuple{
		starlark.MakeInt64(cmdTag.RowsAffected()),
		starlark.None,
	}, nil
}
