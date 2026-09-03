package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"ubom-v4/api"
	"ubom-v4/app"
	"ubom-v4/store"
)

func main() {
	config := Config{
		HTTP:     HTTPConfig{ListenAddress: "127.0.0.1:8080"},
		Database: DatabaseConfig{},
	}
	flag.StringVar(&config.Database.Path, "db", "", "SQLite database path; empty uses the in-memory sample")
	flag.StringVar(&config.Database.DSN, "database-url", os.Getenv("UBOM_DATABASE_URL"), "Postgres connection URL (or UBOM_DATABASE_URL)")
	flag.StringVar(&config.HTTP.ListenAddress, "listen", config.HTTP.ListenAddress, "HTTP listen address")
	flag.StringVar(&config.Artifacts.SnapshotRoot, "artifact-snapshot-root", "", "durable artifact snapshot content root")
	flag.StringVar(&config.Artifacts.LocalSource.ID, "local-source-id", "", "existing local artifact source ID")
	flag.StringVar(&config.Artifacts.LocalSource.Root, "local-source-root", "", "local artifact source root")
	flag.Parse()

	var persistence store.Store
	var closeStore func() error
	if config.Database.DSN != "" {
		postgres, err := store.OpenPostgresStore(config.Database.DSN)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		persistence = postgres
		closeStore = postgres.Close
	} else if config.Database.Path == "" {
		memory := store.NewMemoryStore()
		if _, err := app.LoadSampleData(memory); err != nil {
			log.Fatalf("load sample data: %v", err)
		}
		persistence = memory
	} else {
		sqlite, err := store.OpenSQLiteStore(config.Database.Path)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		persistence = sqlite
		closeStore = sqlite.Close
	}
	if closeStore != nil {
		defer closeStore()
	}

	service := app.NewService(persistence)
	snapshotService, err := config.Artifacts.snapshotService(persistence)
	if err != nil {
		log.Fatalf("configure artifacts: %v", err)
	}
	server := api.NewServer(service)
	if snapshotService != nil {
		server = api.NewServerWithSnapshotService(service, snapshotService)
	}
	if pinger, ok := persistence.(interface{ Ping() error }); ok {
		server.SetReadiness(pinger.Ping)
	}
	if config.Database.DSN != "" {
		log.Printf("using Postgres database")
	} else if config.Database.Path != "" {
		log.Printf("using SQLite database %s", config.Database.Path)
	} else {
		log.Println("using in-memory sample data")
	}
	log.Printf("listening on http://%s", config.HTTP.ListenAddress)
	if err := http.ListenAndServe(config.HTTP.ListenAddress, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
