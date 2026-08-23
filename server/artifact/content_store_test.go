package artifact

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMemorySnapshotStorePutOpenAndDeduplicates(t *testing.T) {
	store := NewMemorySnapshotStore()
	first, err := store.Put(context.Background(), strings.NewReader("content"), "text/plain")
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	second, err := store.Put(context.Background(), strings.NewReader("content"), "application/octet-stream")
	if err != nil {
		t.Fatalf("second Put() error = %v", err)
	}
	if first.Key != second.Key || first.Digest != second.Digest || first.Size != 7 || second.MediaType != "application/octet-stream" {
		t.Fatalf("stored metadata = %#v and %#v", first, second)
	}
	opened, err := store.Open(context.Background(), first.Key)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer opened.Close()
	data, err := io.ReadAll(opened)
	if err != nil || string(data) != "content" {
		t.Fatalf("opened content = %q, %v", data, err)
	}
	if _, err := store.Open(context.Background(), "missing"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing Open() error = %v, want os.ErrNotExist", err)
	}
}

func TestFileSnapshotStorePutOpen(t *testing.T) {
	store, err := NewFileSnapshotStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileSnapshotStore() error = %v", err)
	}
	stored, err := store.Put(context.Background(), strings.NewReader("file content"), "text/plain")
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	opened, err := store.Open(context.Background(), stored.Key)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	data, err := io.ReadAll(opened)
	closeErr := opened.Close()
	if err != nil || closeErr != nil || string(data) != "file content" {
		t.Fatalf("opened file = %q, read error %v, close error %v", data, err, closeErr)
	}
	if _, err := store.Open(context.Background(), "../outside"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path traversal Open() error = %v, want os.ErrNotExist", err)
	}
}

func TestSnapshotStoreRejectsOversizedContent(t *testing.T) {
	store := NewMemorySnapshotStore()
	_, err := store.Put(context.Background(), strings.NewReader(strings.Repeat("x", int(MaxSnapshotSize)+1)), "text/plain")
	if !errors.Is(err, ErrSnapshotTooLarge) {
		t.Fatalf("Put() error = %v, want ErrSnapshotTooLarge", err)
	}
}
