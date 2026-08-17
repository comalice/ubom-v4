package artifact

import "fmt"

type SourceID string
type ID string
type SnapshotID string

// Source identifies the external system that provides artifacts.
type Source struct {
	ID   SourceID
	Name string
	Kind string
}

func (s Source) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("artifact source has no ID")
	}
	if s.Name == "" {
		return fmt.Errorf("artifact source %q has no name", s.ID)
	}
	if s.Kind == "" {
		return fmt.Errorf("artifact source %q has no kind", s.ID)
	}
	return nil
}

// Artifact identifies a logical item at a path within a source.
type Artifact struct {
	ID       ID
	Name     string
	Kind     string
	SourceID SourceID
	Path     string
}

func (a Artifact) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("artifact has no ID")
	}
	if a.Name == "" {
		return fmt.Errorf("artifact %q has no name", a.ID)
	}
	if a.Kind == "" {
		return fmt.Errorf("artifact %q has no kind", a.ID)
	}
	if a.SourceID == "" {
		return fmt.Errorf("artifact %q has no source ID", a.ID)
	}
	return nil
}

// Snapshot records an observed source state and optionally retained content.
type Snapshot struct {
	ID          SnapshotID
	ArtifactID  ID
	Reference   string
	Commit      string
	SnapshotKey string
	Digest      string
	MediaType   string
	Size        int64
}

func (s Snapshot) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("artifact snapshot has no ID")
	}
	if s.ArtifactID == "" {
		return fmt.Errorf("artifact snapshot %q has no artifact ID", s.ID)
	}
	if s.Commit == "" && s.Digest == "" && s.SnapshotKey == "" {
		return fmt.Errorf("artifact snapshot %q has no immutable identity", s.ID)
	}
	if s.Size < 0 {
		return fmt.Errorf("artifact snapshot %q has a negative size", s.ID)
	}
	return nil
}
