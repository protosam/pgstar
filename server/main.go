package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protosam/pgstar/router"
)

var PGSTAR_SSL_CERTIFICATE = os.Getenv("PGSTAR_SSL_CERTIFICATE")
var PGSTAR_SSL_PRIVATE_KEY = os.Getenv("PGSTAR_SSL_PRIVATE_KEY")
var PGSTAR_POSTGRES_CONFIG = os.Getenv("PGSTAR_POSTGRES_CONFIG")
var server = &http.Server{
	Addr:    CoalesceEnv("PGSTAR_BIND_ADDR", "127.0.0.1:5000"),
	Handler: nil,
}

func CoalesceEnv(varname, fallback string) string {
	value := os.Getenv(varname)
	if value != "" {
		return value
	}
	return fallback
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("script file must be provided as argument")
	}
	starfile := os.Args[1]

	// Postgres connection pool setup.
	// The connection pool is configured via environment variable.
	// Example: export PGSTAR_POSTGRES_CONFIG="host=localhost port=5433 user=yugabyte password=yugabyte database=mydatabase sslmode=disable"
	dbpool, err := pgxpool.New(context.Background(), PGSTAR_POSTGRES_CONFIG)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	// ensure dbpool is passed to router
	router.SetDBPool(dbpool)

	// Ping the database to verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %s", err)
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
