package modpostgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type Rows struct {
	starlarkstruct.Struct
	module *Module
	fields []string
	rows   pgx.Rows
	thread *starlark.Thread
}

func (it *Rows) Freeze() {}

func (it *Rows) Truth() starlark.Bool {
	return true
}

func (it *Rows) Hash() (uint32, error) {
	return 0, fmt.Errorf("db.rows: unsupported operator")
}

func (it *Rows) Type() string {
	return "db.rows"
}

func (it *Rows) Iterate() starlark.Iterator {
	// Implement Iterate method for the iterable
	return &Row{module: it.module, fields: it.fields, rows: it.rows, thread: it.thread}
}

type Row struct {
	module *Module
	fields []string
	rows   pgx.Rows
	thread *starlark.Thread
}

func (it *Row) Next(p *starlark.Value) bool {
	if it.rows.Next() {
		values, _ := it.rows.Values()
		*p = parseRow(values, it.fields)
		return true
	}
	return false
}

func (it *Row) Done() {
	it.rows.Close()

	if it.module.autosavepoints {
		if _, err := it.module.tx.Exec(it.module.ctx, "RELEASE SAVEPOINT "+it.module.savepointname); err != nil {
			it.thread.Cancel(fmt.Sprintf("failed to release db savepoint: %s", err))
		}
	}
}
