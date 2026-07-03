package remote

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/asnasn/Brave-File-Manager/internal/fs"
)

// SFTPClient provides read/write access over SFTP.
type SFTPClient struct {
	host     string
	user     string
	password string
	port     string
	client   *sftp.Client
	ssh      *ssh.Client
	cwd      string
}

// SFTPConfig holds connection parameters.
type SFTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// ConnectSFTP opens an SFTP session.
func ConnectSFTP(cfg SFTPConfig) (*SFTPClient, error) {
	if cfg.Port == "" {
		cfg.Port = "22"
	}
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{ssh.Password(cfg.Password)},
		//nolint:gosec // user-provided host key verification deferred for MVP
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial: %w", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("sftp client: %w", err)
	}

	c := &SFTPClient{
		host:     cfg.Host,
		user:     cfg.User,
		password: cfg.Password,
		port:     cfg.Port,
		client:   client,
		ssh:      conn,
		cwd:      ".",
	}
	if wd, err := client.Getwd(); err == nil {
		c.cwd = wd
	}
	return c, nil
}

// Close closes the SFTP and SSH connections.
func (c *SFTPClient) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	if c.ssh != nil {
		return c.ssh.Close()
	}
	return nil
}

// Root returns the current remote directory.
func (c *SFTPClient) Root() string {
	return c.cwd
}

// SetRoot changes the remote working directory.
func (c *SFTPClient) SetRoot(dir string) error {
	info, err := c.client.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	c.cwd = dir
	return nil
}

// List returns entries in the current remote directory.
func (c *SFTPClient) List() ([]fs.Entry, error) {
	entries, err := c.client.ReadDir(c.cwd)
	if err != nil {
		return nil, err
	}

	result := make([]fs.Entry, 0, len(entries))
	for _, e := range entries {
		result = append(result, fs.Entry{
			Name:    e.Name(),
			Path:    path.Join(c.cwd, e.Name()),
			Size:    e.Size(),
			IsDir:   e.IsDir(),
			ModTime: e.ModTime(),
			Mode:    e.Mode(),
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
func (c *SFTPClient) Parent() string {
	if c.cwd == "/" || c.cwd == "." {
		return ""
	}
	return path.Dir(c.cwd)
}

// Mkdir creates a remote directory.
func (c *SFTPClient) Mkdir(name string) error {
	return c.client.Mkdir(path.Join(c.cwd, name))
}

// Remove deletes a remote file.
func (c *SFTPClient) Remove(remotePath string) error {
	return c.client.Remove(remotePath)
}

// RemoveAll deletes a remote file or directory tree.
func (c *SFTPClient) RemoveAll(remotePath string) error {
	return c.client.RemoveAll(remotePath)
}

// Rename renames a remote file or directory.
func (c *SFTPClient) Rename(oldPath, newPath string) error {
	return c.client.Rename(oldPath, newPath)
}

// Download copies a remote file to a local path.
func (c *SFTPClient) Download(remotePath, localPath string) error {
	src, err := c.client.Open(remotePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// Upload copies a local file to the remote path.
func (c *SFTPClient) Upload(localPath, remotePath string) error {
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := c.client.Create(remotePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
