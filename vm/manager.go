package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

"github.com/AmirhoseinMasoumi/lenv/config"
)

func EnsureState(projectDir string) error {
if err := os.MkdirAll(StateDir(projectDir), 0o755); err != nil {
return fmt.Errorf("create state dir: %w", err)
}
gitignore := filepathJoin(StateDir(projectDir), ".gitignore")
if _, err := os.Stat(gitignore); err != nil {
if werr := os.WriteFile(gitignore, []byte("*\n"), 0o644); werr != nil {
return fmt.Errorf("write .lenv/.gitignore: %w", werr)
}
}
return nil
}

func EnsurePort(projectDir string) (int, error) {
if b, err := os.ReadFile(PortPath(projectDir)); err == nil {
if p, err := strconv.Atoi(trimSpace(string(b))); err == nil && p > 0 {
return p, nil
}
}
p, err := FindFreePort()
if err != nil {
return 0, fmt.Errorf("find free port: %w", err)
}
if err := os.WriteFile(PortPath(projectDir), []byte(strconv.Itoa(p)), 0o644); err != nil {
return 0, fmt.Errorf("write ssh port: %w", err)
}
return p, nil
}

func Start(cfg *config.Config, projectDir string, sshPort int) error {
	qemu, err := resolveQEMUPath()
	if err != nil {
		return fmt.Errorf("QEMU not found. Install from https://www.qemu.org/download/ or set LENV_QEMU_PATH")
	}
	if useDirectKernelBoot() {
		if _, err := os.Stat(cfg.KernelPath); err != nil {
			return fmt.Errorf("kernel image not found at %q; set LENV_KERNEL_PATH to a valid Linux kernel image", cfg.KernelPath)
		}
	}
	if _, err := os.Stat(DiskPath(projectDir)); err != nil {
		return fmt.Errorf("disk image not found at %q; set LENV_DISK_PATH to a bootable qcow2 image", DiskPath(projectDir))
	}
	args := BuildArgs(cfg, projectDir, sshPort)
	cmd := exec.Command(qemu, args...)
	if runtime.GOOS == "windows" {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start qemu: %w", err)
		}
		pid := cmd.Process.Pid
		if err := os.WriteFile(PIDPath(projectDir), []byte(strconv.Itoa(pid)), 0o644); err != nil {
			return fmt.Errorf("write qemu pid: %w", err)
		}
		_ = writeInstanceRecord(projectDir, cfg, sshPort, pid)
		return nil
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("start qemu: %w (%s)", err, string(out))
	}
	if pid, err := readPID(projectDir); err == nil && pid > 0 {
		_ = writeInstanceRecord(projectDir, cfg, sshPort, pid)
	}
	return nil
}

func Stop(projectDir string) error {
b, err := os.ReadFile(PIDPath(projectDir))
if err != nil {
_ = removeInstanceRecord(projectDir)
return nil
}
pid, err := strconv.Atoi(trimSpace(string(b)))
if err != nil {
_ = removeInstanceRecord(projectDir)
return nil
}
proc, err := os.FindProcess(pid)
if err != nil {
_ = removeInstanceRecord(projectDir)
return nil
}
 if err := proc.Kill(); err != nil {
 return err
 }
 _ = removeInstanceRecord(projectDir)
 return nil
}

type Status struct {
Instance string
ProjectDir string
Accel    string
Distro   string
PID      int
SSHPort  int
Running  bool
}

func GetStatus(projectDir string) Status {
s := Status{Instance: InstanceName(projectDir), ProjectDir: projectDir}
if b, err := os.ReadFile(PIDPath(projectDir)); err == nil {
if p, e := strconv.Atoi(trimSpace(string(b))); e == nil {
s.PID = p
s.Running = processRunning(p)
}
}
if b, err := os.ReadFile(PortPath(projectDir)); err == nil {
if p, e := strconv.Atoi(trimSpace(string(b))); e == nil {
s.SSHPort = p
}
}
return s
}

func readPID(projectDir string) (int, error) {
	b, err := os.ReadFile(PIDPath(projectDir))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(trimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid)).CombinedOutput()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), strconv.Itoa(pid))
	}
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid=").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func trimSpace(s string) string {
for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') { s = s[1:] }
for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') { s = s[:len(s)-1] }
return s
}

func filepathJoin(a, b string) string {
	return a + string(os.PathSeparator) + b
}

func resolveQEMUPath() (string, error) {
	if explicit := os.Getenv("LENV_QEMU_PATH"); explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("LENV_QEMU_PATH is invalid: %w", err)
		}
		return explicit, nil
	}
	return exec.LookPath("qemu-system-x86_64")
}

func ResolveKernelPath(cfg *config.Config, projectDir string) {
	if _, ok := os.LookupEnv("LENV_KERNEL_PATH"); ok {
		p := os.Getenv("LENV_KERNEL_PATH")
		cfg.KernelPath = p
		return
	}
	if filepath.IsAbs(cfg.KernelPath) {
		return
	}
	cfg.KernelPath = filepath.Join(StateDir(projectDir), cfg.KernelPath)
}

func useDirectKernelBoot() bool {
	kernel, hasKernel := os.LookupEnv("LENV_KERNEL_PATH")
	if hasKernel && strings.TrimSpace(kernel) == "" {
		return false
	}
	return true
}
