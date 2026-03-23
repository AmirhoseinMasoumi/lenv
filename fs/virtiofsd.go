package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

func StateDir(projectDir string) string { return filepath.Join(projectDir, ".lenv") }
func PIDPath(projectDir string) string  { return filepath.Join(StateDir(projectDir), "virtiofsd.pid") }
func SocketPath(projectDir string) string {
return filepath.Join(StateDir(projectDir), "virtiofsd.sock")
}

func Start(projectDir string) error {
	if err := os.MkdirAll(StateDir(projectDir), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	bin, err := resolveVirtiofsdPath()
	if err != nil {
		return fmt.Errorf("virtiofsd not found; set LENV_VIRTIOFSD_PATH or add virtiofsd to PATH")
	}
	cmd := exec.Command(bin, "--socket-path", SocketPath(projectDir), "--shared-dir", projectDir, "--cache", "auto")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start virtiofsd: %w", err)
	}
	if err := os.WriteFile(PIDPath(projectDir), []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		return fmt.Errorf("write virtiofsd pid: %w", err)
	}
	return nil
}

func Available() bool {
	_, err := resolveVirtiofsdPath()
	return err == nil
}

func Stop(projectDir string) error {
b, err := os.ReadFile(PIDPath(projectDir))
if err != nil {
return nil
}
pid, err := strconv.Atoi(string(trimSpace(b)))
if err != nil {
return nil
}
p, err := os.FindProcess(pid)
if err != nil {
return nil
}
return p.Kill()
}

func virtiofsdBin() string {
	if runtime.GOOS == "windows" {
		return "virtiofsd.exe"
	}
	return "virtiofsd"
}

func resolveVirtiofsdPath() (string, error) {
	if explicit := os.Getenv("LENV_VIRTIOFSD_PATH"); explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("LENV_VIRTIOFSD_PATH is invalid: %w", err)
		}
		return explicit, nil
	}
	return exec.LookPath(virtiofsdBin())
}

func trimSpace(b []byte) string {
s := string(b)
for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
s = s[:len(s)-1]
}
for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
s = s[1:]
}
return s
}
