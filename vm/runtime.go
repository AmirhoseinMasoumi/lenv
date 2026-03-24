package vm

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

type RuntimeStatus struct {
	RootDir      string
	ManagedDir   string
	ManagedReady bool
	QEMUPath     string
	QEMUImgPath  string
}

type RuntimeManifest struct {
	Version   string `json:"version"`
	SourceURL string `json:"source_url"`
	SHA256    string `json:"sha256"`
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

func GetRuntimeStatus() (RuntimeStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return RuntimeStatus{}, err
	}
	root := filepath.Join(home, ".lenv", "runtime")
	managedDir, err := runtimeQEMUDir()
	if err != nil {
		return RuntimeStatus{}, err
	}
	st := RuntimeStatus{
		RootDir:    root,
		ManagedDir: managedDir,
	}
	if p, err := managedQEMUPath(); err == nil {
		st.ManagedReady = true
		st.QEMUPath = p
	}
	if p, err := managedQEMUImgPath(); err == nil {
		st.QEMUImgPath = p
	}
	return st, nil
}

func VerifyManagedRuntime() error {
	st, err := GetRuntimeStatus()
	if err != nil {
		return err
	}
	if !st.ManagedReady {
		return fmt.Errorf("managed runtime is not installed")
	}
	if st.QEMUPath == "" {
		return fmt.Errorf("managed runtime missing qemu-system-x86_64")
	}
	if st.QEMUImgPath == "" {
		return fmt.Errorf("managed runtime missing qemu-img")
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LENV_RUNTIME_MANIFEST_REQUIRED")), "1") {
		if _, _, err := fetchRuntimeManifestAndSig(); err != nil {
			return err
		}
	}
	return nil
}

func ClearManagedRuntime() error {
	st, err := GetRuntimeStatus()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(st.ManagedDir); err != nil {
		return err
	}
	return nil
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
	if err := verifyRuntimeManifest(archivePath); err != nil {
		return err
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

func managedQEMUManifestURL() string {
	if v := strings.TrimSpace(os.Getenv("LENV_QEMU_RUNTIME_MANIFEST_URL")); v != "" {
		return v
	}
	return managedQEMUURL() + ".manifest.json"
}

func managedQEMUManifestSigURL() string {
	if v := strings.TrimSpace(os.Getenv("LENV_QEMU_RUNTIME_MANIFEST_SIG_URL")); v != "" {
		return v
	}
	return managedQEMUManifestURL() + ".sig"
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

func verifyRuntimeManifest(archivePath string) error {
	required := strings.EqualFold(strings.TrimSpace(os.Getenv("LENV_RUNTIME_MANIFEST_REQUIRED")), "1")
	manifest, sig, err := fetchRuntimeManifestAndSig()
	if err != nil {
		if required {
			return err
		}
		return nil
	}
	if err := verifyManifestSignature(manifest, sig); err != nil {
		if required {
			return err
		}
		return nil
	}
	var m RuntimeManifest
	if err := json.Unmarshal(manifest, &m); err != nil {
		if required {
			return fmt.Errorf("parse runtime manifest: %w", err)
		}
		return nil
	}
	if strings.TrimSpace(m.SHA256) == "" {
		if required {
			return fmt.Errorf("runtime manifest missing sha256")
		}
		return nil
	}
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, strings.TrimSpace(m.SHA256)) {
		return fmt.Errorf("runtime manifest checksum mismatch")
	}
	return nil
}

func fetchRuntimeManifestAndSig() ([]byte, []byte, error) {
	manifestURL := managedQEMUManifestURL()
	sigURL := managedQEMUManifestSigURL()
	manifest, err := fetchHTTPBytes(manifestURL)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch runtime manifest: %w", err)
	}
	sig, err := fetchHTTPBytes(sigURL)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch runtime manifest signature: %w", err)
	}
	return manifest, sig, nil
}

func verifyManifestSignature(manifest, sig []byte) error {
	pub := strings.TrimSpace(os.Getenv("LENV_RUNTIME_MANIFEST_PUBKEY"))
	if pub == "" {
		return fmt.Errorf("runtime manifest public key is not set (LENV_RUNTIME_MANIFEST_PUBKEY)")
	}
	pubBytes, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return fmt.Errorf("decode runtime manifest public key: %w", err)
	}
	sigRaw := strings.TrimSpace(string(sig))
	sigBytes, err := base64.StdEncoding.DecodeString(sigRaw)
	if err != nil {
		return fmt.Errorf("decode runtime manifest signature: %w", err)
	}
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), manifest, sigBytes) {
		return fmt.Errorf("runtime manifest signature verification failed")
	}
	return nil
}

func fetchHTTPBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
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
