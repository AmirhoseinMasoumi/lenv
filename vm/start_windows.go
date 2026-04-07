//go:build windows

package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

const (
	CREATE_NO_WINDOW = 0x08000000
)

func startQEMU(qemu string, cfg *config.Config, projectDir string, sshPort int) error {
	args := BuildArgs(cfg, projectDir, sshPort)
	vmLog.Debug("QEMU args", "args", args)
	
	// Quote all arguments that might contain spaces for the batch file
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, " ") || strings.Contains(arg, "=") {
			quotedArgs[i] = fmt.Sprintf(`"%s"`, arg)
		} else {
			quotedArgs[i] = arg
		}
	}
	
	// Write a batch file to launch QEMU
	batchPath := filepath.Join(StateDir(projectDir), "start_qemu.bat")
	cmdLine := fmt.Sprintf(`@echo off
start /b "" "%s" %s >nul 2>&1
`, qemu, strings.Join(quotedArgs, " "))
	
	if err := os.WriteFile(batchPath, []byte(cmdLine), 0o644); err != nil {
		return fmt.Errorf("write batch file: %w", err)
	}
	
	// Run the batch file with CREATE_NO_WINDOW and detached from parent
	cmd := exec.Command(batchPath)
	cmd.Dir = projectDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW | syscall.CREATE_NEW_PROCESS_GROUP,
	}
	
	if err := cmd.Start(); err != nil {
		vmLog.Error("failed to start QEMU", "error", err)
		return fmt.Errorf("start qemu: %w", err)
	}
	
	// Wait for QEMU to start and find its PID
	time.Sleep(2 * time.Second)
	
	pid, err := findQEMUPid()
	if err != nil {
		vmLog.Warn("could not find QEMU PID", "error", err)
		pid = 0
	}
	
	vmLog.Info("QEMU started", "pid", pid)
	if err := os.WriteFile(PIDPath(projectDir), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return err
	}
	_ = writeInstanceRecord(projectDir, cfg, sshPort, pid)
	return nil
}

func findQEMUPid() (int, error) {
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq qemu-system-x86_64.exe", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			pidStr := strings.Trim(fields[1], "\" ")
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("no QEMU process found")
}

