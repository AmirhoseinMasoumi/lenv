package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

func HandleKernelProfileConfig(cfg *config.Config, projectDir string) error {
	if cfg == nil || len(cfg.KernelConfig) == 0 {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LENV_KERNEL_REBUILD")), "1") {
		return fmt.Errorf("kernel rebuild requested by profile, but automated kernel rebuild pipeline is not implemented yet")
	}
	note := filepath.Join(StateDir(projectDir), "kernel-profile.todo")
	body := "Profile kernel config requested:\n- " + strings.Join(cfg.KernelConfig, "\n- ") + "\n\nSet LENV_KERNEL_REBUILD=1 to enforce hard failure until rebuild automation lands.\n"
	if err := os.WriteFile(note, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write kernel profile note: %w", err)
	}
	return nil
}
