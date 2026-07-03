package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/asnasn/Brave-File-Manager/internal/ext4"
	"github.com/asnasn/Brave-File-Manager/internal/fs"
	"github.com/asnasn/Brave-File-Manager/internal/remote"
	"github.com/asnasn/Brave-File-Manager/internal/search"
)

type locationKind int

const (
	locationLocal locationKind = iota
	locationSFTP
	locationFTP
	locationExt4
)

type location struct {
	kind  locationKind
	label string
}

// App is the main Brave File Manager window controller.
type App struct {
	window   fyne.Window
	backend  fs.Backend
	entries  []fs.Entry
	filtered []fs.Entry
	selected int

	sftpClient *remote.SFTPClient
	ftpClient  *remote.FTPClient

	pathEntry *widget.Entry
	searchEntry *widget.Entry
	statusLabel *widget.Label
	fileList    *widget.List
	locations   []location
}

// NewApp creates and wires the main application UI.
func NewApp(w fyne.Window) (*App, error) {
	local, err := fs.NewLocalBackend("")
	if err != nil {
		return nil, err
	}

	a := &App{
		window:  w,
		backend: local,
		locations: []location{
			{kind: locationLocal, label: "Local"},
			{kind: locationSFTP, label: "SFTP"},
			{kind: locationFTP, label: "FTP"},
			{kind: locationExt4, label: "Ext4"},
		},
		selected: -1,
	}
	a.buildUI()
	a.refresh()
	return a, nil
}

func (a *App) buildUI() {
	a.pathEntry = widget.NewEntry()
	a.pathEntry.OnSubmitted = func(path string) {
		if err := a.backend.SetRoot(path); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.refresh()
	}

	a.searchEntry = widget.NewEntry()
	a.searchEntry.SetPlaceHolder("Filter or search files...")
	a.searchEntry.OnChanged = func(_ string) {
		a.applyFilter()
	}

	a.statusLabel = widget.NewLabel("")
	a.fileList = widget.NewList(
		func() int { return len(a.filtered) },
		func() fyne.CanvasObject {
			name := widget.NewLabel("")
			meta := widget.NewLabel("")
			return container.NewBorder(nil, nil, nil, meta, name)
		},
		func(i int, obj fyne.CanvasObject) {
			if i < 0 || i >= len(a.filtered) {
				return
			}
			entry := a.filtered[i]
			border := obj.(*fyne.Container)
			name := border.Objects[0].(*widget.Label)
			meta := border.Objects[1].(*widget.Label)

			prefix := "📄 "
			if entry.IsDir {
				prefix = "📁 "
			}
			name.SetText(prefix + entry.Name)
			meta.SetText(formatEntryMeta(entry))
		},
	)
	a.fileList.OnSelected = func(id int) { a.selected = id }
	a.fileList.OnUnselected = func(_ int) { a.selected = -1 }

	toolbar := container.NewBorder(nil, nil,
		widget.NewButton("Up", a.goUp),
		container.NewHBox(
			widget.NewButton("Open", a.openSelected),
			widget.NewButton("Refresh", a.refresh),
			widget.NewButton("New Folder", a.newFolder),
			widget.NewButton("Rename", a.renameSelected),
			widget.NewButton("Delete", a.deleteSelected),
			widget.NewButton("Copy Path", a.copyPath),
		),
		a.pathEntry,
	)

	searchBar := container.NewBorder(nil, nil, widget.NewLabel("Search"), nil, a.searchEntry)

	sidebar := widget.NewList(
		func() int { return len(a.locations) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(a.locations[i].label)
		},
	)
	sidebar.OnSelected = func(id int) {
		a.switchLocation(a.locations[id].kind)
	}

	content := container.NewBorder(
		container.NewVBox(toolbar, searchBar, a.statusLabel),
		nil, nil, nil,
		a.fileList,
	)

	a.window.SetContent(container.NewHSplit(sidebar, content))
	a.window.Resize(fyne.NewSize(1000, 650))
}

