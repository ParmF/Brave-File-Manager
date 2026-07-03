package search

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/asnasn/Brave-File-Manager/internal/fs"
)

// Filter returns entries whose names contain query (case-insensitive).
func Filter(entries []fs.Entry, query string) []fs.Entry {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return entries
	}
	filtered := make([]fs.Entry, 0, len(entries))
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), query) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Walk finds files and directories under root whose names contain query.
func Walk(root, query string, maxResults int) ([]fs.Entry, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return nil, nil
	}
	if maxResults <= 0 {
		maxResults = 500
	}

	var results []fs.Entry
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if len(results) >= maxResults {
			return filepath.SkipAll
		}
		if !strings.Contains(strings.ToLower(d.Name()), query) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		results = append(results, fs.Entry{
			Name:    d.Name(),
			Path:    path,
			Size:    info.Size(),
			IsDir:   d.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
		return nil
	})
	return results, err
}
