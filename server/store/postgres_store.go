package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	ubom "ubom-v4"
	"ubom-v4/artifact"
)

type PostgresStore struct {
	mu sync.Mutex
	db *sql.DB
}

var _ Store = (*PostgresStore)(nil)

func OpenPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	store, err := NewPostgresStore(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func NewPostgresStore(db *sql.DB) (*PostgresStore, error) {
	db.SetMaxOpenConns(10)
	store := &PostgresStore{db: db}
	if err := store.init(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Close() error { return s.db.Close() }

func (s *PostgresStore) Ping() error { return s.db.Ping() }

func (s *PostgresStore) init() error {
	return runPostgresMigrations(s.db)
}

func (s *PostgresStore) CreateSource(source artifact.Source) error {
	if err := source.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSource(source.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.exec("INSERT INTO artifact_sources (id, name, kind) VALUES (?, ?, ?)", source.ID, source.Name, source.Kind)
	return err
}

func (s *PostgresStore) GetSource(id artifact.SourceID) (artifact.Source, error) {
	var source artifact.Source
	if err := s.queryRow("SELECT id, name, kind FROM artifact_sources WHERE id = ?", id).Scan(&source.ID, &source.Name, &source.Kind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Source{}, ErrNotFound
		}
		return artifact.Source{}, err
	}
	return source, nil
}

func (s *PostgresStore) ListSources() ([]artifact.Source, error) {
	rows, err := s.query("SELECT id, name, kind FROM artifact_sources ORDER BY id")
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

func (s *PostgresStore) CreateArtifact(item artifact.Artifact) error {
	if err := item.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSource(item.SourceID); err != nil {
		return err
	}
	if _, err := s.GetArtifact(item.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.exec("INSERT INTO artifacts (id, name, kind, source_id, path) VALUES (?, ?, ?, ?, ?)", item.ID, item.Name, item.Kind, item.SourceID, item.Path)
	return err
}

func (s *PostgresStore) GetArtifact(id artifact.ID) (artifact.Artifact, error) {
	var item artifact.Artifact
	if err := s.queryRow("SELECT id, name, kind, source_id, path FROM artifacts WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Kind, &item.SourceID, &item.Path); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Artifact{}, ErrNotFound
		}
		return artifact.Artifact{}, err
	}
	return item, nil
}

func (s *PostgresStore) ListArtifacts() ([]artifact.Artifact, error) {
	rows, err := s.query("SELECT id, name, kind, source_id, path FROM artifacts ORDER BY id")
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

func (s *PostgresStore) CreateSnapshot(snapshot artifact.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if _, err := s.GetArtifact(snapshot.ArtifactID); err != nil {
		return err
	}
	if _, err := s.GetSnapshot(snapshot.ID); !errors.Is(err, ErrNotFound) {
		return err
	}
	_, err := s.exec(`INSERT INTO artifact_snapshots
		(id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, snapshot.ID, snapshot.ArtifactID, snapshot.Reference, snapshot.Commit, snapshot.SnapshotKey, snapshot.Digest, snapshot.MediaType, snapshot.Size)
	return err
}

func (s *PostgresStore) GetSnapshot(id artifact.SnapshotID) (artifact.Snapshot, error) {
	var snapshot artifact.Snapshot
	if err := s.queryRow(`SELECT id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size
		FROM artifact_snapshots WHERE id = ?`, id).Scan(&snapshot.ID, &snapshot.ArtifactID, &snapshot.Reference, &snapshot.Commit, &snapshot.SnapshotKey, &snapshot.Digest, &snapshot.MediaType, &snapshot.Size); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifact.Snapshot{}, ErrNotFound
		}
		return artifact.Snapshot{}, err
	}
	return snapshot, nil
}

func (s *PostgresStore) ListSnapshotsByArtifact(id artifact.ID) ([]artifact.Snapshot, error) {
	if _, err := s.GetArtifact(id); err != nil {
		return nil, err
	}
	rows, err := s.query(`SELECT id, artifact_id, reference, source_commit, snapshot_key, digest, media_type, size
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

func (s *PostgresStore) CreateSeqDef(def ubom.SeqDef) error {
	def = def.WithLegacyDefaults()
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
	_, err = s.exec("INSERT INTO seq_defs (id, ast) VALUES (?, ?)", def.ID, ast)
	return err
}

func (s *PostgresStore) CreateSeqDefVersion(parent ubom.SeqDefID, def ubom.SeqDef) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	original, err := s.GetSeqDef(parent)
	if err != nil {
		return err
	}
	if def.DefinitionID != original.DefinitionID || def.ParentID != parent || def.Version <= original.Version {
		return errors.New("invalid sequence definition version link")
	}
	existing, err := s.ListSeqDefs()
	if err != nil {
		return err
	}
	for _, candidate := range existing {
		candidate = candidate.WithLegacyDefaults()
		if candidate.DefinitionID == def.DefinitionID && candidate.Version == def.Version {
			return ErrAlreadyExists
		}
	}
	return s.CreateSeqDef(def)
}

func (s *PostgresStore) GetSeqDef(id ubom.SeqDefID) (ubom.SeqDef, error) {
	var ast []byte
	if err := s.queryRow("SELECT ast FROM seq_defs WHERE id = ?", id).Scan(&ast); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.SeqDef{}, ErrNotFound
		}
		return ubom.SeqDef{}, err
	}
	return unmarshalSeqDef(ast)
}

func (s *PostgresStore) ListSeqDefs() ([]ubom.SeqDef, error) {
	rows, err := s.query("SELECT ast FROM seq_defs ORDER BY id")
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

func (s *PostgresStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
	seqDef, err := s.GetSeqDef(def.SeqDef)
	if err != nil {
		return err
	}
	if taxonomyHasTypedPredicates(def.Taxonomy.Root) {
		if err := def.ValidateAgainst(seqDef); err != nil {
			return err
		}
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
	if _, err := postgresTxExec(tx, "INSERT INTO taxonomy_defs (id, definition_id, version, parent_id, seq_def_id, attribute_defs) VALUES (?, ?, ?, ?, ?, ?)", def.ID, postgresNullable(def.DefinitionID), postgresNullableInt(def.Version), postgresNullable(string(def.ParentID)), def.SeqDef, attributeDefs); err != nil {
		tx.Rollback()
		return err
	}
	if err := postgresInsertTaxonomyNode(tx, def.ID, nil, def.Taxonomy.Root, 0); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func postgresNullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func postgresNullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func (s *PostgresStore) CreateTaxonomyRevision(seqDef ubom.SeqDefID, parent ubom.TaxonomyDefID, def ubom.TaxonomyDef) (ubom.TaxonomyDef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	defer tx.Rollback()
	var current string
	err = postgresTxQueryRow(tx, "SELECT taxonomy_def_id FROM current_taxonomies WHERE seq_def_id = ? AND state = 'current'", seqDef).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		err = postgresTxQueryRow(tx, "SELECT id FROM taxonomy_defs WHERE seq_def_id = ? ORDER BY COALESCE(version, 1) DESC, id DESC LIMIT 1", seqDef).Scan(&current)
		if err == nil {
			_, err = postgresTxExec(tx, "INSERT INTO current_taxonomies(seq_def_id, taxonomy_def_id) VALUES (?, ?)", seqDef, current)
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ubom.TaxonomyDef{}, err
	}
	if (parent == "" && current != "") || (parent != "" && current != string(parent)) {
		return ubom.TaxonomyDef{}, ErrAlreadyExists
	}
	if parent != "" {
		var linked string
		if err := postgresTxQueryRow(tx, "SELECT seq_def_id FROM taxonomy_defs WHERE id = ?", parent).Scan(&linked); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ubom.TaxonomyDef{}, ErrNotFound
			}
			return ubom.TaxonomyDef{}, err
		}
		if linked != string(seqDef) {
			return ubom.TaxonomyDef{}, ErrAlreadyExists
		}
	}
	if err := def.Validate(); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	attributeDefs, err := json.Marshal(def.AttributeDefs)
	if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	if _, err = postgresTxExec(tx, "INSERT INTO taxonomy_defs (id, definition_id, version, parent_id, seq_def_id, attribute_defs) VALUES (?, ?, ?, ?, ?, ?)", def.ID, def.DefinitionID, def.Version, postgresNullable(string(def.ParentID)), seqDef, attributeDefs); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	if err = postgresInsertTaxonomyNode(tx, def.ID, nil, def.Taxonomy.Root, 0); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	if _, err = postgresTxExec(tx, "INSERT INTO current_taxonomies(seq_def_id, taxonomy_def_id, state) VALUES (?, ?, 'current') ON CONFLICT(seq_def_id) DO UPDATE SET taxonomy_def_id=excluded.taxonomy_def_id, state='current'", seqDef, def.ID); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	if err = tx.Commit(); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	return def, nil
}

func (s *PostgresStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	var seqDefID ubom.SeqDefID
	var definitionID sql.NullString
	var version sql.NullInt64
	var parentID sql.NullString
	var attributeData []byte
	if err := s.queryRow("SELECT seq_def_id, definition_id, version, parent_id, attribute_defs FROM taxonomy_defs WHERE id = ?", id).Scan(&seqDefID, &definitionID, &version, &parentID, &attributeData); err != nil {
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
	// Legacy rows contain only Matches. Upgrade them at the persistence
	// boundary so normal interpretation sees typed predicates.
	taxonomy, err = ubom.UpgradeLegacyTaxonomy(taxonomy)
	if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	result := ubom.TaxonomyDef{ID: id, SeqDef: seqDefID, AttributeDefs: attributeDefs, Taxonomy: taxonomy}
	if definitionID.Valid {
		result.DefinitionID = definitionID.String
	}
	if version.Valid {
		result.Version = int(version.Int64)
	}
	if parentID.Valid {
		result.ParentID = ubom.TaxonomyDefID(parentID.String)
	}
	return result, nil
}

func (s *PostgresStore) ListTaxonomyDefs() ([]ubom.TaxonomyDef, error) {
	rows, err := s.query("SELECT id FROM taxonomy_defs ORDER BY id")
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

func (s *PostgresStore) GetCurrentTaxonomyDef(seqDef ubom.SeqDefID) (ubom.TaxonomyDef, error) {
	var id ubom.TaxonomyDefID
	err := s.queryRow("SELECT taxonomy_def_id FROM current_taxonomies WHERE seq_def_id = ? AND state = 'current'", seqDef).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = s.queryRow("SELECT id FROM taxonomy_defs WHERE seq_def_id = ? ORDER BY COALESCE(version, 1) DESC, id DESC LIMIT 1", seqDef).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.TaxonomyDef{}, ErrNotFound
		}
		if err != nil {
			return ubom.TaxonomyDef{}, err
		}
		if _, err := s.exec("INSERT INTO current_taxonomies(seq_def_id, taxonomy_def_id) VALUES (?, ?) ON CONFLICT (seq_def_id) DO NOTHING", seqDef, id); err != nil {
			return ubom.TaxonomyDef{}, err
		}
	} else if err != nil {
		return ubom.TaxonomyDef{}, err
	}
	return s.GetTaxonomyDef(id)
}

func (s *PostgresStore) ListTaxonomyDefHistory(definitionID string) ([]ubom.TaxonomyDef, error) {
	rows, err := s.query("SELECT id FROM taxonomy_defs WHERE COALESCE(definition_id, id) = ? ORDER BY COALESCE(version, 1), id", definitionID)
	if err != nil {
		return nil, err
	}
	ids := []ubom.TaxonomyDefID{}
	for rows.Next() {
		var id ubom.TaxonomyDefID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
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
	result := []ubom.TaxonomyDef{}
	for _, id := range ids {
		def, err := s.GetTaxonomyDef(id)
		if err != nil {
			return nil, err
		}
		result = append(result, def)
	}
	return result, nil
}

func postgresInsertTaxonomyNode(tx *sql.Tx, taxonomyID ubom.TaxonomyDefID, parentID *string, node ubom.TaxonomyNode, position int) error {
	matches, err := json.Marshal(node.Matches)
	if err != nil {
		return err
	}
	// Matches remains a lossless legacy exact-match representation. Typed
	// predicates are separate and taxonomy validation rejects duplicate fields.
	predicates, err := json.Marshal(node.Predicates)
	if err != nil {
		return err
	}
	attributes, err := json.Marshal(node.Attributes)
	if err != nil {
		return err
	}
	if _, err := postgresTxExec(tx, `INSERT INTO taxonomy_nodes
		(taxonomy_def_id, id, parent_id, label, description, image, image_reference, matches, predicates, attributes, position)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, taxonomyID, node.ID, parentID, node.Label, node.Description, node.Image, node.ImageReference, matches, predicates, attributes, position); err != nil {
		return err
	}
	parent := string(node.ID)
	for i, child := range node.Children {
		if err := postgresInsertTaxonomyNode(tx, taxonomyID, &parent, child, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) loadTaxonomy(id ubom.TaxonomyDefID) (ubom.Taxonomy, error) {
	rows, err := s.query(`SELECT id, parent_id, label, description, image, image_reference, matches, predicates, attributes, position
		FROM taxonomy_nodes WHERE taxonomy_def_id = ?`, id)
	if err != nil {
		return ubom.Taxonomy{}, err
	}
	defer rows.Close()

	nodes := map[ubom.TaxonomyNodeID]*ubom.TaxonomyNode{}
	type relation struct {
		id       ubom.TaxonomyNodeID
		parent   sql.NullString
		position int
	}
	var relations []relation
	for rows.Next() {
		var id string
		var parent sql.NullString
		var label string
		var description string
		var image string
		var imageReference string
		var data []byte
		var predicateData []byte
		var attributeData []byte
		var position int
		if err := rows.Scan(&id, &parent, &label, &description, &image, &imageReference, &data, &predicateData, &attributeData, &position); err != nil {
			return ubom.Taxonomy{}, err
		}
		matches := map[string]string{}
		if err := json.Unmarshal(data, &matches); err != nil {
			return ubom.Taxonomy{}, err
		}
		predicates := []ubom.TaxonomyPredicate{}
		if err := json.Unmarshal(predicateData, &predicates); err != nil {
			return ubom.Taxonomy{}, err
		}
		attributes := []ubom.AttributeAssignment{}
		if err := json.Unmarshal(attributeData, &attributes); err != nil {
			return ubom.Taxonomy{}, err
		}
		node := &ubom.TaxonomyNode{ID: ubom.TaxonomyNodeID(id), Label: label, Description: description, Image: image, ImageReference: imageReference, Matches: matches, Predicates: predicates, Attributes: attributes}
		nodes[node.ID] = node
		relations = append(relations, relation{id: node.ID, parent: parent, position: position})
	}
	if err := rows.Err(); err != nil {
		return ubom.Taxonomy{}, err
	}

	children := make(map[ubom.TaxonomyNodeID][]relation)
	var rootID ubom.TaxonomyNodeID
	for _, relation := range relations {
		if !relation.parent.Valid {
			if rootID != "" {
				return ubom.Taxonomy{}, errors.New("taxonomy has multiple roots")
			}
			rootID = relation.id
			continue
		}
		children[ubom.TaxonomyNodeID(relation.parent.String)] = append(children[ubom.TaxonomyNodeID(relation.parent.String)], relation)
	}
	if rootID == "" {
		return ubom.Taxonomy{}, nil
	}
	for parentID := range children {
		sort.SliceStable(children[parentID], func(i, j int) bool { return children[parentID][i].position < children[parentID][j].position })
	}
	visited := make(map[ubom.TaxonomyNodeID]bool, len(nodes))
	var build func(ubom.TaxonomyNodeID) (ubom.TaxonomyNode, error)
	build = func(nodeID ubom.TaxonomyNodeID) (ubom.TaxonomyNode, error) {
		if visited[nodeID] {
			return ubom.TaxonomyNode{}, errors.New("taxonomy contains a cycle or duplicate path")
		}
		node, ok := nodes[nodeID]
		if !ok {
			return ubom.TaxonomyNode{}, errors.New("taxonomy node has missing parent")
		}
		visited[nodeID] = true
		result := *node
		result.Children = nil
		for _, child := range children[nodeID] {
			built, err := build(child.id)
			if err != nil {
				return ubom.TaxonomyNode{}, err
			}
			result.Children = append(result.Children, built)
		}
		return result, nil
	}
	root, err := build(rootID)
	if err != nil {
		return ubom.Taxonomy{}, err
	}
	if len(visited) != len(nodes) {
		return ubom.Taxonomy{}, errors.New("taxonomy contains an unreachable node")
	}
	return ubom.Taxonomy{Root: root}, nil
}

func (s *PostgresStore) CreatePartNumber(part ubom.PartNumber) (ubom.PartNumber, error) {
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
	_, err = s.exec(`INSERT INTO part_numbers
		(value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status)
		VALUES (?, ?, ?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID, attributeData, part.Status)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	var id int64
	if err := s.queryRow("SELECT id FROM part_numbers WHERE value = ?", part.Value).Scan(&id); err != nil {
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	return part, nil
}

func (s *PostgresStore) ArchivePartNumber(id ubom.PartNumberID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	partID, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ErrNotFound
	}
	result, err := s.exec("UPDATE part_numbers SET status = ? WHERE id = ? AND status != ?", ubom.PartNumberStatusArchived, partID, ubom.PartNumberStatusArchived)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		if _, err := s.getPartNumberByID(partID); err != nil {
			return err
		}
		return ErrAlreadyExists
	}
	return nil
}

func (s *PostgresStore) GetPartNumber(value string) (ubom.PartNumber, error) {
	var part ubom.PartNumber
	var id int64
	var attributeData []byte
	if err := s.queryRow(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status
		FROM part_numbers WHERE value = ?`, value).Scan(
		&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID, &attributeData, &part.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ubom.PartNumber{}, ErrNotFound
		}
		return ubom.PartNumber{}, err
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
	if err := json.Unmarshal(attributeData, &part.Attributes); err != nil {
		return ubom.PartNumber{}, err
	}
	rows, err := s.query("SELECT id FROM part_revisions WHERE part_number_id = ? ORDER BY id", id)
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

func (s *PostgresStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	partID, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ubom.PartNumber{}, ErrNotFound
	}
	part, err := s.getPartNumberByID(partID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	rows, err := s.query("SELECT id FROM part_revisions WHERE part_number_id = ? ORDER BY id", partID)
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

func (s *PostgresStore) ListPartNumbers() ([]ubom.PartNumber, error) {
	result, err := s.ListPartNumbersWithRevisionSummaries()
	if err != nil {
		return nil, err
	}
	return result.Parts, nil
}

func (s *PostgresStore) ListPartNumbersWithRevisionSummaries() (TaxonomyNodePartNumbers, error) {
	rows, err := s.query(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status
		FROM part_numbers ORDER BY value`)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	parts, partsByID, err := scanPartNumberRows(rows)
	if closeErr := rows.Close(); err != nil {
		return TaxonomyNodePartNumbers{}, err
	} else if closeErr != nil {
		return TaxonomyNodePartNumbers{}, closeErr
	}

	revisionRows, err := s.query(`SELECT r.part_number_id, r.id, r.revision
		FROM part_revisions r
		JOIN part_numbers p ON p.id = r.part_number_id
		ORDER BY p.value, r.id`)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	revisionSummaries, err := scanPartRevisionRows(revisionRows, parts, partsByID)
	if closeErr := revisionRows.Close(); err != nil {
		return TaxonomyNodePartNumbers{}, err
	} else if closeErr != nil {
		return TaxonomyNodePartNumbers{}, closeErr
	}
	return TaxonomyNodePartNumbers{Parts: parts, RevisionSummaries: revisionSummaries}, nil
}

func (s *PostgresStore) ListPartNumbersByTaxonomyNode(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	taxonomy, err := s.GetTaxonomyDef(taxonomyID)
	if err != nil {
		return nil, err
	}
	if !taxonomyNodeExists(taxonomy.Taxonomy.Root, nodeID) {
		return nil, ErrNotFound
	}
	return s.ListPartNumbersByTaxonomyNodeUnvalidated(taxonomyID, nodeID)
}

func (s *PostgresStore) ListPartNumbersByTaxonomyNodeUnvalidated(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	result, err := s.ListPartNumbersByTaxonomyNodeWithRevisionSummaries(taxonomyID, nodeID)
	if err != nil {
		return nil, err
	}
	return result.Parts, nil
}

func (s *PostgresStore) ListPartNumbersByTaxonomyNodeWithRevisionSummaries(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) (TaxonomyNodePartNumbers, error) {
	rows, err := s.query(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status
		FROM part_numbers WHERE taxonomy_def_id = ? AND taxonomy_node_id = ? ORDER BY value`, taxonomyID, nodeID)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	defer rows.Close()
	parts, partsByID, err := scanPartNumberRows(rows)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	if err := rows.Close(); err != nil {
		return TaxonomyNodePartNumbers{}, err
	}

	revisionRows, err := s.query(`SELECT r.part_number_id, r.id, r.revision
		FROM part_revisions r
		JOIN part_numbers p ON p.id = r.part_number_id
		WHERE p.taxonomy_def_id = ? AND p.taxonomy_node_id = ?
		ORDER BY p.value, r.id`, taxonomyID, nodeID)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	defer revisionRows.Close()
	revisionSummaries, err := scanPartRevisionRows(revisionRows, parts, partsByID)
	if err != nil {
		return TaxonomyNodePartNumbers{}, err
	}
	return TaxonomyNodePartNumbers{Parts: parts, RevisionSummaries: revisionSummaries}, nil
}

func (s *PostgresStore) CreatePartNumberDraft(draft ubom.PartNumberDraft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	_, err = s.exec(`INSERT INTO part_number_drafts
		(id, seq_def_id, taxonomy_def_id, context_node_id, bindings, attributes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, draft.ID, draft.SeqDefID, draft.TaxonomyDefID, draft.ContextNodeID, bindings, attributes, draft.Status)
	if isPostgresConstraint(err) {
		return ErrAlreadyExists
	}
	return err
}

func (s *PostgresStore) GetPartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumberDraft, error) {
	var draft ubom.PartNumberDraft
	var bindings, attributes []byte
	var released sql.NullInt64
	err := s.queryRow(`SELECT id, seq_def_id, taxonomy_def_id, context_node_id, bindings, attributes, status, released_part_number_id
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

func (s *PostgresStore) UpdatePartNumberDraft(draft ubom.PartNumberDraft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	result, err := s.exec(`UPDATE part_number_drafts SET bindings = ?, attributes = ? WHERE id = ? AND status = ?`, bindings, attributes, draft.ID, ubom.DraftStatusDraft)
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

func (s *PostgresStore) ReleasePartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumber, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	schema := ubom.NewSchema(seqDef, taxonomy, ubom.SchemaPolicy{RequireClassification: true})
	interpreted, err := schema.Generate(draft.Bindings)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	part, err := schema.NewPartNumber(interpreted.Parsed.Value, draft.Attributes)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if !taxonomyNodeIsAtOrBelow(taxonomy.Taxonomy.Root, draft.ContextNodeID, part.TaxonomyNodeID) || !taxonomyNodeIsLeaf(taxonomy.Taxonomy.Root, part.TaxonomyNodeID) {
		return ubom.PartNumber{}, ubom.NewDomainError(ubom.ErrorCategoryClassification, "part number does not match draft taxonomy context")
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
	_, err = postgresTxExec(tx, `INSERT INTO part_numbers (value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status) VALUES (?, ?, ?, ?, ?, ?)`, part.Value, part.SeqDefID, part.TaxonomyDefID, part.TaxonomyNodeID, attributeData, part.Status)
	if err != nil {
		tx.Rollback()
		if isPostgresConstraint(err) {
			return ubom.PartNumber{}, ErrAlreadyExists
		}
		return ubom.PartNumber{}, err
	}
	var partID int64
	if err := postgresTxQueryRow(tx, "SELECT id FROM part_numbers WHERE value = ?", part.Value).Scan(&partID); err != nil {
		tx.Rollback()
		return ubom.PartNumber{}, err
	}
	updated, err := postgresTxExec(tx, `UPDATE part_number_drafts SET status = ?, released_part_number_id = ? WHERE id = ? AND status = ?`, ubom.DraftStatusReleased, partID, id, ubom.DraftStatusDraft)
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

func isPostgresConstraint(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "foreign key constraint"))
}

func (s *PostgresStore) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var nextID int64
	if err := s.queryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM part_revisions").Scan(&nextID); err != nil {
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
	_, err = postgresTxExec(tx, "INSERT INTO part_revisions (part_number_id, revision, revision_seq_def_id) VALUES (?, ?, ?)", partNumberID, revision.Revision, revision.RevisionSeqDefID)
	if err != nil {
		tx.Rollback()
		if isPostgresConstraint(err) {
			return ubom.PartRevision{}, ErrAlreadyExists
		}
		return ubom.PartRevision{}, err
	}
	var id int64
	if err := postgresTxQueryRow(tx, "SELECT id FROM part_revisions WHERE part_number_id = ? AND revision = ?", partNumberID, revision.Revision).Scan(&id); err != nil {
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
		if _, err := postgresTxExec(tx, `INSERT INTO bom_line_items
			(parent_revision_id, position, child_part_number_id, child_revision_id)
			VALUES (?, ?, ?, ?)`, id, position, childPartID, childRevisionID); err != nil {
			tx.Rollback()
			if isPostgresConstraint(err) {
				return ubom.PartRevision{}, ErrNotFound
			}
			return ubom.PartRevision{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ubom.PartRevision{}, err
	}
	revision.ID = ubom.PartRevisionID(strconv.FormatInt(id, 10))
	return revision, nil
}

func (s *PostgresStore) GetPartRevision(id ubom.PartRevisionID) (ubom.PartRevision, error) {
	revisionID, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ubom.PartRevision{}, ErrNotFound
	}
	var partNumberID int64
	var revisionValue string
	var revisionSeqDefID string
	if err := s.queryRow("SELECT part_number_id, revision, revision_seq_def_id FROM part_revisions WHERE id = ?", revisionID).Scan(&partNumberID, &revisionValue, &revisionSeqDefID); err != nil {
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
	rows, err := s.query(`SELECT child_part_number_id, child_revision_id
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

func (s *PostgresStore) getPartNumberByID(id int64) (ubom.PartNumber, error) {
	var part ubom.PartNumber
	var attributeData []byte
	if err := s.queryRow(`SELECT id, value, seq_def_id, taxonomy_def_id, taxonomy_node_id, attributes, status
		FROM part_numbers WHERE id = ?`, id).Scan(
		&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID, &attributeData, &part.Status); err != nil {
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
