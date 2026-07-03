package fs

import (
	"os"
	"time"
)

// Entry represents a file or directory in any backend.
type Entry struct {
	Name    string
	Path    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	Mode    os.FileMode
}
