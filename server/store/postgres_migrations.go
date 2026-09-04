package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// Migrations are deliberately small and embedded so a deployment has no
// external migration file dependency.
var postgresMigrations = []string{
	`CREATE TABLE IF NOT EXISTS seq_defs (id TEXT PRIMARY KEY, ast TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS artifact_sources (id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS artifacts (id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL, source_id TEXT NOT NULL REFERENCES artifact_sources(id), path TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS artifact_snapshots (id TEXT PRIMARY KEY, artifact_id TEXT NOT NULL REFERENCES artifacts(id), reference TEXT NOT NULL, source_commit TEXT NOT NULL, snapshot_key TEXT NOT NULL, digest TEXT NOT NULL, media_type TEXT NOT NULL, size BIGINT NOT NULL);
CREATE TABLE IF NOT EXISTS taxonomy_defs (id TEXT PRIMARY KEY, definition_id TEXT, version INTEGER, parent_id TEXT, seq_def_id TEXT NOT NULL REFERENCES seq_defs(id), attribute_defs TEXT NOT NULL DEFAULT '[]');
CREATE TABLE IF NOT EXISTS current_taxonomies (seq_def_id TEXT PRIMARY KEY REFERENCES seq_defs(id), taxonomy_def_id TEXT NOT NULL REFERENCES taxonomy_defs(id), state TEXT NOT NULL DEFAULT 'current');
 CREATE TABLE IF NOT EXISTS taxonomy_nodes (taxonomy_def_id TEXT NOT NULL REFERENCES taxonomy_defs(id), id TEXT NOT NULL, parent_id TEXT, label TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', image TEXT NOT NULL DEFAULT '', image_reference TEXT NOT NULL DEFAULT '', matches TEXT NOT NULL, predicates TEXT NOT NULL DEFAULT '[]', branch TEXT, attributes TEXT NOT NULL DEFAULT '[]', position INTEGER NOT NULL, PRIMARY KEY (taxonomy_def_id, id), FOREIGN KEY (taxonomy_def_id, parent_id) REFERENCES taxonomy_nodes(taxonomy_def_id, id));
CREATE TABLE IF NOT EXISTS part_numbers (id BIGSERIAL PRIMARY KEY, value TEXT NOT NULL UNIQUE, seq_def_id TEXT NOT NULL REFERENCES seq_defs(id), taxonomy_def_id TEXT NOT NULL REFERENCES taxonomy_defs(id), taxonomy_node_id TEXT NOT NULL, attributes TEXT NOT NULL DEFAULT '[]', status TEXT NOT NULL DEFAULT '', FOREIGN KEY (taxonomy_def_id, taxonomy_node_id) REFERENCES taxonomy_nodes(taxonomy_def_id, id));
CREATE INDEX IF NOT EXISTS idx_part_numbers_taxonomy_node_value ON part_numbers (taxonomy_def_id, taxonomy_node_id, value);
CREATE TABLE IF NOT EXISTS part_number_drafts (id TEXT PRIMARY KEY, seq_def_id TEXT NOT NULL REFERENCES seq_defs(id), taxonomy_def_id TEXT NOT NULL, context_node_id TEXT NOT NULL, bindings TEXT NOT NULL, attributes TEXT NOT NULL, status TEXT NOT NULL, released_part_number_id BIGINT REFERENCES part_numbers(id), FOREIGN KEY (taxonomy_def_id, context_node_id) REFERENCES taxonomy_nodes(taxonomy_def_id, id));
CREATE TABLE IF NOT EXISTS part_revisions (id BIGSERIAL PRIMARY KEY, part_number_id BIGINT NOT NULL REFERENCES part_numbers(id), revision TEXT NOT NULL, revision_seq_def_id TEXT NOT NULL REFERENCES seq_defs(id), UNIQUE (part_number_id, revision));
CREATE INDEX IF NOT EXISTS idx_part_revisions_part_number_id_id ON part_revisions (part_number_id, id);
CREATE TABLE IF NOT EXISTS bom_line_items (parent_revision_id BIGINT NOT NULL REFERENCES part_revisions(id), position INTEGER NOT NULL, child_part_number_id BIGINT NOT NULL REFERENCES part_numbers(id), child_revision_id BIGINT NOT NULL REFERENCES part_revisions(id), PRIMARY KEY (parent_revision_id, position));`,
	`ALTER TABLE taxonomy_nodes ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE taxonomy_nodes ADD COLUMN IF NOT EXISTS image TEXT NOT NULL DEFAULT '';
ALTER TABLE taxonomy_nodes ADD COLUMN IF NOT EXISTS image_reference TEXT NOT NULL DEFAULT '';`,
	`CREATE INDEX IF NOT EXISTS idx_part_numbers_taxonomy_node_value ON part_numbers (taxonomy_def_id, taxonomy_node_id, value);
CREATE INDEX IF NOT EXISTS idx_part_revisions_part_number_id_id ON part_revisions (part_number_id, id);`,
	`ALTER TABLE taxonomy_nodes ADD COLUMN IF NOT EXISTS branch TEXT;`,
	`CREATE TABLE IF NOT EXISTS part_number_schemas (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		taxonomy_def_id TEXT NOT NULL REFERENCES taxonomy_defs(id),
		require_classification BOOLEAN NOT NULL DEFAULT FALSE
	);`,
}

func runPostgresMigrations(db *sql.DB) error {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	for version, migration := range postgresMigrations {
		version++
		var applied bool
		if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(migration); err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(version) VALUES ($1)", version)
		}
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("postgres migration %d: %w", version, err)
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func postgresPlaceholders(query string) string {
	var b strings.Builder
	argument := 1
	quoted := byte(0)
	for i := 0; i < len(query); i++ {
		if quoted != 0 {
			b.WriteByte(query[i])
			if query[i] == quoted {
				if i+1 < len(query) && query[i+1] == quoted {
					b.WriteByte(query[i+1])
					i++
				} else {
					quoted = 0
				}
			}
			continue
		}
		if query[i] == '\'' || query[i] == '"' {
			quoted = query[i]
			b.WriteByte(query[i])
			continue
		}
		if query[i] == '?' {
			b.WriteString(fmt.Sprintf("$%d", argument))
			argument++
			continue
		}
		b.WriteByte(query[i])
	}
	return b.String()
}

func (s *PostgresStore) exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(postgresPlaceholders(query), args...)
}

func (s *PostgresStore) query(query string, args ...any) (*sql.Rows, error) {
	return s.db.Query(postgresPlaceholders(query), args...)
}

func (s *PostgresStore) queryRow(query string, args ...any) *sql.Row {
	return s.db.QueryRow(postgresPlaceholders(query), args...)
}

func postgresTxExec(tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	return tx.Exec(postgresPlaceholders(query), args...)
}

func postgresTxQueryRow(tx *sql.Tx, query string, args ...any) *sql.Row {
	return tx.QueryRow(postgresPlaceholders(query), args...)
}
