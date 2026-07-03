package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LocalBackend reads and writes the local filesystem.
type LocalBackend struct {
	root string
}

// NewLocalBackend creates a backend rooted at path (defaults to home on empty).
func NewLocalBackend(path string) (*LocalBackend, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = home
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &LocalBackend{root: abs}, nil
}

// Root returns the current directory path.
func (b *LocalBackend) Root() string {
	return b.root
}

// SetRoot changes the current directory after validating it exists.
func (b *LocalBackend) SetRoot(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", abs)
	}
	b.root = abs
	return nil
}

// List returns directory entries sorted with directories first.
func (b *LocalBackend) List() ([]Entry, error) {
	entries, err := os.ReadDir(b.root)
	if err != nil {
		return nil, err
	}

	result := make([]Entry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, Entry{
			Name:    e.Name(),
			Path:    filepath.Join(b.root, e.Name()),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

// Parent returns the parent directory path, or empty if at filesystem root.
func (b *LocalBackend) Parent() string {
	parent := filepath.Dir(b.root)
	if parent == b.root {
		return ""
	}
	return parent
}

// Mkdir creates a directory under the current root.
func (b *LocalBackend) Mkdir(name string) error {
	return os.Mkdir(filepath.Join(b.root, name), 0o755)
}

// Remove deletes a file or empty directory.
func (b *LocalBackend) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll deletes a file or directory tree.
func (b *LocalBackend) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Rename renames a file or directory.
func (b *LocalBackend) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// Copy copies a file or directory tree to dest.
func (b *LocalBackend) Copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest)
}

// Move moves a file or directory.
func (b *LocalBackend) Move(src, dest string) error {
	return os.Rename(src, dest)
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func copyDir(src, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}
