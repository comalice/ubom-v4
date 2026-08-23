package app

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"ubom-v4/artifact"
	"ubom-v4/store"
)

func TestSnapshotServiceCapturesLocalFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "widget.txt"), []byte("widget"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	metadata := store.NewMemoryStore()
	source := artifact.Source{ID: "local", Name: "Local", Kind: "filesystem"}
	item := artifact.Artifact{ID: "widget", Name: "Widget", Kind: "text", SourceID: source.ID, Path: "widget.txt"}
	if err := metadata.CreateSource(source); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateArtifact(item); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewLocalFileAdapter(source, root)
	if err != nil {
		t.Fatal(err)
	}
	content := artifact.NewMemorySnapshotStore()
	service := NewSnapshotService(metadata, content, adapter)
	snapshot, err := service.Capture(context.Background(), item.ID, "working")
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	repeated, err := service.Capture(context.Background(), item.ID, "working")
	if err != nil {
		t.Fatalf("repeated Capture() error = %v", err)
	}
	if repeated != snapshot || snapshot.Size != 6 || snapshot.Digest == "" || snapshot.SnapshotKey == "" {
		t.Fatalf("snapshot = %#v, repeated = %#v", snapshot, repeated)
	}
	opened, err := content.Open(context.Background(), snapshot.SnapshotKey)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	data, readErr := io.ReadAll(opened)
	closeErr := opened.Close()
	if readErr != nil || closeErr != nil || string(data) != "widget" {
		t.Fatalf("content = %q, read error %v, close error %v", data, readErr, closeErr)
	}
}

func TestLocalFileAdapterRejectsPathEscape(t *testing.T) {
	source := artifact.Source{ID: "local", Name: "Local", Kind: "filesystem"}
	adapter, err := NewLocalFileAdapter(source, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = adapter.Resolve(context.Background(), artifact.Artifact{ID: "x", Name: "x", Kind: "file", SourceID: source.ID, Path: "../secret"}, "ref")
	if err == nil {
		t.Fatal("Resolve() accepted a path outside the source root")
	}
}

func TestLocalFileAdapterRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	source := artifact.Source{ID: "local", Name: "Local", Kind: "filesystem"}
	adapter, err := NewLocalFileAdapter(source, root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = adapter.Resolve(context.Background(), artifact.Artifact{ID: "x", Name: "x", Kind: "file", SourceID: source.ID, Path: "link/secret.txt"}, "ref")
	if err == nil {
		t.Fatal("Resolve() accepted a symlink path outside the source root")
	}
}

func TestSnapshotServiceRejectsEmptyReference(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "widget.txt"), []byte("widget"), 0o644); err != nil {
		t.Fatal(err)
	}
	metadata := store.NewMemoryStore()
	source := artifact.Source{ID: "local", Name: "Local", Kind: "filesystem"}
	item := artifact.Artifact{ID: "widget", Name: "Widget", Kind: "text", SourceID: source.ID, Path: "widget.txt"}
	if err := metadata.CreateSource(source); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateArtifact(item); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewLocalFileAdapter(source, root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewSnapshotService(metadata, artifact.NewMemorySnapshotStore(), adapter).Capture(context.Background(), item.ID, "")
	if err == nil {
		t.Fatal("Capture() accepted an empty reference")
	}
}
