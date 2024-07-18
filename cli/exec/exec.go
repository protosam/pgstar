package exec

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/cli/customerrors"
	"github.com/protosam/pgstar/router"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "exec",
	Usage: "Run a script with an optional configuration file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "json-data",
			Usage: "JSON Encoded data to be passed in request",
		},
		&cli.StringSliceFlag{
			Name:  "header",
			Usage: "Headers to use, this can be passed multiple times (note these are normalized by Go to be capitalized)",
		},
		&cli.BoolFlag{
			Name:  "no-print",
			Usage: "Disables print() function instead of printing to stderr",
		},
		&cli.StringFlag{
			Name:     "postgres-config",
			Usage:    "Connection string for postgres connection",
			EnvVars:  []string{"PGSTAR_POSTGRES_CONFIG"},
			Required: true,
		},
	},
	Action: main,
}

var server = &http.Server{
	Handler: nil,
}

func main(c *cli.Context) error {
	if c.Args().Len() != 3 {
		return fmt.Errorf("a config file, http method, and path are required")
	}

	PGSTAR_POSTGRES_CONFIG := c.String("postgres-config")
	noPrint := c.Bool("no-print")
	jsonDataStr := c.String("json-data")
	headers := c.StringSlice("header")
	starfile := c.Args().Get(0)
	method := c.Args().Get(1)
	path := c.Args().Get(2)

	var request *http.Request

	if jsonDataStr != "" {
		switch method {
		default:
			return fmt.Errorf("can not use json data with method %s", method)
		case "POST":
		case "PUT":
		case "PATCH":
		case "DELETE":
			// NOOP
		}

		req, err := http.NewRequest(method, path, bytes.NewBuffer([]byte(jsonDataStr)))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		request = req
	} else {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			return err
		}
		request = req
	}

	for i := range headers {
		headerNameValue := strings.SplitN(headers[i], "=", 2)
		if len(headerNameValue) != 2 {
			return fmt.Errorf("invalid header flag value: %s", headers[i])
		}
		request.Header.Add(headerNameValue[0], headerNameValue[1])
	}

	// Postgres connection pool setup.
	dbpool, err := pgxpool.New(context.Background(), PGSTAR_POSTGRES_CONFIG)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	// ensure dbpool is passed to router
	router.SetDBPool(dbpool)

	// Ping the database to verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %s", err)
	}

	// log print output to stderr instead of stdout
	log.SetOutput(os.Stderr)

	var opts []router.WithOption
	if noPrint {
		opts = append(opts, router.WithNullPrinter())
	}

	router, err := router.ConfigureAndBuildRouter(starfile, opts...)
	if err != nil {
		return err
	}
	server.Handler = router

	// Create a new ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	router.ServeHTTP(rr, request)

	// print the response
	fmt.Println(rr.Body.String())

	return &customerrors.ExitWithCode{Code: rr.Code}
}
