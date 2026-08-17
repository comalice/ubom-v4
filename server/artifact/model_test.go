package artifact

import "testing"

func TestModelsValidate(t *testing.T) {
	source := Source{ID: "git-main", Name: "Main repository", Kind: "git"}
	if err := source.Validate(); err != nil {
		t.Fatalf("Source.Validate() error = %v", err)
	}

	item := Artifact{
		ID:       "drawing",
		Name:     "Widget drawing",
		Kind:     "drawing",
		SourceID: source.ID,
		Path:     "cad/widget.drw",
	}
	if err := item.Validate(); err != nil {
		t.Fatalf("Artifact.Validate() error = %v", err)
	}

	snapshot := Snapshot{
		ID:         "drawing-v1",
		ArtifactID: item.ID,
		Reference:  "v1.0.0",
		Commit:     "abc123",
		Digest:     "sha256:deadbeef",
		Size:       42,
	}
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("Snapshot.Validate() error = %v", err)
	}
}

func TestSnapshotValidateRequiresImmutableIdentity(t *testing.T) {
	snapshot := Snapshot{ID: "snapshot-1", ArtifactID: "artifact-1"}
	if err := snapshot.Validate(); err == nil {
		t.Fatal("Snapshot.Validate() accepted a snapshot without immutable identity")
	}
}

func TestSnapshotValidateRejectsNegativeSize(t *testing.T) {
	snapshot := Snapshot{
		ID:         "snapshot-1",
		ArtifactID: "artifact-1",
		Commit:     "abc123",
		Size:       -1,
	}
	if err := snapshot.Validate(); err == nil {
		t.Fatal("Snapshot.Validate() accepted a negative size")
	}
}
