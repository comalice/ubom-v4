package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	_ "modernc.org/sqlite"
	ubom "ubom-v4"
)

type SQLiteStore struct {
	db *sql.DB
}

var _ Store = (*SQLiteStore)(nil)

func OpenSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	store, err := NewSQLiteStore(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) init() error {
	_, err := s.db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS seq_defs (
			id TEXT PRIMARY KEY,
			ast TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS taxonomy_defs (
			id TEXT PRIMARY KEY,
			seq_def_id TEXT NOT NULL,
			taxonomy TEXT NOT NULL,
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id)
		);
		CREATE TABLE IF NOT EXISTS part_numbers (
			value TEXT PRIMARY KEY,
			seq_def_id TEXT NOT NULL,
			taxonomy_def_id TEXT NOT NULL,
			taxonomy_node_id TEXT NOT NULL,
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id),
			FOREIGN KEY (taxonomy_def_id) REFERENCES taxonomy_defs(id)
		);`)
	return err
}

func (s *SQLiteStore) CreateSeqDef(def ubom.SeqDef) error {
	if _, err := s.GetSeqDef(def.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	ast, err := marshalSeqDef(def)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO seq_defs (id, ast) VALUES (?, ?)", def.ID, ast)
	return err
}

func (s *SQLiteStore) GetSeqDef(id ubom.SeqDefID) (ubom.SeqDef, error) {
	var ast []byte
	if err := s.db.QueryRow("SELECT ast FROM seq_defs WHERE id = ?", id).Scan(&ast); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.SeqDef{}, ErrNotFound
		}
		return ubom.SeqDef{}, err
	}
	return unmarshalSeqDef(ast)
}

func (s *SQLiteStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	if _, err := s.GetTaxonomyDef(def.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	taxonomy, err := json.Marshal(def.Taxonomy)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO taxonomy_defs (id, seq_def_id, taxonomy) VALUES (?, ?, ?)", def.ID, def.SeqDef, taxonomy)
	return err
}

func (s *SQLiteStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	var seqDefID ubom.SeqDefID
	var data []byte
	if err := s.db.QueryRow("SELECT seq_def_id, taxonomy FROM taxonomy_defs WHERE id = ?", id).Scan(&seqDefID, &data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.TaxonomyDef{}, ErrNotFound
		}
		return ubom.TaxonomyDef{}, err
	}
	var taxonomy ubom.Taxonomy
	if err := json.Unmarshal(data, &taxonomy); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	return ubom.TaxonomyDef{ID: id, SeqDef: seqDefID, Taxonomy: taxonomy}, nil
}

func (s *SQLiteStore) CreatePartNumber(part ubom.PartNumber) error {
	if _, err := s.GetPartNumber(part.Value); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.db.Exec(`INSERT INTO part_numbers
		(value, seq_def_id, taxonomy_def_id, taxonomy_node_id)
		VALUES (?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID)
	return err
}

func (s *SQLiteStore) GetPartNumber(value string) (ubom.PartNumber, error) {
	var part ubom.PartNumber
	if err := s.db.QueryRow(`SELECT value, seq_def_id, taxonomy_def_id, taxonomy_node_id
		FROM part_numbers WHERE value = ?`, value).Scan(
		&part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartNumber{}, ErrNotFound
		}
		return ubom.PartNumber{}, err
	}
	return part, nil
}
