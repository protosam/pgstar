package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/router"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "server",
	Usage: "Start a server with the specified configuration file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "bind-addr",
			Usage:   "Path to the optional configuration file",
			EnvVars: []string{"PGSTAR_BIND_ADDR"},
			Value:   "127.0.0.1:5000",
		},
		&cli.StringFlag{
			Name:     "postgres-config",
			Usage:    "Connection string for postgres connection",
			EnvVars:  []string{"PGSTAR_POSTGRES_CONFIG"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "ssl-cert",
			Usage:   "Certificate for enabling SSL",
			EnvVars: []string{"PGSTAR_SSL_CERTIFICATE"},
		},
		&cli.StringFlag{
			Name:    "ssl-key",
			Usage:   "Private key to enable SSL",
			EnvVars: []string{"PGSTAR_SSL_PRIVATE_KEY"},
		},
	},
	Action: main,
}

var server = &http.Server{
	Handler: nil,
}

func main(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("you must provide a path to the configuration file")
	}
	server.Addr = c.String("bind-addr")
	PGSTAR_POSTGRES_CONFIG := c.String("postgres-config")
	PGSTAR_SSL_CERTIFICATE := c.String("ssl-cert")
	PGSTAR_SSL_PRIVATE_KEY := c.String("ssl-key")
	starfile := c.Args().Get(0)

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

	// load initial configuration
	loadConfig(starfile)

	// autoreloading for config changes
	go configReloader(starfile)

	if PGSTAR_SSL_CERTIFICATE != "" || PGSTAR_SSL_PRIVATE_KEY != "" {
		if err := server.ListenAndServeTLS(PGSTAR_SSL_CERTIFICATE, PGSTAR_SSL_PRIVATE_KEY); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v\n", err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v\n", err)
		}
	}

	return nil
}

func loadConfig(starfile string) {
	log.Printf("loading configuration: %s", starfile)
	router, err := router.ConfigureAndBuildRouter(starfile)
	if err != nil {
		log.Fatal(err)
	}
	server.Handler = router
	log.Printf("configuration reloaded successfully: %s", starfile)
}

func configReloader(starfile string) {
	// Create a new watcher instance
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Close()

	// Add the file to the watcher
	err = watcher.Add(starfile)
	if err != nil {
		log.Fatalf("error starting config watcher for file: %v", err)
	}

	log.Printf("watching config file for changes: %s", starfile)

	// Start watching for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("configuration updated")
				loadConfig(starfile)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("config watcher error: %v", err)
		}
	}
}
