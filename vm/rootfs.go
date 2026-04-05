package vm

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/distro"
	"github.com/AmirhoseinMasoumi/lenv/internal/logger"
	"github.com/AmirhoseinMasoumi/lenv/internal/progress"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/kdomanski/iso9660"
)

var log = logger.WithComponent("rootfs")

func EnsureDisk(cfg *config.Config, projectDir string) error {
	if override := strings.TrimSpace(os.Getenv("LENV_DISK_PATH")); override != "" {
		log.Debug("using disk override", "path", override)
		return prepareFirstBootSeed(projectDir)
	}

	finalDiskPath := DiskPath(projectDir)
	if _, err := os.Stat(finalDiskPath); err == nil {
		log.Debug("disk already exists", "path", finalDiskPath)
		return prepareFirstBootSeed(projectDir)
	}

	meta, ok := distro.Registry[cfg.Distro]
	if !ok || strings.TrimSpace(meta.RootFSURL) == "" {
		return fmt.Errorf("no rootfs source configured for distro %q", cfg.Distro)
	}

	downloadPath := downloadedImagePath(projectDir, meta.RootFSURL)
	ui.Step("Fetching rootfs for " + cfg.Distro)
	log.Info("downloading rootfs", "distro", cfg.Distro, "url", meta.RootFSURL)
	if err := downloadToFile(meta.RootFSURL, downloadPath); err != nil {
		log.Error("rootfs download failed", "error", err)
		return err
	}
	if err := verifyChecksum(meta, downloadPath); err != nil {
		log.Error("checksum verification failed", "error", err)
		return err
	}
	if err := ensureQcow2Disk(downloadPath, finalDiskPath); err != nil {
		log.Error("qcow2 conversion failed", "error", err)
		return err
	}
	if err := prepareFirstBootSeed(projectDir); err != nil {
		log.Error("seed preparation failed", "error", err)
		return err
	}
	log.Info("rootfs ready", "path", finalDiskPath)
	ui.Done("rootfs ready")
	return nil
}

func downloadedImagePath(projectDir, rootfsURL string) string {
	base := path.Base(rootfsURL)
	ext := filepath.Ext(base)
	if ext == "" {
		return filepath.Join(StateDir(projectDir), "rootfs.download")
	}
	return filepath.Join(StateDir(projectDir), "rootfs.download"+ext)
}

func ensureQcow2Disk(downloadPath, finalDiskPath string) error {
	if strings.EqualFold(filepath.Ext(downloadPath), ".qcow2") {
		if err := os.Rename(downloadPath, finalDiskPath); err != nil {
			return fmt.Errorf("place downloaded qcow2: %w", err)
		}
		return nil
	}

	qemuImg, err := resolveQEMUImgPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(qemuImg, "convert", "-O", "qcow2", downloadPath, finalDiskPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("convert rootfs to qcow2: %w (%s)", err, string(out))
	}
	_ = os.Remove(downloadPath)
	return nil
}

func resolveQEMUImgPath() (string, error) {
	if explicit := strings.TrimSpace(os.Getenv("LENV_QEMU_IMG_PATH")); explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("LENV_QEMU_IMG_PATH is invalid: %w", err)
		}
		return explicit, nil
	}
	if p, err := exec.LookPath("qemu-img"); err == nil {
		return p, nil
	}
	if p, err := managedQEMUImgPath(); err == nil {
		return p, nil
	}
	if err := ensureManagedQEMU(); err != nil {
		return "", fmt.Errorf("qemu-img not found in PATH and managed runtime setup failed: %w", err)
	}
	return managedQEMUImgPath()
}

func downloadToFile(url, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		if err := downloadAttempt(url, outPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("download rootfs after retries: %w", lastErr)
}

func downloadAttempt(url, outPath string) error {
	log.Debug("download attempt starting", "url", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create rootfs request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download rootfs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download rootfs failed: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	log.Debug("download started", "size", totalSize)

	tmp := outPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temporary rootfs file: %w", err)
	}

	// Create progress bar if we know the total size
	var reader io.Reader = resp.Body
	var bar *progress.Bar
	if totalSize > 0 {
		bar = progress.NewBar(totalSize, "Downloading")
		reader = progress.NewReader(resp.Body, bar)
	}

	if _, err := io.Copy(f, reader); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write downloaded rootfs: %w", err)
	}

	if bar != nil {
		bar.Finish()
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("sync downloaded rootfs: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close downloaded rootfs temp file: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalize downloaded rootfs: %w", err)
	}
	log.Debug("download complete", "path", outPath)
	return nil
}

