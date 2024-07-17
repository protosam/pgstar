package executor

import (
	"log"

	"go.starlark.net/starlark"
)

func PrintLog(thread *starlark.Thread, msg string) {
	log.Printf("%s: %s", thread.Name, msg)
}
