package router

import (
	"github.com/protosam/pgstar/executor"
	"go.starlark.net/starlark"
)

type WithPrinter struct {
	PrintFn func(thread *starlark.Thread, msg string)
}

func (opt *WithPrinter) Apply(thread *executor.ManagedThread) error {
	thread.Thread.Print = opt.PrintFn
	return nil
}

func WithNullPrinter() *WithPrinter {
	return &WithPrinter{PrintFn: func(thread *starlark.Thread, msg string) {}}
}
