package main

import (
	"log"
	"net/http"

	"ubom-v4/api"
	"ubom-v4/app"
	"ubom-v4/store"
)

func main() {
	persistence := store.NewMemoryStore()
	parent, err := app.LoadSampleData(persistence)
	if err != nil {
		log.Fatalf("load sample data: %v", err)
	}

	service := app.NewService(persistence)
	server := api.NewServer(service)
	log.Printf("sample part number available at /api/part-numbers/%s", parent.ID)
	log.Println("listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
