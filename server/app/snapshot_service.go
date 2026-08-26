package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"ubom-v4/artifact"
	"ubom-v4/store"
)

type SnapshotService struct {
	metadata store.Store
	content  artifact.SnapshotStore
	adapters map[artifact.SourceID]artifact.SourceAdapter
}

func NewSnapshotService(metadata store.Store, content artifact.SnapshotStore, adapters ...artifact.SourceAdapter) *SnapshotService {
	bySource := make(map[artifact.SourceID]artifact.SourceAdapter, len(adapters))
	for _, adapter := range adapters {
		if adapter == nil {
			continue
		}
		bySource[adapter.Source().ID] = adapter
	}
	return &SnapshotService{metadata: metadata, content: content, adapters: bySource}
}

func (s *SnapshotService) Capture(ctx context.Context, artifactID artifact.ID, reference string) (artifact.Snapshot, error) {
	if s == nil || s.metadata == nil {
		return artifact.Snapshot{}, errors.New("artifact metadata store is not configured")
	}
	if s.content == nil {
		return artifact.Snapshot{}, errors.New("artifact snapshot store is not configured")
	}
	if err := ctx.Err(); err != nil {
		return artifact.Snapshot{}, err
	}
	if strings.TrimSpace(reference) == "" {
		return artifact.Snapshot{}, artifact.ErrInvalidReference
	}
	item, err := s.metadata.GetArtifact(artifactID)
	if err != nil {
		return artifact.Snapshot{}, err
	}
	adapter, ok := s.adapters[item.SourceID]
	if !ok {
		return artifact.Snapshot{}, fmt.Errorf("no adapter for source %q", item.SourceID)
	}
	if adapter.Source().ID != item.SourceID {
		return artifact.Snapshot{}, fmt.Errorf("adapter source %q does not own artifact source %q", adapter.Source().ID, item.SourceID)
	}
	resolved, err := adapter.Resolve(ctx, item, reference)
	if err != nil {
		return artifact.Snapshot{}, err
	}
	if resolved.Content == nil {
		return artifact.Snapshot{}, artifact.ErrSnapshotContentNil
	}
	if strings.TrimSpace(resolved.Reference) == "" {
		resolved.Content.Close()
		return artifact.Snapshot{}, artifact.ErrInvalidReference
	}
	defer resolved.Content.Close()
	stored, err := s.content.Put(ctx, resolved.Content, resolved.MediaType)
	if err != nil {
		return artifact.Snapshot{}, err
	}
	if err := stored.Validate(); err != nil {
		return artifact.Snapshot{}, fmt.Errorf("invalid snapshot store result: %w", err)
	}
	snapshot := artifact.Snapshot{
		ID: snapshotID(artifactID, resolved.Reference, stored.Digest), ArtifactID: artifactID,
		Reference: resolved.Reference, Commit: resolved.Commit, SnapshotKey: stored.Key,
		Digest: stored.Digest, MediaType: stored.MediaType, Size: stored.Size,
	}
	if existing, err := s.metadata.GetSnapshot(snapshot.ID); err == nil {
		return existing, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return artifact.Snapshot{}, err
	}
	if err := s.metadata.CreateSnapshot(snapshot); err != nil {
		if existing, getErr := s.metadata.GetSnapshot(snapshot.ID); getErr == nil {
			return existing, nil
		}
		return artifact.Snapshot{}, err
	}
	return snapshot, nil
}

func snapshotID(artifactID artifact.ID, reference, digest string) artifact.SnapshotID {
	value := string(artifactID) + "\x00" + reference + "\x00" + digest
	sum := sha256.Sum256([]byte(value))
	return artifact.SnapshotID("snapshot-" + hex.EncodeToString(sum[:]))
}