func verifyChecksum(meta distro.Distro, filePath string) error {
	algo := strings.ToLower(strings.TrimSpace(meta.ChecksumAlgo))
	if algo == "" {
		log.Debug("no checksum configured, skipping verification")
		return nil
	}
	ui.Step("Verifying image checksum")
	log.Info("verifying checksum", "algorithm", algo, "file", filePath)
	expected, err := fetchExpectedChecksum(meta)
	if err != nil {
		return err
	}
	actual, err := fileChecksum(algo, filePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(expected, actual) {
		log.Error("checksum mismatch", "expected", expected, "actual", actual)
		return fmt.Errorf("checksum mismatch for %s", path.Base(meta.RootFSURL))
	}
	log.Debug("checksum verified", "hash", actual[:16]+"...")
	ui.Done("checksum verified")
	return nil
}

func fileChecksum(algo, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	switch algo {
	case "sha256":
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", fmt.Errorf("compute sha256: %w", err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	case "sha512":
		h := sha512.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", fmt.Errorf("compute sha512: %w", err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	default:
		return "", fmt.Errorf("unsupported checksum algorithm %q", algo)
	}
}

func fetchExpectedChecksum(meta distro.Distro) (string, error) {
	if strings.TrimSpace(meta.ChecksumURL) == "" {
		return "", fmt.Errorf("checksum URL is not configured for distro %q", meta.Name)
	}
	resp, err := http.Get(meta.ChecksumURL)
	if err != nil {
		return "", fmt.Errorf("fetch checksum metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch checksum metadata failed: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read checksum metadata: %w", err)
	}
	body := strings.TrimSpace(string(b))
	if !strings.Contains(body, "\n") {
		return strings.ToLower(strings.TrimSpace(body)), nil
	}

	target := strings.TrimSpace(meta.ChecksumFile)
	if target == "" {
		target = path.Base(meta.RootFSURL)
	}
	lineRe := regexp.MustCompile(`(?i)^([a-f0-9]{64}|[a-f0-9]{128})\s+\*?(.+)$`)
	for _, ln := range strings.Split(body, "\n") {
		m := lineRe.FindStringSubmatch(strings.TrimSpace(ln))
		if len(m) != 3 {
			continue
		}
		if strings.TrimSpace(m[2]) == target {
			return strings.ToLower(strings.TrimSpace(m[1])), nil
		}
	}
	return "", fmt.Errorf("checksum entry not found for %q", target)
}

func prepareFirstBootSeed(projectDir string) error {
	seedPath := SeedISOPath(projectDir)
	if _, err := os.Stat(seedPath); err == nil {
		return nil
	}

	userData := `#cloud-config
ssh_pwauth: true
chpasswd:
  list: |
    root:lenv
  expire: false
disable_root: false
ssh_deletekeys: false
ssh_genkeytypes: ['rsa', 'ecdsa', 'ed25519']
runcmd:
  - [ sh, -c, "mkdir -p /run/sshd" ]
  - [ sh, -c, "mkdir -p /etc/ssh/sshd_config.d" ]
  - [ sh, -c, "printf 'PermitRootLogin yes\nPasswordAuthentication yes\n' > /etc/ssh/sshd_config.d/99-lenv.conf" ]
  - [ sh, -c, "ssh-keygen -A" ]
  - [ sh, -c, "if command -v rc-update >/dev/null 2>&1; then rc-update add sshd default || true; fi" ]
  - [ sh, -c, "if command -v systemctl >/dev/null 2>&1; then systemctl restart ssh || systemctl restart sshd || true; fi" ]
`
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: lenv\n", InstanceName(projectDir))

	w, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("create iso writer: %w", err)
	}
	defer w.Cleanup()
	if err := w.AddFile(bytes.NewReader([]byte(userData)), "user-data"); err != nil {
		return fmt.Errorf("add user-data: %w", err)
	}
	if err := w.AddFile(bytes.NewReader([]byte(metaData)), "meta-data"); err != nil {
		return fmt.Errorf("add meta-data: %w", err)
	}

	f, err := os.Create(seedPath)
	if err != nil {
		return fmt.Errorf("create seed iso: %w", err)
	}
	defer f.Close()
	if err := w.WriteTo(f, "cidata"); err != nil {
		return fmt.Errorf("write seed iso: %w", err)
	}
	return nil
}