func (a *App) switchLocation(kind locationKind) {
	switch kind {
	case locationLocal:
		a.disconnectRemote()
		local, err := fs.NewLocalBackend(a.backend.Root())
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.backend = local
		a.refresh()
	case locationSFTP:
		a.showSFTPDialog()
	case locationFTP:
		a.showFTPDialog()
	case locationExt4:
		a.showExt4Dialog()
	}
}

func (a *App) disconnectRemote() {
	if a.sftpClient != nil {
		a.sftpClient.Close()
		a.sftpClient = nil
	}
	if a.ftpClient != nil {
		a.ftpClient.Close()
		a.ftpClient = nil
	}
}

func (a *App) showSFTPDialog() {
	host := widget.NewEntry()
	host.SetPlaceHolder("host.example.com")
	port := widget.NewEntry()
	port.SetText("22")
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()

	form := dialog.NewForm("Connect SFTP", "Connect", "Cancel", []*widget.FormItem{
		{Text: "Host", Widget: host},
		{Text: "Port", Widget: port},
		{Text: "User", Widget: user},
		{Text: "Password", Widget: pass},
	}, func(ok bool) {
		if !ok {
			return
		}
		client, err := remote.ConnectSFTP(remote.SFTPConfig{
			Host:     strings.TrimSpace(host.Text),
			Port:     strings.TrimSpace(port.Text),
			User:     strings.TrimSpace(user.Text),
			Password: pass.Text,
		})
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.disconnectRemote()
		a.sftpClient = client
		a.backend = client
		a.refresh()
	}, a.window)
	form.Resize(fyne.NewSize(420, 280))
	form.Show()
}

func (a *App) showFTPDialog() {
	host := widget.NewEntry()
	host.SetPlaceHolder("host.example.com")
	port := widget.NewEntry()
	port.SetText("21")
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	useTLS := widget.NewCheck("Use FTPS (TLS)", nil)

	form := dialog.NewForm("Connect FTP", "Connect", "Cancel", []*widget.FormItem{
		{Text: "Host", Widget: host},
		{Text: "Port", Widget: port},
		{Text: "User", Widget: user},
		{Text: "Password", Widget: pass},
		{Text: "", Widget: useTLS},
	}, func(ok bool) {
		if !ok {
			return
		}
		client, err := remote.ConnectFTP(remote.FTPConfig{
			Host:     strings.TrimSpace(host.Text),
			Port:     strings.TrimSpace(port.Text),
			User:     strings.TrimSpace(user.Text),
			Password: pass.Text,
			UseTLS:   useTLS.Checked,
		})
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.disconnectRemote()
		a.ftpClient = client
		a.backend = client
		a.refresh()
	}, a.window)
	form.Resize(fyne.NewSize(420, 300))
	form.Show()
}

func (a *App) showExt4Dialog() {
	if !ext4.IsSupported() {
		dialog.ShowInformation("Ext4 Support", ext4.Requirements(), a.window)
		return
	}

	device := widget.NewEntry()
	device.SetPlaceHolder("/dev/disk2s1 or /path/to/image.img")
	mountPoint := widget.NewEntry()
	mountPoint.SetPlaceHolder("/Volumes/my-ext4")
	readOnly := widget.NewCheck("Read only", nil)

	form := dialog.NewForm("Mount ext4 Volume", "Mount", "Cancel", []*widget.FormItem{
		{Text: "Device", Widget: device},
		{Text: "Mount point", Widget: mountPoint},
		{Text: "", Widget: readOnly},
	}, func(ok bool) {
		if !ok {
			return
		}
		dev := strings.TrimSpace(device.Text)
		mp := strings.TrimSpace(mountPoint.Text)
		if mp == "" {
			mp = ext4.DefaultMountPoint(dev)
		}
		if _, err := ext4.Mount(dev, mp, readOnly.Checked); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.disconnectRemote()
		local, err := fs.NewLocalBackend(mp)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.backend = local
		a.refresh()
		dialog.ShowInformation("Ext4 Mounted", fmt.Sprintf("Mounted at %s", mp), a.window)
	}, a.window)
	form.Resize(fyne.NewSize(480, 280))
	form.Show()
}

