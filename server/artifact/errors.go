package artifact

import "errors"

var (
	ErrInvalidSource       = errors.New("invalid artifact source")
	ErrInvalidArtifact     = errors.New("invalid artifact")
	ErrInvalidSnapshot     = errors.New("invalid artifact snapshot")
	ErrInvalidReference    = errors.New("artifact reference is empty")
	ErrSourceOwnership     = errors.New("artifact belongs to another source")
	ErrPathEscape          = errors.New("artifact path escapes local source root")
	ErrInvalidContentKey   = errors.New("invalid artifact content key")
	ErrSnapshotContentNil  = errors.New("source returned no artifact content")
	ErrSnapshotStoreConfig = errors.New("invalid snapshot store configuration")
)
