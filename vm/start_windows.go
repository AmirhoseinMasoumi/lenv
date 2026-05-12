//go:build windows

package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

const (
	CREATE_NO_WINDOW  = 0x08000000
	DETACHED_PROCESS  = 0x00000008
)

func startQEMU(qemu string, cfg *config.Config, projectDir string, sshPort int) error {
	args := BuildArgs(cfg, projectDir, sshPort)
	vmLog.Debug("QEMU args", "args", args)

	// Redirect QEMU stdio to a log file so the child does not inherit lenv's
	// terminal handles. With -nographic, QEMU muxes serial to stdio; if those
	// handles are the parent's terminal the guest can stall and cloud-init
	// never finishes (sshd starts before the password is set, so auth fails).
	logPath := filepath.Join(StateDir(projectDir), "qemu.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create qemu log: %w", err)
	}
	defer logFile.Close()

	devNull, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open nul: %w", err)
	}
	defer devNull.Close()

	cmd := exec.Command(qemu, args...)
	cmd.Dir = projectDir
	cmd.Stdin = devNull
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW | DETACHED_PROCESS | syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}

	if err := cmd.Start(); err != nil {
		vmLog.Error("failed to start QEMU", "error", err)
		return fmt.Errorf("start qemu: %w", err)
	}

	pid := cmd.Process.Pid
	// Release so Go does not wait for the detached process.
	_ = cmd.Process.Release()

	vmLog.Info("QEMU started", "pid", pid)
	if err := os.WriteFile(PIDPath(projectDir), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return err
	}
	_ = writeInstanceRecord(projectDir, cfg, sshPort, pid)
	return nil
}
