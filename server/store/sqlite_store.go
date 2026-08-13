package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

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
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id)
		);
		CREATE TABLE IF NOT EXISTS taxonomy_nodes (
			taxonomy_def_id TEXT NOT NULL,
			id TEXT NOT NULL,
			parent_id TEXT,
			label TEXT NOT NULL,
			matches TEXT NOT NULL,
			position INTEGER NOT NULL,
			PRIMARY KEY (taxonomy_def_id, id),
			FOREIGN KEY (taxonomy_def_id) REFERENCES taxonomy_defs(id),
			FOREIGN KEY (taxonomy_def_id, parent_id)
				REFERENCES taxonomy_nodes(taxonomy_def_id, id)
		);
		CREATE TABLE IF NOT EXISTS part_numbers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT NOT NULL UNIQUE,
			seq_def_id TEXT NOT NULL,
			taxonomy_def_id TEXT NOT NULL,
			taxonomy_node_id TEXT NOT NULL,
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id),
			FOREIGN KEY (taxonomy_def_id) REFERENCES taxonomy_defs(id),
			FOREIGN KEY (taxonomy_def_id, taxonomy_node_id)
				REFERENCES taxonomy_nodes(taxonomy_def_id, id)
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO taxonomy_defs (id, seq_def_id) VALUES (?, ?)", def.ID, def.SeqDef); err != nil {
		tx.Rollback()
		return err
	}
	if err := insertTaxonomyNode(tx, def.ID, nil, def.Taxonomy.Root, 0); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	var seqDefID ubom.SeqDefID
	if err := s.db.QueryRow("SELECT seq_def_id FROM taxonomy_defs WHERE id = ?", id).Scan(&seqDefID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.TaxonomyDef{}, ErrNotFound
		}
		return ubom.TaxonomyDef{}, err
	}
	taxonomy, err := s.loadTaxonomy(id)
	if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	return ubom.TaxonomyDef{ID: id, SeqDef: seqDefID, Taxonomy: taxonomy}, nil
}

func insertTaxonomyNode(tx *sql.Tx, taxonomyID ubom.TaxonomyDefID, parentID *string, node ubom.TaxonomyNode, position int) error {
	matches, err := json.Marshal(node.Matches)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO taxonomy_nodes
		(taxonomy_def_id, id, parent_id, label, matches, position)
		VALUES (?, ?, ?, ?, ?, ?)`, taxonomyID, node.ID, parentID, node.Label, matches, position); err != nil {
		return err
	}
	parent := string(node.ID)
	for i, child := range node.Children {
		if err := insertTaxonomyNode(tx, taxonomyID, &parent, child, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) loadTaxonomy(id ubom.TaxonomyDefID) (ubom.Taxonomy, error) {
	rows, err := s.db.Query(`SELECT id, parent_id, label, matches
		FROM taxonomy_nodes WHERE taxonomy_def_id = ?
		ORDER BY parent_id, position`, id)
	if err != nil {
		return ubom.Taxonomy{}, err
	}
	defer rows.Close()

	nodes := map[ubom.TaxonomyNodeID]*ubom.TaxonomyNode{}
	type relation struct {
		id     ubom.TaxonomyNodeID
		parent sql.NullString
	}
	var relations []relation
	for rows.Next() {
		var id string
		var parent sql.NullString
		var label string
		var data []byte
		if err := rows.Scan(&id, &parent, &label, &data); err != nil {
			return ubom.Taxonomy{}, err
		}
		matches := map[string]string{}
		if err := json.Unmarshal(data, &matches); err != nil {
			return ubom.Taxonomy{}, err
		}
		node := &ubom.TaxonomyNode{ID: ubom.TaxonomyNodeID(id), Label: label, Matches: matches}
		nodes[node.ID] = node
		relations = append(relations, relation{id: node.ID, parent: parent})
	}
	if err := rows.Err(); err != nil {
		return ubom.Taxonomy{}, err
	}

	var root *ubom.TaxonomyNode
	for _, relation := range relations {
		node := nodes[relation.id]
		if !relation.parent.Valid {
			if root != nil {
				return ubom.Taxonomy{}, errors.New("taxonomy has multiple roots")
			}
			root = node
			continue
		}
		parent := nodes[ubom.TaxonomyNodeID(relation.parent.String)]
		if parent == nil {
			return ubom.Taxonomy{}, errors.New("taxonomy node has missing parent")
		}
		parent.Children = append(parent.Children, *node)
	}
	if root == nil {
		return ubom.Taxonomy{}, nil
	}
	return ubom.Taxonomy{Root: *root}, nil
}

func (s *SQLiteStore) CreatePartNumber(part ubom.PartNumber) (ubom.PartNumber, error) {
	if _, err := s.GetPartNumber(part.Value); !errors.Is(err, ErrNotFound) {
		return ubom.PartNumber{}, err
	}
	result, err := s.db.Exec(`INSERT INTO part_numbers
		(value, seq_def_id, taxonomy_def_id, taxonomy_node_id)
		VALUES (?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	return part, nil
}

func (s *SQLiteStore) GetPartNumber(value string) (ubom.PartNumber, error) {
	var part ubom.PartNumber
	var id int64
	if err := s.db.QueryRow(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id
		FROM part_numbers WHERE value = ?`, value).Scan(
		&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartNumber{}, ErrNotFound
		}
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	return part, nil
}
