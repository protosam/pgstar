package moddb

import (
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
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
		// slvals := starlark.NewList(nil)
		slvals := starlark.NewDict(0)
		for idx := range values {
			switch values[idx].(type) {
			case int:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.MakeInt(values[idx].(int)))
			case string:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.String(values[idx].(string)))
			case float32:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.Float(values[idx].(float32)))
			case float64:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.Float(values[idx].(float64)))
			case bool:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.Bool(values[idx].(bool)))
			case time.Time:
				val := values[idx].(time.Time)
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.String(val.String()))
			case [16]uint8:
				val := values[idx].([16]uint8)
				id, _ := uuid.FromBytes(val[:])
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.String(id.String()))
			case fmt.Stringer:
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.String(fmt.Sprintf("%s", values[idx])))
				log.Printf("failed back to Sprintf for type in db.rows.next(): %#v", values[idx])
			default:
				// slvals.Append(starlark.None)
				slvals.SetKey(starlark.String(it.fields[idx]), starlark.None)
				log.Printf("This type is not handled yet: %#v", values[idx])
			}
		}
		*p = slvals
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
