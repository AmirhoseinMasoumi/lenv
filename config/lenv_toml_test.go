package config

import (
	"crypto/sha256"
	"encoding/hex"
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

func TestValidateRejectsEmptyProfileName(t *testing.T) {
	err := Validate(&LenvToml{Env: EnvConfig{Profiles: []string{"usb", " "}}})
	if err == nil {
		t.Fatal("expected validation error for empty profile name")
	}
}

func TestApplyProfilesMergesProfileData(t *testing.T) {
	cfg := &Config{
		Distro:   "alpine",
		Packages: []string{"git"},
	}
	err := ApplyProfiles(cfg, []string{"usb", "audio"})
	if err != nil {
		t.Fatalf("ApplyProfiles returned error: %v", err)
	}
	if len(cfg.ExtraQEMUArgs) == 0 {
		t.Fatal("expected qemu args from profiles")
	}
	foundUSBUtils := false
	foundALSA := false
	for _, p := range cfg.Packages {
		if p == "usbutils" {
			foundUSBUtils = true
		}
		if p == "alsa-utils" {
			foundALSA = true
		}
	}
	if !foundUSBUtils || !foundALSA {
		t.Fatalf("expected merged packages to include usbutils and alsa-utils, got %v", cfg.Packages)
	}
}

func TestLoadProfileChecksumMismatchFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	pdir := filepath.Join(dir, ".lenv", "profiles")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(pdir, "bad.toml")
	profile := []byte("[profile]\nname=\"bad\"\nversion=\"1.0.0\"\n")
	if err := os.WriteFile(p, profile, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p+".sha256", []byte("deadbeef\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProfile("bad"); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestLoadProfileChecksumMatchSucceeds(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	pdir := filepath.Join(dir, ".lenv", "profiles")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(pdir, "ok.toml")
	profile := []byte("[profile]\nname=\"ok\"\nversion=\"1.0.0\"\n")
	if err := os.WriteFile(p, profile, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(profile)
	if err := os.WriteFile(p+".sha256", []byte(hex.EncodeToString(sum[:])+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProfile("ok"); err != nil {
		t.Fatalf("expected profile load success, got: %v", err)
	}
}
