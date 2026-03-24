package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

func HandleKernelProfileConfig(cfg *config.Config, projectDir string) error {
	return handleKernelProfileConfig(cfg, projectDir, false)
}

func RebuildKernelProfileConfig(cfg *config.Config, projectDir string) error {
	return handleKernelProfileConfig(cfg, projectDir, true)
}

func handleKernelProfileConfig(cfg *config.Config, projectDir string, force bool) error {
	if cfg == nil || len(cfg.KernelConfig) == 0 {
		return nil
	}
	cfgHash := hashKernelConfig(cfg.KernelConfig)
	kernelDir := filepath.Join(StateDir(projectDir), "kernel")
	if err := os.MkdirAll(kernelDir, 0o755); err != nil {
		return fmt.Errorf("create kernel state dir: %w", err)
	}
	appliedPath := filepath.Join(kernelDir, "profile-config.applied")
	if !force {
		if b, err := os.ReadFile(appliedPath); err == nil && strings.Contains(string(b), "hash="+cfgHash) {
			return nil
		}
	}

	if !useDirectKernelBoot(cfg) {
		msg := "Kernel profile config requested but direct-kernel mode is disabled.\nRequested config:\n- " +
			strings.Join(cfg.KernelConfig, "\n- ") + "\n"
		if err := os.WriteFile(filepath.Join(kernelDir, "profile-config.pending"), []byte(msg), 0o644); err != nil {
			return fmt.Errorf("write kernel pending note: %w", err)
		}
		if force {
			return fmt.Errorf("kernel rebuild requires direct-kernel boot mode and LENV_KERNEL_PATH")
		}
		return nil
	}

	buildCmd := strings.TrimSpace(os.Getenv("LENV_KERNEL_BUILD_CMD"))
	if buildCmd == "" {
		if strings.EqualFold(strings.TrimSpace(os.Getenv("LENV_KERNEL_REBUILD")), "1") || force {
			return fmt.Errorf("kernel rebuild requested but LENV_KERNEL_BUILD_CMD is not set")
		}
		note := filepath.Join(kernelDir, "profile-config.pending")
		body := "Profile kernel config requested:\n- " + strings.Join(cfg.KernelConfig, "\n- ") +
			"\n\nSet LENV_KERNEL_BUILD_CMD to a kernel build command and rerun `lenv kernel rebuild`.\n"
		if err := os.WriteFile(note, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write kernel profile note: %w", err)
		}
		return nil
	}

	fragmentPath := filepath.Join(kernelDir, "profile.config")
	fragment := strings.Join(cfg.KernelConfig, "\n") + "\n"
	if err := os.WriteFile(fragmentPath, []byte(fragment), 0o644); err != nil {
		return fmt.Errorf("write kernel config fragment: %w", err)
	}

	out, err := runKernelBuildCommand(buildCmd, projectDir, fragmentPath)
	if err != nil {
		return fmt.Errorf("kernel rebuild command failed: %w (%s)", err, out)
	}
	applied := "hash=" + cfgHash + "\nfragment=" + fragmentPath + "\n"
	if err := os.WriteFile(appliedPath, []byte(applied), 0o644); err != nil {
		return fmt.Errorf("write kernel applied state: %w", err)
	}
	return nil
}

func hashKernelConfig(values []string) string {
	cp := append([]string{}, values...)
	for i := range cp {
		cp[i] = strings.TrimSpace(cp[i])
	}
	sort.Strings(cp)
	h := sha256.Sum256([]byte(strings.Join(cp, "\n")))
	return hex.EncodeToString(h[:])
}

func runKernelBuildCommand(cmdline, projectDir, configPath string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", cmdline)
	} else {
		cmd = exec.Command("sh", "-lc", cmdline)
	}
	cmd.Env = append(os.Environ(),
		"LENV_PROJECT_DIR="+projectDir,
		"LENV_KERNEL_CONFIG_FILE="+configPath,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
