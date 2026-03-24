package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	tmp := t.TempDir()
	lt, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if lt == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestResolveDefaultsToAlpine(t *testing.T) {
	cfg, err := Resolve(&LenvToml{})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.Distro != "alpine" {
		t.Fatalf("expected alpine, got %s", cfg.Distro)
	}
	if cfg.Workspace != "/workspace" {
		t.Fatalf("unexpected workspace: %s", cfg.Workspace)
	}
}

func TestWriteResolved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	err := WriteResolved(path, &Config{Distro: "alpine", CPUs: 2, Memory: "2G", Workspace: "/workspace"})
	if err != nil {
		t.Fatalf("WriteResolved error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestValidateRejectsBadDistro(t *testing.T) {
	err := Validate(&LenvToml{Env: EnvConfig{Distro: "fedora"}})
	if err == nil {
		t.Fatal("expected validation error for unsupported distro")
	}
}

func TestSaveWritesValidConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "lenv.toml")
	lt := &LenvToml{
		Env: EnvConfig{
			Distro: "alpine",
			CPUs:   2,
			Memory: "2G",
		},
		Mount: MountConfig{Workspace: "/workspace"},
	}
	if err := Save(p, lt); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
}
