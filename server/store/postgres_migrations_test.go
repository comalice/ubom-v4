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
	var hasIndexUpgrade, hasBranchUpgrade, hasSchemaTable bool
	for _, migration := range postgresMigrations[1:] {
		hasIndexUpgrade = hasIndexUpgrade || strings.Contains(migration, "CREATE INDEX IF NOT EXISTS idx_part_numbers_taxonomy_node_value")
		hasBranchUpgrade = hasBranchUpgrade || strings.Contains(migration, "ALTER TABLE taxonomy_nodes ADD COLUMN IF NOT EXISTS branch TEXT")
		hasSchemaTable = hasSchemaTable || strings.Contains(migration, "CREATE TABLE IF NOT EXISTS part_number_schemas")
	}
	if !hasIndexUpgrade {
		t.Fatal("an upgrade migration is missing the part-number taxonomy-node index")
	}
	if !hasBranchUpgrade {
		t.Fatal("an upgrade migration is missing the taxonomy branch column upgrade")
	}
	if !hasSchemaTable {
		t.Fatal("an upgrade migration is missing the part-number schema table")
	}
}
