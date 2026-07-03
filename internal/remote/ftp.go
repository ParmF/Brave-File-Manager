package remote

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"

	"github.com/asnasn/Brave-File-Manager/internal/fs"
)

// FTPClient provides access over FTP/FTPS.
type FTPClient struct {
	host     string
	user     string
	password string
	port     string
	useTLS   bool
	client   *ftp.ServerConn
	cwd      string
}

// FTPConfig holds connection parameters.
type FTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	UseTLS   bool
}

// ConnectFTP opens an FTP or FTPS session.
func ConnectFTP(cfg FTPConfig) (*FTPClient, error) {
	if cfg.Port == "" {
		if cfg.UseTLS {
			cfg.Port = "990"
		} else {
			cfg.Port = "21"
		}
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var conn *ftp.ServerConn
	var err error

	if cfg.UseTLS {
		conn, err = ftp.Dial(addr,
			ftp.DialWithTimeout(10*time.Second),
			ftp.DialWithExplicitTLS(&tls.Config{
				//nolint:gosec // user-provided server certificates
				InsecureSkipVerify: true,
			}),
		)
	} else {
		conn, err = ftp.Dial(addr, ftp.DialWithTimeout(10*time.Second))
	}
	if err != nil {
		return nil, fmt.Errorf("ftp dial: %w", err)
	}

	if err := conn.Login(cfg.User, cfg.Password); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("ftp login: %w", err)
	}

	c := &FTPClient{
		host:     cfg.Host,
		user:     cfg.User,
		password: cfg.Password,
		port:     cfg.Port,
		useTLS:   cfg.UseTLS,
		client:   conn,
		cwd:      "/",
	}
	if pwd, err := conn.CurrentDir(); err == nil {
		c.cwd = pwd
	}
	return c, nil
}

// Close closes the FTP connection.
func (c *FTPClient) Close() error {
	if c.client != nil {
		return c.client.Quit()
	}
	return nil
}

// Root returns the current remote directory.
func (c *FTPClient) Root() string {
	return c.cwd
}

// SetRoot changes the remote working directory.
func (c *FTPClient) SetRoot(dir string) error {
	if err := c.client.ChangeDir(dir); err != nil {
		return err
	}
	pwd, err := c.client.CurrentDir()
	if err != nil {
		return err
	}
	c.cwd = pwd
	return nil
}

// List returns entries in the current remote directory.
func (c *FTPClient) List() ([]fs.Entry, error) {
	entries, err := c.client.List(c.cwd)
	if err != nil {
		return nil, err
	}

	result := make([]fs.Entry, 0, len(entries))
	for _, e := range entries {
		isDir := e.Type == ftp.EntryTypeFolder
		result = append(result, fs.Entry{
			Name:    e.Name,
			Path:    path.Join(c.cwd, e.Name),
			Size:    int64(e.Size),
			IsDir:   isDir,
			ModTime: e.Time,
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

// Parent returns the parent directory path.
func (c *FTPClient) Parent() string {
	if c.cwd == "/" || c.cwd == "" {
		return ""
	}
	return path.Dir(c.cwd)
}

// Mkdir creates a remote directory.
func (c *FTPClient) Mkdir(name string) error {
	return c.client.MakeDir(path.Join(c.cwd, name))
}

// Remove deletes a remote file.
func (c *FTPClient) Remove(remotePath string) error {
	return c.client.Delete(remotePath)
}

// RemoveAll deletes a remote file or directory tree.
func (c *FTPClient) RemoveAll(remotePath string) error {
	return c.client.RemoveDir(remotePath)
}

// Rename renames a remote file.
func (c *FTPClient) Rename(oldPath, newPath string) error {
	return c.client.Rename(oldPath, newPath)
}

// Download copies a remote file to a local path.
func (c *FTPClient) Download(remotePath, localPath string) error {
	resp, err := c.client.Retr(remotePath)
	if err != nil {
		return err
	}
	defer resp.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, resp)
	return err
}

// Upload copies a local file to the remote path.
func (c *FTPClient) Upload(localPath, remotePath string) error {
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	return c.client.Stor(remotePath, src)
}
