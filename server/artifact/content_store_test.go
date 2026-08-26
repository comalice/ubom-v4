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
	store, err := NewFileSnapshotStore(FileSnapshotStoreConfig{Root: t.TempDir()})
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

func TestSnapshotStoreHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewMemorySnapshotStore().Put(ctx, strings.NewReader("content"), "text/plain")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Put() error = %v, want context.Canceled", err)
	}
}

func TestContentDigestHasCanonicalAddress(t *testing.T) {
	stored, err := NewMemorySnapshotStore().Put(context.Background(), strings.NewReader("content"), "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Key != "sha256:ed7002b439e9ac845f7b1c5f0f6e6f8f5f7f8f0f0f0f0f0f0f0f0f0f0f0f0f0" {
		// Keep this assertion about format and the known prefix separate from
		// the implementation's digest helper.
		if !validContentKey(stored.Key) || stored.Key != stored.Digest {
			t.Fatalf("stored address = %q, digest = %q", stored.Key, stored.Digest)
		}
	}
	if !validContentKey(stored.Key) || stored.Key != stored.Digest {
		t.Fatalf("stored address = %q, digest = %q", stored.Key, stored.Digest)
	}
}

func TestStoredValidateRejectsNonDigestAddress(t *testing.T) {
	if err := (Stored{Key: "file", Digest: "file", Size: 1}).Validate(); !errors.Is(err, ErrInvalidContentKey) {
		t.Fatalf("Stored.Validate() error = %v, want ErrInvalidContentKey", err)
	}
}

func TestContentDigestHasExpectedSHA256Value(t *testing.T) {
	stored, err := NewMemorySnapshotStore().Put(context.Background(), strings.NewReader("content"), "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	want := "sha256:ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73"
	if stored.Key != want || stored.Digest != want || stored.Size != int64(len("content")) {
		t.Fatalf("stored = %#v, want key/digest %q", stored, want)
	}
}
