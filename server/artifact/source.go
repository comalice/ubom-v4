package artifact

import (
	"context"
	"io"
)

// Resolved contains the source facts and content for one artifact reference.
// The caller owns Content and must close it.
type Resolved struct {
	Reference string
	Commit    string
	MediaType string
	Size      int64
	Content   io.ReadCloser
}

// SourceAdapter resolves references and reads artifacts from a source system.
// It does not persist content or create UBOM domain records.
type SourceAdapter interface {
	Source() Source
	Resolve(context.Context, Artifact, string) (Resolved, error)
}
