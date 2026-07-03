package search_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asnasn/Brave-File-Manager/internal/fs"
	"github.com/asnasn/Brave-File-Manager/internal/search"
)

func TestFilterAndWalk(t *testing.T) {
	entries := []fs.Entry{
		{Name: "Notes.txt"},
		{Name: "photo.jpg"},
	}

	filtered := search.Filter(entries, "note")
	if len(filtered) != 1 || filtered[0].Name != "Notes.txt" {
		t.Fatalf("unexpected filter result: %+v", filtered)
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "ignored"), 0o755); err != nil {
		t.Fatal(err)
	}

	results, err := search.Walk(root, "target", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "target.log" {
		t.Fatalf("unexpected walk result: %+v", results)
	}
}
