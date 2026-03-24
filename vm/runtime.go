package vm

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func runtimeQEMUDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".lenv", "runtime", "qemu", runtime.GOOS+"-"+runtime.GOARCH), nil
}

func managedQEMUPath() (string, error) {
	dir, err := runtimeQEMUDir()
	if err != nil {
		return "", err
	}
	name := "qemu-system-x86_64"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return findBinaryRecursive(dir, name)
}

func managedQEMUImgPath() (string, error) {
	dir, err := runtimeQEMUDir()
	if err != nil {
		return "", err
	}
	name := "qemu-img"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return findBinaryRecursive(dir, name)
}

func ensureManagedQEMU() error {
	if _, err := managedQEMUPath(); err == nil {
		return nil
	}
	url := managedQEMUURL()
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("managed runtime URL is empty")
	}
	dir, err := runtimeQEMUDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	archivePath := filepath.Join(dir, "runtime.zip")
	if err := downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download managed qemu runtime: %w", err)
	}
	checksumURL := managedQEMUChecksumURL()
	if strings.TrimSpace(checksumURL) != "" {
		if err := verifyDownloadedSHA256(archivePath, checksumURL); err != nil {
			return err
		}
	}
	if err := extractZip(archivePath, dir); err != nil {
		return fmt.Errorf("extract managed qemu runtime: %w", err)
	}
	_ = os.Remove(archivePath)
	if _, err := managedQEMUPath(); err != nil {
		return fmt.Errorf("managed runtime extracted but qemu-system-x86_64 was not found")
	}
	return nil
}

func managedQEMUURL() string {
	if v := strings.TrimSpace(os.Getenv("LENV_QEMU_RUNTIME_URL")); v != "" {
		return v
	}
	return fmt.Sprintf("https://github.com/AmirhoseinMasoumi/lenv-qemu-runtime/releases/latest/download/qemu-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
}

func managedQEMUChecksumURL() string {
	if v := strings.TrimSpace(os.Getenv("LENV_QEMU_RUNTIME_SHA256_URL")); v != "" {
		return v
	}
	return managedQEMUURL() + ".sha256"
}

func verifyDownloadedSHA256(path, checksumURL string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("fetch managed runtime checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch managed runtime checksum failed: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read managed runtime checksum: %w", err)
	}
	fields := strings.Fields(string(b))
	if len(fields) == 0 {
		return fmt.Errorf("invalid managed runtime checksum payload")
	}
	expected := strings.ToLower(strings.TrimSpace(fields[0]))
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("managed runtime checksum mismatch")
	}
	return nil
}

func extractZip(zipPath, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		outPath := filepath.Join(dst, f.Name)
		cleanRoot := filepath.Clean(dst) + string(os.PathSeparator)
		if !strings.HasPrefix(filepath.Clean(outPath), cleanRoot) {
			return fmt.Errorf("invalid zip path %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		w, err := os.Create(outPath)
		if err != nil {
			_ = rc.Close()
			return err
		}
		if _, err := io.Copy(w, rc); err != nil {
			_ = rc.Close()
			_ = w.Close()
			return err
		}
		_ = rc.Close()
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

func downloadFile(url, out string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func findBinaryRecursive(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), name) {
			found = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("%s not found", name)
	}
	return found, nil
}
