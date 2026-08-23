package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
	ubom "ubom-v4"
	"ubom-v4/artifact"
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
	db.SetMaxOpenConns(1)
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
		CREATE TABLE IF NOT EXISTS artifact_sources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			kind TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS artifacts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			source_id TEXT NOT NULL,
			path TEXT NOT NULL,
			FOREIGN KEY (source_id) REFERENCES artifact_sources(id)
		);
		CREATE TABLE IF NOT EXISTS artifact_snapshots (
			id TEXT PRIMARY KEY,
			artifact_id TEXT NOT NULL,
			reference TEXT NOT NULL,
			source_commit TEXT NOT NULL,
			snapshot_key TEXT NOT NULL,
			digest TEXT NOT NULL,
			media_type TEXT NOT NULL,
			size INTEGER NOT NULL,
			FOREIGN KEY (artifact_id) REFERENCES artifacts(id)
		);
		CREATE TABLE IF NOT EXISTS taxonomy_defs (
			id TEXT PRIMARY KEY,
			seq_def_id TEXT NOT NULL,
			attribute_defs TEXT NOT NULL DEFAULT '[]',
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id)
		);
		CREATE TABLE IF NOT EXISTS taxonomy_nodes (
			taxonomy_def_id TEXT NOT NULL,
			id TEXT NOT NULL,
			parent_id TEXT,
			label TEXT NOT NULL,
			matches TEXT NOT NULL,
			attributes TEXT NOT NULL DEFAULT '[]',
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
			attributes TEXT NOT NULL DEFAULT '[]',
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id),
			FOREIGN KEY (taxonomy_def_id) REFERENCES taxonomy_defs(id),
			FOREIGN KEY (taxonomy_def_id, taxonomy_node_id)
				REFERENCES taxonomy_nodes(taxonomy_def_id, id)
		);
		CREATE TABLE IF NOT EXISTS part_number_drafts (
			id TEXT PRIMARY KEY,
			seq_def_id TEXT NOT NULL,
			taxonomy_def_id TEXT NOT NULL,
			context_node_id TEXT NOT NULL,
			bindings TEXT NOT NULL,
			attributes TEXT NOT NULL,
			status TEXT NOT NULL,
			released_part_number_id INTEGER,
			FOREIGN KEY (seq_def_id) REFERENCES seq_defs(id),
			FOREIGN KEY (taxonomy_def_id, context_node_id)
				REFERENCES taxonomy_nodes(taxonomy_def_id, id),
			FOREIGN KEY (released_part_number_id) REFERENCES part_numbers(id)
		);
		CREATE TABLE IF NOT EXISTS part_revisions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			part_number_id INTEGER NOT NULL,
			revision TEXT NOT NULL,
			revision_seq_def_id TEXT NOT NULL,
			UNIQUE (part_number_id, revision),
			FOREIGN KEY (revision_seq_def_id) REFERENCES seq_defs(id),
			FOREIGN KEY (part_number_id) REFERENCES part_numbers(id)
		);
		CREATE TABLE IF NOT EXISTS bom_line_items (
			parent_revision_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			child_part_number_id INTEGER NOT NULL,
			child_revision_id INTEGER NOT NULL,
			PRIMARY KEY (parent_revision_id, position),
			FOREIGN KEY (parent_revision_id) REFERENCES part_revisions(id),
			FOREIGN KEY (child_part_number_id) REFERENCES part_numbers(id),
			FOREIGN KEY (child_revision_id) REFERENCES part_revisions(id)
		);`)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) CreateSource(source artifact.Source) error {
	if err := source.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSource(source.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.db.Exec("INSERT INTO artifact_sources (id, name, kind) VALUES (?, ?, ?)", source.ID, source.Name, source.Kind)
	return err
}

func (s *SQLiteStore) GetSource(id artifact.SourceID) (artifact.Source, error) {
	var source artifact.Source
	if err := s.db.QueryRow("SELECT id, name, kind FROM artifact_sources WHERE id = ?", id).Scan(&source.ID, &source.Name, &source.Kind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Source{}, ErrNotFound
		}
		return artifact.Source{}, err
	}
	return source, nil
}

func (s *SQLiteStore) ListSources() ([]artifact.Source, error) {
	rows, err := s.db.Query("SELECT id, name, kind FROM artifact_sources ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []artifact.Source{}
	for rows.Next() {
		var source artifact.Source
		if err := rows.Scan(&source.ID, &source.Name, &source.Kind); err != nil {
			return nil, err
		}
		items = append(items, source)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) CreateArtifact(item artifact.Artifact) error {
	if err := item.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSource(item.SourceID); err != nil {
		return err
	}
	if _, err := s.GetArtifact(item.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.db.Exec("INSERT INTO artifacts (id, name, kind, source_id, path) VALUES (?, ?, ?, ?, ?)", item.ID, item.Name, item.Kind, item.SourceID, item.Path)
	return err
}

func (s *SQLiteStore) GetArtifact(id artifact.ID) (artifact.Artifact, error) {
	var item artifact.Artifact
	if err := s.db.QueryRow("SELECT id, name, kind, source_id, path FROM artifacts WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Kind, &item.SourceID, &item.Path); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Artifact{}, ErrNotFound
		}
		return artifact.Artifact{}, err
	}
	return item, nil
}

func (s *SQLiteStore) ListArtifacts() ([]artifact.Artifact, error) {
	rows, err := s.db.Query("SELECT id, name, kind, source_id, path FROM artifacts ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []artifact.Artifact{}
	for rows.Next() {
		var item artifact.Artifact
		if err := rows.Scan(&item.ID, &item.Name, &item.Kind, &item.SourceID, &item.Path); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) CreateSnapshot(snapshot artifact.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if _, err := s.GetArtifact(snapshot.ArtifactID); err != nil {
		return err
	}
	if _, err := s.GetSnapshot(snapshot.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.db.Exec(`INSERT INTO artifact_snapshots
		(id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, snapshot.ID, snapshot.ArtifactID, snapshot.Reference, snapshot.Commit, snapshot.SnapshotKey, snapshot.Digest, snapshot.MediaType, snapshot.Size)
	return err
}

