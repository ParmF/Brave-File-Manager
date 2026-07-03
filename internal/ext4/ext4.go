package ext4

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MountInfo describes an ext4 volume mounted via external tools.
type MountInfo struct {
	Device     string
	MountPoint string
	ReadOnly   bool
}

// IsSupported reports whether ext4 mounting is available on this platform.
func IsSupported() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath("ext4fuse")
	return err == nil
}

// Mount attempts to mount an ext4 image or device using ext4fuse (macOS).
// Requires macFUSE and ext4fuse to be installed separately.
func Mount(device, mountPoint string, readOnly bool) (*MountInfo, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("ext4 mounting is only supported on macOS in this release")
	}
	if _, err := exec.LookPath("ext4fuse"); err != nil {
		return nil, fmt.Errorf("ext4fuse not found: install via Homebrew (brew install ext4fuse)")
	}

	if err := os.MkdirAll(mountPoint, 0o755); err != nil {
		return nil, err
	}

	args := []string{device, mountPoint}
	if readOnly {
		args = append(args, "-o", "ro")
	}

	cmd := exec.Command("ext4fuse", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ext4fuse mount failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return &MountInfo{
		Device:     device,
		MountPoint: mountPoint,
		ReadOnly:   readOnly,
	}, nil
}

// Unmount unmounts an ext4 volume mounted at mountPoint.
func Unmount(mountPoint string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("ext4 unmount is only supported on macOS in this release")
	}
	cmd := exec.Command("umount", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// DefaultMountPoint returns a suggested mount path for a device.
func DefaultMountPoint(device string) string {
	base := filepath.Base(device)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join("/Volumes", base+"-ext4")
}

// Requirements returns setup instructions for ext4 support on macOS.
func Requirements() string {
	return "ext4 read/write on macOS requires macFUSE (https://osxfuse.github.io/) and ext4fuse (brew install ext4fuse). Mount volumes from the Ext4 sidebar section."
}
