package main

import (
	"log"
	"os"

	"github.com/protosam/pgstar/cli/customerrors"
	"github.com/protosam/pgstar/cli/exec"
	"github.com/protosam/pgstar/cli/server"
	"github.com/urfave/cli/v2"
)

var App = &cli.App{
	Name:  "pgstar",
	Usage: "A command-line interface for pgstar",
	Commands: []*cli.Command{
		server.Command,
		exec.Command,
	},
}

func main() {
	err := App.Run(os.Args)

	if err != nil {
		// capture
		if exitCode, ok := err.(*customerrors.ExitWithCode); ok {
			os.Exit(exitCode.Code)
		}

		log.Fatal(err)
	}
}
