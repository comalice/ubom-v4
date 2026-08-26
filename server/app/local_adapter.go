package app

import (
	"context"
	"errors"
	"fmt"
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

type LocalFileAdapterConfig struct {
	Source artifact.Source
	Root   string
}

func NewLocalFileAdapter(config LocalFileAdapterConfig) (*LocalFileAdapter, error) {
	if err := config.Source.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.Root) == "" {
		return nil, fmt.Errorf("local adapter root is empty")
	}
	absoluteRoot, err := filepath.Abs(config.Root)
	if err != nil {
		return nil, err
	}
	realRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve local adapter root: %w", err)
	}
	info, err := os.Stat(realRoot)
	if err != nil {
		return nil, fmt.Errorf("stat local adapter root: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("local adapter root is not a directory")
	}
	return &LocalFileAdapter{source: config.Source, root: realRoot}, nil
}

func (a *LocalFileAdapter) Source() artifact.Source { return a.source }

func (a *LocalFileAdapter) Resolve(ctx context.Context, item artifact.Artifact, reference string) (artifact.Resolved, error) {
	if err := ctx.Err(); err != nil {
		return artifact.Resolved{}, err
	}
	if item.SourceID != a.source.ID {
		return artifact.Resolved{}, artifact.ErrSourceOwnership
	}
	if strings.TrimSpace(reference) == "" {
		return artifact.Resolved{}, artifact.ErrInvalidReference
	}
	if err := item.Validate(); err != nil {
		return artifact.Resolved{}, err
	}
	if filepath.IsAbs(item.Path) {
		return artifact.Resolved{}, artifact.ErrPathEscape
	}
	path := filepath.Join(a.root, filepath.Clean(item.Path))
	relative, err := filepath.Rel(a.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return artifact.Resolved{}, artifact.ErrPathEscape
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return artifact.Resolved{}, fmt.Errorf("resolve artifact path: %w", err)
	}
	realRelative, err := filepath.Rel(a.root, realPath)
	if err != nil || realRelative == ".." || strings.HasPrefix(realRelative, ".."+string(filepath.Separator)) {
		return artifact.Resolved{}, artifact.ErrPathEscape
	}
	file, err := os.Open(realPath)
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
	if info.Size() > artifact.MaxSnapshotSize {
		file.Close()
		return artifact.Resolved{}, artifact.ErrSnapshotTooLarge
	}
	if err := ctx.Err(); err != nil {
		file.Close()
		return artifact.Resolved{}, err
	}
	mediaType := mime.TypeByExtension(filepath.Ext(path))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	return artifact.Resolved{Reference: reference, MediaType: mediaType, Size: info.Size(), Content: file}, nil
}
