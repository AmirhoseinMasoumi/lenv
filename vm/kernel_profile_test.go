package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AmirhoseinMasoumi/lenv/config"
)

func TestHandleKernelProfileConfigWritesPendingWhenNoBuildCmd(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		KernelPath:   "vmlinuz",
		KernelConfig: []string{"CONFIG_USB=y"},
	}
	if err := EnsureState(dir); err != nil {
		t.Fatal(err)
	}
	if err := HandleKernelProfileConfig(cfg, dir); err != nil {
		t.Fatalf("HandleKernelProfileConfig failed: %v", err)
	}
	pending := filepath.Join(StateDir(dir), "kernel", "profile-config.pending")
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("expected pending note, got: %v", err)
	}
}
