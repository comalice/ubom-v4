package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
)

type ArtifactEvent struct {
	Source     string
	ArtifactID string
	Reference  string
}

type ArtifactSource interface {
	EmitChangedArtifact(context.Context, EventSink) error
	Read(context.Context, string) ([]byte, error)
}

type EventSink interface {
	HandleChangedArtifact(context.Context, ArtifactEvent) error
}

type Snapshot struct {
	Digest  string
	content []byte
}

func (s Snapshot) Content() []byte {
	return append([]byte(nil), s.content...)
}

type SnapshotStore interface {
	Save(context.Context, Snapshot) error
	Get(context.Context, string) (Snapshot, error)
}

type SnapshotService interface {
	Capture(context.Context, ArtifactSource, ArtifactEvent) (Snapshot, error)
	Load(context.Context, string) (Snapshot, error)
}

type Normalizer interface {
	Normalize(context.Context, Snapshot) (NormalizationResult, error)
}

type NormalizationResult struct {
	Rows []BOMRow
}

type BOMRow struct {
	Parent string
	Child  string
	Qty    int
}

type fakeArtifactSource struct {
	content []byte
}

func (fakeArtifactSource) EmitChangedArtifact(ctx context.Context, sink EventSink) error {
	if sink == nil {
		return errors.New("artifact source: nil event sink")
	}
	event := ArtifactEvent{
		Source:     "fake-artifact-source",
		ArtifactID: "demo-widget",
		Reference:  "fake://demo-widget/revision-1",
	}
	return sink.HandleChangedArtifact(ctx, event)
}

func (s fakeArtifactSource) Read(_ context.Context, reference string) ([]byte, error) {
	if reference == "" {
		return nil, errors.New("artifact source: empty artifact reference")
	}
	if reference != "fake://demo-widget/revision-1" {
		return nil, fmt.Errorf("artifact source: unknown reference %q", reference)
	}
	if len(s.content) == 0 {
		return nil, errors.New("artifact source: empty artifact content")
	}
	return append([]byte(nil), s.content...), nil
}

type memorySnapshotStore struct {
	snapshots map[string]Snapshot
}

func newMemorySnapshotStore() *memorySnapshotStore {
	return &memorySnapshotStore{snapshots: make(map[string]Snapshot)}
}

func (s *memorySnapshotStore) Save(_ context.Context, snapshot Snapshot) error {
	if snapshot.Digest == "" {
		return errors.New("snapshot store: empty digest")
	}
	if len(snapshot.content) == 0 {
		return errors.New("snapshot store: empty content")
	}
	s.snapshots[snapshot.Digest] = Snapshot{
		Digest:  snapshot.Digest,
		content: append([]byte(nil), snapshot.content...),
	}
	return nil
}

func (s *memorySnapshotStore) Get(_ context.Context, digest string) (Snapshot, error) {
	snapshot, ok := s.snapshots[digest]
	if !ok {
		return Snapshot{}, fmt.Errorf("snapshot store: digest %q not found", digest)
	}
	snapshot.content = append([]byte(nil), snapshot.content...)
	return snapshot, nil
}

type demoSnapshotService struct {
	store SnapshotStore
}

func (s demoSnapshotService) Capture(ctx context.Context, source ArtifactSource, event ArtifactEvent) (Snapshot, error) {
	if s.store == nil {
		return Snapshot{}, errors.New("snapshot service: nil snapshot store")
	}
	if source == nil {
		return Snapshot{}, errors.New("snapshot service: nil artifact source")
	}
	content, err := source.Read(ctx, event.Reference)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot service: read %s: %w", event.ArtifactID, err)
	}
	if len(content) == 0 {
		return Snapshot{}, errors.New("snapshot service: source returned empty content")
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(content))
	snapshot := Snapshot{
		Digest:  digest,
		content: append([]byte(nil), content...),
	}
	if err := s.store.Save(ctx, snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot service: save snapshot: %w", err)
	}
	return snapshot, nil
}

func (s demoSnapshotService) Load(ctx context.Context, digest string) (Snapshot, error) {
	if s.store == nil {
		return Snapshot{}, errors.New("snapshot service: nil snapshot store")
	}
	return s.store.Get(ctx, digest)
}

type demoSink struct {
	source    ArtifactSource
	snapshots SnapshotService
	event     ArtifactEvent
	snapshot  Snapshot
}

func (s *demoSink) HandleChangedArtifact(ctx context.Context, event ArtifactEvent) error {
	if s.snapshots == nil {
		return errors.New("core/demo sink: nil snapshot service")
	}
	snapshot, err := s.snapshots.Capture(ctx, s.source, event)
	if err != nil {
		return fmt.Errorf("core/demo sink: capture %s: %w", event.ArtifactID, err)
	}
	s.event = event
	s.snapshot = snapshot
	return nil
}

type fakeSnapshotNormalizer struct{}

func (fakeSnapshotNormalizer) Normalize(_ context.Context, snapshot Snapshot) (NormalizationResult, error) {
	if snapshot.Digest == "" || len(snapshot.content) == 0 {
		return NormalizationResult{}, errors.New("fake snapshot normalizer: invalid snapshot")
	}
	return NormalizationResult{Rows: []BOMRow{
		{Parent: "demo-widget", Child: "demo-board", Qty: 1},
		{Parent: "demo-widget", Child: "demo-screw", Qty: 4},
	}}, nil
}

func run(ctx context.Context) error {
	source := fakeArtifactSource{content: []byte("part,directive\nwidget,root\n")}
	store := newMemorySnapshotStore()
	snapshots := demoSnapshotService{store: store}
	sink := &demoSink{source: source, snapshots: snapshots}

	if err := source.EmitChangedArtifact(ctx, sink); err != nil {
		return err
	}
	if sink.snapshot.Digest == "" {
		return errors.New("demo flow: source did not produce a snapshot")
	}
	normalizer := fakeSnapshotNormalizer{}
	result, err := normalizer.Normalize(ctx, sink.snapshot)
	if err != nil {
		return fmt.Errorf("normalize demo snapshot: %w", err)
	}

	fmt.Println("source: fake-artifact-source")
	fmt.Printf("event: changed artifact=%s\n", sink.event.ArtifactID)
	fmt.Printf("snapshot digest: %s\n", sink.snapshot.Digest)
	fmt.Println("normalizer: fake-snapshot-normalizer")
	fmt.Printf("result: %d BOM rows\n", len(result.Rows))
	for _, row := range result.Rows {
		fmt.Printf("  %s -> %s (qty=%d)\n", row.Parent, row.Child, row.Qty)
	}
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
