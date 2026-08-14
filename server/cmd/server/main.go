package main

import (
	"flag"
	"log"
	"net/http"

	"ubom-v4/api"
	"ubom-v4/app"
	"ubom-v4/store"
)

func main() {
	dbPath := flag.String("db", "", "SQLite database path; empty uses the in-memory sample")
	flag.Parse()

	var persistence store.Store
	var closeStore func() error
	if *dbPath == "" {
		memory := store.NewMemoryStore()
		if _, err := app.LoadSampleData(memory); err != nil {
			log.Fatalf("load sample data: %v", err)
		}
		persistence = memory
	} else {
		sqlite, err := store.OpenSQLiteStore(*dbPath)
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
	server := api.NewServer(service)
	if *dbPath != "" {
		log.Printf("using SQLite database %s", *dbPath)
	} else {
		log.Println("using in-memory sample data")
	}
	log.Println("listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
