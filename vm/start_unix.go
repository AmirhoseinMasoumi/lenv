//go:build !windows

package vm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

func startQEMU(qemu string, cfg *config.Config, projectDir string, sshPort int) error {
	args := BuildArgs(cfg, projectDir, sshPort)
	vmLog.Debug("QEMU args", "args", strings.Join(args, " "))
	cmd := exec.Command(qemu, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		vmLog.Error("failed to start QEMU", "error", err, "output", string(out))
		return fmt.Errorf("start qemu: %w (%s)", err, string(out))
	}
	if pid, err := readPID(projectDir); err == nil && pid > 0 {
		vmLog.Info("QEMU started", "pid", pid)
		_ = writeInstanceRecord(projectDir, cfg, sshPort, pid)
	}
	return nil
}