func (s *SQLiteStore) GetSnapshot(id artifact.SnapshotID) (artifact.Snapshot, error) {
	var snapshot artifact.Snapshot
	if err := s.db.QueryRow(`SELECT id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size
		FROM artifact_snapshots WHERE id = ?`, id).Scan(&snapshot.ID, &snapshot.ArtifactID, &snapshot.Reference, &snapshot.Commit, &snapshot.SnapshotKey, &snapshot.Digest, &snapshot.MediaType, &snapshot.Size); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Snapshot{}, ErrNotFound
		}
		return artifact.Snapshot{}, err
	}
	return snapshot, nil
}

func (s *SQLiteStore) ListSnapshotsByArtifact(id artifact.ID) ([]artifact.Snapshot, error) {
	if _, err := s.GetArtifact(id); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size
		FROM artifact_snapshots WHERE artifact_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []artifact.Snapshot{}
	for rows.Next() {
		var snapshot artifact.Snapshot
		if err := rows.Scan(&snapshot.ID, &snapshot.ArtifactID, &snapshot.Reference, &snapshot.Commit, &snapshot.SnapshotKey, &snapshot.Digest, &snapshot.MediaType, &snapshot.Size); err != nil {
			return nil, err
		}
		items = append(items, snapshot)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) CreateSeqDef(def ubom.SeqDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
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

func (s *SQLiteStore) ListSeqDefs() ([]ubom.SeqDef, error) {
	rows, err := s.db.Query("SELECT ast FROM seq_defs ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ubom.SeqDef{}
	for rows.Next() {
		var ast []byte
		if err := rows.Scan(&ast); err != nil {
			return nil, err
		}
		def, err := unmarshalSeqDef(ast)
		if err != nil {
			return nil, err
		}
		items = append(items, def)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSeqDef(def.SeqDef); err != nil {
		return err
	}
	if _, err := s.GetTaxonomyDef(def.ID); !errors.Is(err, ErrNotFound) {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	attributeDefs, err := json.Marshal(def.AttributeDefs)
	if err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec("INSERT INTO taxonomy_defs (id, seq_def_id, attribute_defs) VALUES (?, ?, ?)", def.ID, def.SeqDef, attributeDefs); err != nil {
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
	var attributeData []byte
	if err := s.db.QueryRow("SELECT seq_def_id, attribute_defs FROM taxonomy_defs WHERE id = ?", id).Scan(&seqDefID, &attributeData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.TaxonomyDef{}, ErrNotFound
		}
		return ubom.TaxonomyDef{}, err
	}
	taxonomy, err := s.loadTaxonomy(id)
	if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	attributeDefs := []ubom.AttributeDef{}
	if err := json.Unmarshal(attributeData, &attributeDefs); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	return ubom.TaxonomyDef{ID: id, SeqDef: seqDefID, AttributeDefs: attributeDefs, Taxonomy: taxonomy}, nil
}

func (s *SQLiteStore) ListTaxonomyDefs() ([]ubom.TaxonomyDef, error) {
	rows, err := s.db.Query("SELECT id FROM taxonomy_defs ORDER BY id")
	if err != nil {
		return nil, err
	}
	ids := []ubom.TaxonomyDefID{}
	for rows.Next() {
		var id ubom.TaxonomyDefID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	items := make([]ubom.TaxonomyDef, 0, len(ids))
	for _, id := range ids {
		def, err := s.GetTaxonomyDef(id)
		if err != nil {
			return nil, err
		}
		items = append(items, def)
	}
	return items, nil
}

func insertTaxonomyNode(tx *sql.Tx, taxonomyID ubom.TaxonomyDefID, parentID *string, node ubom.TaxonomyNode, position int) error {
	matches, err := json.Marshal(node.Matches)
	if err != nil {
		return err
	}
	attributes, err := json.Marshal(node.Attributes)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO taxonomy_nodes
		(taxonomy_def_id, id, parent_id, label, matches, attributes, position)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, taxonomyID, node.ID, parentID, node.Label, matches, attributes, position); err != nil {
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
	rows, err := s.db.Query(`SELECT id, parent_id, label, matches, attributes
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
		var attributeData []byte
		if err := rows.Scan(&id, &parent, &label, &data, &attributeData); err != nil {
			return ubom.Taxonomy{}, err
		}
		matches := map[string]string{}
		if err := json.Unmarshal(data, &matches); err != nil {
			return ubom.Taxonomy{}, err
		}
		attributes := []ubom.AttributeAssignment{}
		if err := json.Unmarshal(attributeData, &attributes); err != nil {
			return ubom.Taxonomy{}, err
		}
		node := &ubom.TaxonomyNode{ID: ubom.TaxonomyNodeID(id), Label: label, Matches: matches, Attributes: attributes}
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
	if err := part.Validate(); err != nil {
		return ubom.PartNumber{}, err
	}
	if err := validatePartNumberReferences(s, part); err != nil {
		return ubom.PartNumber{}, err
	}
	if _, err := s.GetPartNumber(part.Value); err == nil {
		return ubom.PartNumber{}, ErrAlreadyExists
	} else if !errors.Is(err, ErrNotFound) {
		return ubom.PartNumber{}, err
	}
	attributeData, err := json.Marshal(part.Attributes)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	result, err := s.db.Exec(`INSERT INTO part_numbers
		(value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes)
		VALUES (?, ?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID, attributeData)
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
	var attributeData []byte
	if err := s.db.QueryRow(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes
		FROM part_numbers WHERE value = ?`, value).Scan(
		&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID, &attributeData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartNumber{}, ErrNotFound
		}
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	if err := json.Unmarshal(attributeData, &part.Attributes); err != nil {
		return ubom.PartNumber{}, err
	}
	rows, err := s.db.Query("SELECT id FROM part_revisions WHERE part_number_id = ? ORDER BY id", id)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var revisionID int64
		if err := rows.Scan(&revisionID); err != nil {
			return ubom.PartNumber{}, err
		}
		part.PartRevisionID = append(part.PartRevisionID, ubom.PartRevisionID(strconv.FormatInt(revisionID, 10)))
	}
	if err := rows.Err(); err != nil {
		return ubom.PartNumber{}, err
	}
	return part, nil
}

func (s *SQLiteStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	partID, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ubom.PartNumber{}, ErrNotFound
	}
	part, err := s.getPartNumberByID(partID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	rows, err := s.db.Query("SELECT id FROM part_revisions WHERE part_number_id = ? ORDER BY id", partID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var revisionID int64
		if err := rows.Scan(&revisionID); err != nil {
			return ubom.PartNumber{}, err
		}
		part.PartRevisionID = append(part.PartRevisionID, ubom.PartRevisionID(strconv.FormatInt(revisionID, 10)))
	}
	if err := rows.Err(); err != nil {
		return ubom.PartNumber{}, err
	}
	return part, nil
}

func (s *SQLiteStore) ListPartNumbers() ([]ubom.PartNumber, error) {
	rows, err := s.db.Query("SELECT value FROM part_numbers ORDER BY value")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	parts := make([]ubom.PartNumber, 0, len(values))
	for _, value := range values {
		part, err := s.GetPartNumber(value)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func (s *SQLiteStore) ListPartNumbersByTaxonomyNode(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	taxonomy, err := s.GetTaxonomyDef(taxonomyID)
	if err != nil {
		return nil, err
	}
	if !taxonomyNodeExists(taxonomy.Taxonomy.Root, nodeID) {
		return nil, ErrNotFound
	}
	rows, err := s.db.Query(`SELECT value FROM part_numbers
		WHERE taxonomy_def_id = ? AND taxonomy_node_id = ? ORDER BY value`, taxonomyID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	parts := make([]ubom.PartNumber, 0, len(values))
	for _, value := range values {
		part, err := s.GetPartNumber(value)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func (s *SQLiteStore) CreatePartNumberDraft(draft ubom.PartNumberDraft) error {
	if err := draft.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSeqDef(draft.SeqDefID); err != nil {
		return err
	}
	taxonomy, err := s.GetTaxonomyDef(draft.TaxonomyDefID)
	if err != nil {
		return err
	}
	if taxonomy.SeqDef != draft.SeqDefID || !taxonomyNodeExists(taxonomy.Taxonomy.Root, draft.ContextNodeID) {
		return ErrNotFound
	}
	bindings, err := json.Marshal(draft.Bindings)
	if err != nil {
		return err
	}
	attributes, err := json.Marshal(draft.Attributes)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO part_number_drafts
		(id, seq_def_id, taxonomy_def_id, context_node_id, bindings, attributes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, draft.ID, draft.SeqDefID, draft.TaxonomyDefID, draft.ContextNodeID, bindings, attributes, draft.Status)
	return err
}

func (s *SQLiteStore) GetPartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumberDraft, error) {
	var draft ubom.PartNumberDraft
	var bindings, attributes []byte
	var released sql.NullInt64
	err := s.db.QueryRow(`SELECT id, seq_def_id, taxonomy_def_id, context_node_id, bindings, attributes, status, released_part_number_id
		FROM part_number_drafts WHERE id = ?`, id).Scan(&draft.ID, &draft.SeqDefID, &draft.TaxonomyDefID, &draft.ContextNodeID, &bindings, &attributes, &draft.Status, &released)
	if errors.Is(err, sql.ErrNoRows) {
		return ubom.PartNumberDraft{}, ErrNotFound
	}
	if err != nil {
		return ubom.PartNumberDraft{}, err
	}
	if err := json.Unmarshal(bindings, &draft.Bindings); err != nil {
		return ubom.PartNumberDraft{}, err
	}
	if err := json.Unmarshal(attributes, &draft.Attributes); err != nil {
		return ubom.PartNumberDraft{}, err
	}
	if released.Valid {
		draft.ReleasedPartNumberID = ubom.PartNumberID(strconv.FormatInt(released.Int64, 10))
	}
	return draft, nil
}

func (s *SQLiteStore) UpdatePartNumberDraft(draft ubom.PartNumberDraft) error {
	if err := draft.Validate(); err != nil {
		return err
	}
	current, err := s.GetPartNumberDraft(draft.ID)
	if err != nil {
		return err
	}
	if current.Status != ubom.DraftStatusDraft {
		return errors.New("part number draft is not editable")
	}
	bindings, err := json.Marshal(draft.Bindings)
	if err != nil {
		return err
	}
	attributes, err := json.Marshal(draft.Attributes)
	if err != nil {
		return err
	}
	result, err := s.db.Exec(`UPDATE part_number_drafts SET bindings = ?, attributes = ? WHERE id = ? AND status = ?`, bindings, attributes, draft.ID, ubom.DraftStatusDraft)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) ReleasePartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumber, error) {
	draft, err := s.GetPartNumberDraft(id)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if draft.Status != ubom.DraftStatusDraft {
		return ubom.PartNumber{}, errors.New("part number draft is not releasable")
	}
	seqDef, err := s.GetSeqDef(draft.SeqDefID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	taxonomy, err := s.GetTaxonomyDef(draft.TaxonomyDefID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	value, err := seqDef.Generate(draft.Bindings)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	part, err := ubom.NewPartNumber(value, seqDef, taxonomy)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if !taxonomyNodeIsAtOrBelow(taxonomy.Taxonomy.Root, draft.ContextNodeID, part.TaxonomyNodeID) {
		return ubom.PartNumber{}, errors.New("part number does not match draft taxonomy context")
	}
	part.Attributes = draft.Attributes
	if err := validatePartNumberReferences(s, part); err != nil {
		return ubom.PartNumber{}, err
	}
	attributeData, err := json.Marshal(part.Attributes)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return ubom.PartNumber{}, err
	}
	result, err := tx.Exec(`INSERT INTO part_numbers (value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes) VALUES (?, ?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID, attributeData)
	if err != nil {
		tx.Rollback()
		if isSQLiteConstraint(err) {
			return ubom.PartNumber{}, ErrAlreadyExists
		}
		return ubom.PartNumber{}, err
	}
	partID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return ubom.PartNumber{}, err
	}
	updated, err := tx.Exec(`UPDATE part_number_drafts SET status = ?, released_part_number_id = ? WHERE id = ? AND status = ?`, ubom.DraftStatusReleased, partID, id, ubom.DraftStatusDraft)
	if err != nil {
		tx.Rollback()
		return ubom.PartNumber{}, err
	}
	count, err := updated.RowsAffected()
	if err != nil || count != 1 {
		tx.Rollback()
		if err != nil {
			return ubom.PartNumber{}, err
		}
		return ubom.PartNumber{}, errors.New("part number draft changed during release")
	}
	if err := tx.Commit(); err != nil {
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(partID, 10))
	return part, nil
}

func isSQLiteConstraint(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func (s *SQLiteStore) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	var nextID int64
	if err := s.db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM part_revisions").Scan(&nextID); err != nil {
		return ubom.PartRevision{}, err
	}
	prospectiveID := ubom.PartRevisionID(strconv.FormatInt(nextID, 10))
	if err := revision.Validate(); err != nil {
		return ubom.PartRevision{}, err
	}
	if err := validatePartRevisionCycles(s, revision, prospectiveID); err != nil {
		return ubom.PartRevision{}, err
	}
	if err := validatePartRevisionReferences(s, revision); err != nil {
		return ubom.PartRevision{}, err
	}
	partNumberID, err := strconv.ParseInt(string(revision.PartNumberID), 10, 64)
	if err != nil {
		return ubom.PartRevision{}, ErrNotFound
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ubom.PartRevision{}, err
	}
	result, err := tx.Exec("INSERT INTO part_revisions (part_number_id, revision, revision_seq_def_id) VALUES (?, ?, ?)", partNumberID, revision.Revision, revision.RevisionSeqDefID)
	if err != nil {
		tx.Rollback()
		return ubom.PartRevision{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return ubom.PartRevision{}, err
	}
	for position, item := range revision.BOM {
		childPartID, partErr := strconv.ParseInt(string(item.PartNumberID), 10, 64)
		childRevisionID, revisionErr := strconv.ParseInt(string(item.PartRevisionID), 10, 64)
		if partErr != nil || revisionErr != nil {
			tx.Rollback()
			return ubom.PartRevision{}, ErrNotFound
		}
		if _, err := tx.Exec(`INSERT INTO bom_line_items
			(parent_revision_id, position, child_part_number_id, child_revision_id)
			VALUES (?, ?, ?, ?)`, id, position, childPartID, childRevisionID); err != nil {
			tx.Rollback()
			return ubom.PartRevision{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ubom.PartRevision{}, err
	}
	revision.ID = ubom.PartRevisionID(strconv.FormatInt(id, 10))
	return revision, nil
}

func (s *SQLiteStore) GetPartRevision(id ubom.PartRevisionID) (ubom.PartRevision, error) {
	revisionID, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ubom.PartRevision{}, ErrNotFound
	}
	var partNumberID int64
	var revisionValue string
	var revisionSeqDefID string
	if err := s.db.QueryRow("SELECT part_number_id, revision, revision_seq_def_id FROM part_revisions WHERE id = ?", revisionID).Scan(&partNumberID, &revisionValue, &revisionSeqDefID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartRevision{}, ErrNotFound
		}
		return ubom.PartRevision{}, err
	}
	revision := ubom.PartRevision{
		ID:               id,
		PartNumberID:     ubom.PartNumberID(strconv.FormatInt(partNumberID, 10)),
		Revision:         revisionValue,
		RevisionSeqDefID: ubom.SeqDefID(revisionSeqDefID),
	}
	rows, err := s.db.Query(`SELECT child_part_number_id, child_revision_id
		FROM bom_line_items WHERE parent_revision_id = ? ORDER BY position`, revisionID)
	if err != nil {
		return ubom.PartRevision{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var childPartID, childRevisionID int64
		if err := rows.Scan(&childPartID, &childRevisionID); err != nil {
			return ubom.PartRevision{}, err
		}
		revision.BOM = append(revision.BOM, ubom.LineItem{
			PartNumberID:   ubom.PartNumberID(strconv.FormatInt(childPartID, 10)),
			PartRevisionID: ubom.PartRevisionID(strconv.FormatInt(childRevisionID, 10)),
		})
	}
	if err := rows.Err(); err != nil {
		return ubom.PartRevision{}, err
	}
	return revision, nil
}

func (s *SQLiteStore) getPartNumberByID(id int64) (ubom.PartNumber, error) {
	var part ubom.PartNumber
	var attributeData []byte
	if err := s.db.QueryRow(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes
		FROM part_numbers WHERE id = ?`, id).Scan(
		&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID, &attributeData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartNumber{}, ErrNotFound
		}
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	if err := json.Unmarshal(attributeData, &part.Attributes); err != nil {
		return ubom.PartNumber{}, err
	}
	return part, nil
}
