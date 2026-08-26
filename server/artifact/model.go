package artifact

import (
	"fmt"
	"strings"
)

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
	if strings.TrimSpace(string(s.ID)) == "" {
		return fmt.Errorf("%w: no ID", ErrInvalidSource)
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w %q: no name", ErrInvalidSource, s.ID)
	}
	if strings.TrimSpace(s.Kind) == "" {
		return fmt.Errorf("%w %q: no kind", ErrInvalidSource, s.ID)
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
	if strings.TrimSpace(string(a.ID)) == "" {
		return fmt.Errorf("%w: no ID", ErrInvalidArtifact)
	}
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("%w %q: no name", ErrInvalidArtifact, a.ID)
	}
	if strings.TrimSpace(a.Kind) == "" {
		return fmt.Errorf("%w %q: no kind", ErrInvalidArtifact, a.ID)
	}
	if strings.TrimSpace(string(a.SourceID)) == "" {
		return fmt.Errorf("%w %q: no source ID", ErrInvalidArtifact, a.ID)
	}
	if strings.TrimSpace(a.Path) == "" {
		return fmt.Errorf("%w %q: no source path", ErrInvalidArtifact, a.ID)
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
	if strings.TrimSpace(string(s.ID)) == "" {
		return fmt.Errorf("%w: no ID", ErrInvalidSnapshot)
	}
	if strings.TrimSpace(string(s.ArtifactID)) == "" {
		return fmt.Errorf("%w %q: no artifact ID", ErrInvalidSnapshot, s.ID)
	}
	if strings.TrimSpace(s.Commit) == "" && strings.TrimSpace(s.Digest) == "" && strings.TrimSpace(s.SnapshotKey) == "" {
		return fmt.Errorf("%w %q: no immutable identity", ErrInvalidSnapshot, s.ID)
	}
	if s.Size < 0 {
		return fmt.Errorf("%w %q: negative size", ErrInvalidSnapshot, s.ID)
	}
	return nil
}
