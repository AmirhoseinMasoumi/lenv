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

func EnsureBootSnapshot(projectDir string) error {
	if strings.TrimSpace(os.Getenv("LENV_DISK_PATH")) != "" {
		return nil
	}
	disk := DiskPath(projectDir)
	base := BaseDiskPath(projectDir)
	if _, err := os.Stat(base); err == nil {
		return nil
	}
	if _, err := os.Stat(disk); err != nil {
		return nil
	}
	qemuImg, err := exec.LookPath("qemu-img")
	if err != nil {
		return fmt.Errorf("qemu-img not found for snapshot optimization: %w", err)
	}
	cmd := exec.Command(qemuImg, "create", "-f", "qcow2", "-b", disk, "-F", "qcow2", base)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create base snapshot: %w (%s)", err, string(out))
	}
	return nil
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