func (a *App) refresh() {
	entries, err := a.backend.List()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.entries = entries
	a.pathEntry.SetText(a.backend.Root())
	a.applyFilter()
}

func (a *App) applyFilter() {
	query := strings.TrimSpace(a.searchEntry.Text)
	if query == "" {
		a.filtered = a.entries
		a.statusLabel.SetText(fmt.Sprintf("%d items", len(a.filtered)))
		a.fileList.Refresh()
		return
	}

	filtered := search.Filter(a.entries, query)
	if len(filtered) == 0 && !strings.Contains(query, " ") {
		// Deep search for local backend only when filter finds nothing.
		if local, ok := a.backend.(*fs.LocalBackend); ok {
			if results, err := search.Walk(local.Root(), query, 200); err == nil && len(results) > 0 {
				filtered = results
				a.statusLabel.SetText(fmt.Sprintf("%d matches (recursive)", len(filtered)))
				a.filtered = filtered
				a.fileList.Refresh()
				return
			}
		}
	}

	a.filtered = filtered
	a.statusLabel.SetText(fmt.Sprintf("%d of %d items", len(a.filtered), len(a.entries)))
	a.fileList.Refresh()
}

func (a *App) goUp() {
	parent := a.backend.Parent()
	if parent == "" {
		return
	}
	if err := a.backend.SetRoot(parent); err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.refresh()
}

func (a *App) openSelected() {
	if a.selected < 0 || a.selected >= len(a.filtered) {
		return
	}
	entry := a.filtered[a.selected]
	if !entry.IsDir {
		return
	}
	if err := a.backend.SetRoot(entry.Path); err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.refresh()
}

func (a *App) newFolder() {
	entry := widget.NewEntry()
	dialog.ShowForm("New Folder", "Create", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: entry},
	}, func(ok bool) {
		if !ok || strings.TrimSpace(entry.Text) == "" {
			return
		}
		if err := a.backend.Mkdir(strings.TrimSpace(entry.Text)); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.refresh()
	}, a.window)
}

func (a *App) renameSelected() {
	if a.selected < 0 || a.selected >= len(a.filtered) {
		dialog.ShowInformation("Rename", "Select a file or folder first.", a.window)
		return
	}
	target := a.filtered[a.selected]
	entry := widget.NewEntry()
	entry.SetText(target.Name)

	dialog.ShowForm("Rename", "Rename", "Cancel", []*widget.FormItem{
		{Text: "New name", Widget: entry},
	}, func(ok bool) {
		if !ok || strings.TrimSpace(entry.Text) == "" {
			return
		}
		newName := strings.TrimSpace(entry.Text)
		newPath := filepath.Join(filepath.Dir(target.Path), newName)
		if err := a.backend.Rename(target.Path, newPath); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.refresh()
	}, a.window)
}

func (a *App) deleteSelected() {
	if a.selected < 0 || a.selected >= len(a.filtered) {
		dialog.ShowInformation("Delete", "Select a file or folder first.", a.window)
		return
	}
	target := a.filtered[a.selected]
	dialog.ShowConfirm("Delete", fmt.Sprintf("Delete %q?", target.Name), func(ok bool) {
		if !ok {
			return
		}
		if err := a.backend.RemoveAll(target.Path); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.selected = -1
		a.refresh()
	}, a.window)
}

func (a *App) copyPath() {
	if a.selected < 0 || a.selected >= len(a.filtered) {
		a.window.Clipboard().SetContent(a.backend.Root())
		return
	}
	a.window.Clipboard().SetContent(a.filtered[a.selected].Path)
}

func formatEntryMeta(entry fs.Entry) string {
	if entry.IsDir {
		return "folder"
	}
	return fmt.Sprintf("%s  %s", humanSize(entry.Size), entry.ModTime.Format("2006-01-02 15:04"))
}

func humanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// RunOnClose cleans up remote connections when the window closes.
func (a *App) RunOnClose() {
	a.window.SetCloseIntercept(func() {
		a.disconnectRemote()
		a.window.Close()
	})
}
