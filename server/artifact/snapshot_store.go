package artifact

import (
	"context"
	"fmt"
	"io"
)

// Stored describes content retained by a SnapshotStore.
type Stored struct {
	Key       string
	Digest    string
	MediaType string
	Size      int64
}

func (s Stored) Validate() error {
	if !validContentKey(s.Key) || s.Key != s.Digest {
		return fmt.Errorf("%w: key and digest do not identify the same SHA-256 content", ErrInvalidContentKey)
	}
	if s.Size < 0 || s.Size > MaxSnapshotSize {
		return fmt.Errorf("%w: invalid stored size %d", ErrSnapshotTooLarge, s.Size)
	}
	return nil
}

// SnapshotStore retains and reopens artifact content. It does not persist
// artifact or snapshot domain records.
type SnapshotStore interface {
	Put(context.Context, io.Reader, string) (Stored, error)
	Open(context.Context, string) (io.ReadCloser, error)
}
