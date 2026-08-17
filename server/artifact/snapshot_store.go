package artifact

import (
	"context"
	"io"
)

// Stored describes content retained by a SnapshotStore.
type Stored struct {
	Key       string
	Digest    string
	MediaType string
	Size      int64
}

// SnapshotStore retains and reopens artifact content. It does not persist
// artifact or snapshot domain records.
type SnapshotStore interface {
	Put(context.Context, io.Reader, string) (Stored, error)
	Open(context.Context, string) (io.ReadCloser, error)
}
