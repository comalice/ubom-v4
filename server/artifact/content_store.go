package artifact

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MaxSnapshotSize int64 = 64 << 20

var ErrSnapshotTooLarge = errors.New("artifact snapshot exceeds maximum size")

// MemorySnapshotStore is a content-only SnapshotStore for tests and ephemeral use.
type MemorySnapshotStore struct {
	contents map[string][]byte
}

func NewMemorySnapshotStore() *MemorySnapshotStore {
	return &MemorySnapshotStore{contents: map[string][]byte{}}
}

func (s *MemorySnapshotStore) Put(ctx context.Context, r io.Reader, mediaType string) (Stored, error) {
	data, err := readContent(ctx, r)
	if err != nil {
		return Stored{}, err
	}
	digest := contentDigest(data)
	if _, ok := s.contents[digest]; !ok {
		s.contents[digest] = data
	}
	return Stored{Key: digest, Digest: digest, MediaType: mediaType, Size: int64(len(data))}, nil
}

func (s *MemorySnapshotStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validContentKey(key) {
		return nil, os.ErrNotExist
	}
	data, ok := s.contents[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(append([]byte(nil), data...))), nil
}

// FileSnapshotStore stores content under its SHA-256 digest. Files are written
// through a temporary file and renamed, so readers never see partial content.
type FileSnapshotStore struct {
	root string
}

func NewFileSnapshotStore(root string) (*FileSnapshotStore, error) {
	if root == "" {
		return nil, errors.New("snapshot store root is empty")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &FileSnapshotStore{root: root}, nil
}

func (s *FileSnapshotStore) Put(ctx context.Context, r io.Reader, mediaType string) (Stored, error) {
	data, err := readContent(ctx, r)
	if err != nil {
		return Stored{}, err
	}
	digest := contentDigest(data)
	path := filepath.Join(s.root, digest)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		tmp, err := os.CreateTemp(s.root, ".snapshot-*")
		if err != nil {
			return Stored{}, err
		}
		tmpName := tmp.Name()
		defer os.Remove(tmpName)
		if _, err := tmp.Write(data); err != nil {
			tmp.Close()
			return Stored{}, err
		}
		if err := tmp.Close(); err != nil {
			return Stored{}, err
		}
		if err := os.Rename(tmpName, path); err != nil && !errors.Is(err, os.ErrExist) {
			return Stored{}, err
		}
	} else if err != nil {
		return Stored{}, err
	}
	return Stored{Key: digest, Digest: digest, MediaType: mediaType, Size: int64(len(data))}, nil
}

func (s *FileSnapshotStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validContentKey(key) {
		return nil, os.ErrNotExist
	}
	return os.Open(filepath.Join(s.root, key))
}

func readContent(ctx context.Context, r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errors.New("snapshot content is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(contextReader{ctx: ctx, reader: r}, MaxSnapshotSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > MaxSnapshotSize {
		return nil, ErrSnapshotTooLarge
	}
	return data, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func validContentKey(key string) bool {
	return strings.HasPrefix(key, "sha256:") && len(key) == len("sha256:")+64 && isHex(key[len("sha256:"):])
}

func isHex(value string) bool {
	for _, character := range value {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

func contentDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
