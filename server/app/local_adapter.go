package app

import (
	"context"
	"errors"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"ubom-v4/artifact"
)

// LocalFileAdapter resolves artifact paths beneath one configured directory.
type LocalFileAdapter struct {
	source artifact.Source
	root   string
}

func NewLocalFileAdapter(source artifact.Source, root string) (*LocalFileAdapter, error) {
	if err := source.Validate(); err != nil {
		return nil, err
	}
	if root == "" {
		return nil, errors.New("local adapter root is empty")
	}
	return &LocalFileAdapter{source: source, root: filepath.Clean(root)}, nil
}

func (a *LocalFileAdapter) Source() artifact.Source { return a.source }

func (a *LocalFileAdapter) Resolve(ctx context.Context, item artifact.Artifact, reference string) (artifact.Resolved, error) {
	if err := ctx.Err(); err != nil {
		return artifact.Resolved{}, err
	}
	if item.SourceID != a.source.ID {
		return artifact.Resolved{}, errors.New("artifact belongs to another source")
	}
	path := filepath.Join(a.root, filepath.Clean(item.Path))
	relative, err := filepath.Rel(a.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return artifact.Resolved{}, errors.New("artifact path escapes local source root")
	}
	file, err := os.Open(path)
	if err != nil {
		return artifact.Resolved{}, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return artifact.Resolved{}, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return artifact.Resolved{}, errors.New("artifact path is not a regular file")
	}
	mediaType := mime.TypeByExtension(filepath.Ext(path))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	return artifact.Resolved{Reference: reference, MediaType: mediaType, Size: info.Size(), Content: file}, nil
}
