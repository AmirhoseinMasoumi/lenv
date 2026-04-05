package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BaseDiskPath(projectDir string) string {
	return filepath.Join(StateDir(projectDir), "disk.base.qcow2")
}

func EnsureBootSnapshot(projectDir string) (vmWasStopped bool, err error) {
	if strings.TrimSpace(os.Getenv("LENV_DISK_PATH")) != "" {
		return false, nil
	}
	disk := DiskPath(projectDir)
	base := BaseDiskPath(projectDir)
	if _, err := os.Stat(base); err == nil {
		return false, nil
	}
	if _, err := os.Stat(disk); err != nil {
		return false, nil
	}
	
	// Stop VM first to avoid corruption
	if err := Stop(projectDir); err != nil {
		return false, fmt.Errorf("stop VM for snapshot: %w", err)
	}
	
	qemuImg, err := exec.LookPath("qemu-img")
	if err != nil {
		return false, fmt.Errorf("qemu-img not found for snapshot optimization: %w", err)
	}
	
	// Move current disk to base (this is our clean boot state)
	if err := os.Rename(disk, base); err != nil {
		return false, fmt.Errorf("rename disk to base: %w", err)
	}
	
	// Create new overlay disk on top of base
	cmd := exec.Command(qemuImg, "create", "-f", "qcow2", "-b", base, "-F", "qcow2", disk)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Try to restore on failure
		_ = os.Rename(base, disk)
		return false, fmt.Errorf("create overlay disk: %w (%s)", err, string(out))
	}
	return true, nil
}

func RestoreBootSnapshot(projectDir string) error {
	if strings.TrimSpace(os.Getenv("LENV_DISK_PATH")) != "" {
		return nil
	}
	disk := DiskPath(projectDir)
	base := BaseDiskPath(projectDir)
	if _, err := os.Stat(base); err != nil {
		return nil
	}
	qemuImg, err := exec.LookPath("qemu-img")
	if err != nil {
		return fmt.Errorf("qemu-img not found for snapshot restore: %w", err)
	}
	_ = os.Remove(disk)
	cmd := exec.Command(qemuImg, "create", "-f", "qcow2", "-b", base, "-F", "qcow2", disk)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restore base snapshot: %w (%s)", err, string(out))
	}
	return nil
}
