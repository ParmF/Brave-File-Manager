package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asnasn/Brave-File-Manager/internal/fs"
)

func TestLocalBackendListAndOperations(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	backend, err := fs.NewLocalBackend(root)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := backend.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].IsDir {
		t.Fatalf("expected directory first, got %q", entries[0].Name)
	}

	if err := backend.Mkdir("newdir"); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(root, "a.txt")
	newPath := filepath.Join(root, "b.txt")
	if err := backend.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatal(err)
	}
}
