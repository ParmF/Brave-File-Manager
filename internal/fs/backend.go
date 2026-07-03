package fs

// Backend is implemented by local and remote file sources.
type Backend interface {
	Root() string
	SetRoot(path string) error
	List() ([]Entry, error)
	Parent() string
	Mkdir(name string) error
	Remove(path string) error
	RemoveAll(path string) error
	Rename(oldPath, newPath string) error
}
