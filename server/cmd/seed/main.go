package main

import (
	"flag"
	"fmt"
	"log"

	"ubom-v4/seed"
	"ubom-v4/store"
)

func main() {
	dbPath := flag.String("db", "dev.db", "SQLite database path")
	parts := flag.Int("parts", 50, "number of part numbers")
	maxRevisions := flag.Int("max-revisions", 3, "maximum revisions per part number")
	maxBOMDepth := flag.Int("max-bom-depth", 3, "maximum generated BOM depth")
	seedValue := flag.Int64("seed", 1, "deterministic random seed")
	flag.Parse()

	persistence, err := store.OpenSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer persistence.Close()

	result, err := seed.Populate(persistence, seed.Options{
		Parts:        *parts,
		MaxRevisions: *maxRevisions,
		MaxBOMDepth:  *maxBOMDepth,
		Seed:         *seedValue,
	})
	if err != nil {
		log.Fatalf("seed database: %v", err)
	}
	fmt.Printf("seeded %d parts, %d revisions, and %d BOM lines into %s\n", result.Parts, result.Revisions, result.BOMLines, *dbPath)
}
