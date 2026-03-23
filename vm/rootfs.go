package vm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/distro"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
)

func EnsureDisk(cfg *config.Config, projectDir string) error {
	if override := os.Getenv("LENV_DISK_PATH"); strings.TrimSpace(override) != "" {
		return nil
	}
	diskPath := DiskPath(projectDir)
	if _, err := os.Stat(diskPath); err == nil {
		return nil
	}
	meta, ok := distro.Registry[cfg.Distro]
	if !ok || strings.TrimSpace(meta.RootFSURL) == "" {
		return fmt.Errorf("no rootfs source configured for distro %q", cfg.Distro)
	}
	ui.Step("Fetching rootfs for " + cfg.Distro)
	if err := downloadToFile(meta.RootFSURL, diskPath); err != nil {
		return err
	}
	ui.Done("rootfs ready")
	return nil
}

func downloadToFile(url, outPath string) error {
	client := &http.Client{Timeout: 0}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create rootfs request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download rootfs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download rootfs failed: %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create disk dir: %w", err)
	}
	tmp := outPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temporary disk file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write rootfs: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("flush rootfs: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalize rootfs: %w", err)
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}
