# Brave File Manager

A powerful, elegant, and open-source file manager for macOS (Apple Silicon & Intel) written in Go. Features include:

- Native macOS support (ARM & Intel)
- Robust ext4 filesystem read/write support (via ext4fuse on macOS)
- SSH/SFTP and FTP/FTPS client capabilities
- Minimalist, modern UI (Fyne)
- Comprehensive file operations (local & remote)
- Efficient search and filtering
- Streamlined navigation with sidebar
- Performance optimized for large directories
- 100% open source, clean and well-documented

## Getting Started

1. **Install Go** (https://golang.org/dl/)
2. **Install Fyne dependencies** (see https://developer.fyne.io/started/)
3. **Clone and build:**
   ```bash
   git clone https://github.com/asnasn/Brave-File-Manager.git
   cd Brave-File-Manager
   go run .
   ```

## Usage

- **Local**: Browse your filesystem from the sidebar. Use the path bar to jump to any directory.
- **SFTP / FTP**: Select from the sidebar and enter connection details in the dialog.
- **Ext4**: On macOS with macFUSE and `ext4fuse` installed, mount ext4 volumes from the Ext4 sidebar section.
- **Search**: Type in the search bar to filter the current listing; local searches fall back to recursive matching when no direct hits are found.
- **Operations**: Create folders, rename, delete, copy paths, and navigate with Up/Open from the toolbar.

## Project layout

```
internal/
  fs/       Local filesystem backend and shared entry types
  remote/   SFTP and FTP/FTPS clients
  search/   Filtering and recursive search
  ext4/     macOS ext4 mount helpers (ext4fuse)
  ui/       Fyne application shell
```

## Roadmap

- [x] Project scaffold & UI skeleton
- [x] Local file operations
- [x] SSH/SFTP client
- [x] FTP/FTPS client
- [x] ext4 read/write support (macOS via ext4fuse)
- [x] Search & filtering
- [x] Sidebar navigation
- [ ] Performance tuning
- [ ] Documentation & community

## Development

Run unit tests (no GUI required):

```bash
go test ./internal/...
```

Build for macOS:

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o bravefm .
```

Create a local `.app` bundle and DMG (requires [Fyne CLI](https://docs.fyne.io/started/install.html)):

```bash
go install fyne.io/fyne/v2/cmd/fyne@v2.6.1
fyne package -os darwin -arch arm64 -name "Brave File Manager" -release
hdiutil create -volname "Brave File Manager" -srcfolder "Brave File Manager.app" -ov -format UDZO BraveFileManager.dmg
```

## Releases

Tagged releases (`v*.*.*`) build unsigned macOS DMGs for Apple Silicon (`arm64`) and Intel (`amd64`) and attach them to the GitHub release page automatically.

## License

MIT
