package store

import (
	"strings"
	"testing"
)

func TestPostgresPlaceholders(t *testing.T) {
	got := postgresPlaceholders("SELECT * FROM parts WHERE value = ? AND status = '?' AND id = ?")
	want := "SELECT * FROM parts WHERE value = $1 AND status = '?' AND id = $2"
	if got != want {
		t.Fatalf("postgresPlaceholders() = %q, want %q", got, want)
	}
}

func TestPostgresMigrationsIncludeTaxonomyNodeLookupIndexes(t *testing.T) {
	for _, migration := range postgresMigrations {
		if strings.Contains(migration, "idx_part_numbers_taxonomy_node_value") {
			if !strings.Contains(migration, "idx_part_revisions_part_number_id_id") {
				t.Fatalf("migration with part-number index is missing revision index")
			}
		}
	}
	if !strings.Contains(postgresMigrations[0], "CREATE INDEX IF NOT EXISTS idx_part_numbers_taxonomy_node_value") ||
		!strings.Contains(postgresMigrations[0], "CREATE INDEX IF NOT EXISTS idx_part_revisions_part_number_id_id") {
		t.Fatal("initial Postgres migration is missing taxonomy-node lookup indexes")
	}
	if !strings.Contains(postgresMigrations[len(postgresMigrations)-1], "CREATE INDEX IF NOT EXISTS idx_part_numbers_taxonomy_node_value") {
		t.Fatal("upgrade migration is missing part-number taxonomy-node index")
	}
}
