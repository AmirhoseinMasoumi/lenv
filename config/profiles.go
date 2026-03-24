package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func BuiltInProfile(name string) (*ProfileFile, bool) {
	switch name {
	case "minimal":
		return &ProfileFile{Profile: ProfileMeta{Name: "minimal", Version: "1.0.0", Author: "lenv"}}, true
	case "usb":
		return &ProfileFile{
			Profile:  ProfileMeta{Name: "usb", Version: "1.0.0", Author: "lenv"},
			QEMU:     ProfileQEMU{ExtraArgs: []string{"-device", "qemu-xhci"}},
			Kernel:   ProfileKernel{Config: []string{"CONFIG_USB=y", "CONFIG_USB_XHCI_HCD=y"}},
			Packages: ProfilePkgs{Install: []string{"usbutils", "libusb"}},
		}, true
	case "audio":
		return &ProfileFile{
			Profile:  ProfileMeta{Name: "audio", Version: "1.0.0", Author: "lenv"},
			QEMU:     ProfileQEMU{ExtraArgs: []string{"-device", "intel-hda", "-device", "hda-duplex"}},
			Packages: ProfilePkgs{Install: []string{"alsa-utils"}},
		}, true
	case "embedded":
		return &ProfileFile{
			Profile:  ProfileMeta{Name: "embedded", Version: "1.0.0", Author: "lenv"},
			QEMU:     ProfileQEMU{ExtraArgs: []string{"-serial", "mon:stdio"}},
			Packages: ProfilePkgs{Install: []string{"openocd", "minicom"}},
		}, true
	case "gpu":
		return &ProfileFile{
			Profile: ProfileMeta{Name: "gpu", Version: "1.0.0", Author: "lenv"},
			QEMU:    ProfileQEMU{ExtraArgs: []string{"-device", "virtio-gpu-pci"}},
		}, true
	case "full":
		return &ProfileFile{
			Profile:  ProfileMeta{Name: "full", Version: "1.0.0", Author: "lenv"},
			QEMU:     ProfileQEMU{ExtraArgs: []string{"-device", "qemu-xhci", "-device", "intel-hda", "-device", "hda-duplex", "-device", "virtio-gpu-pci"}},
			Packages: ProfilePkgs{Install: []string{"usbutils", "alsa-utils", "mesa-utils"}},
		}, true
	default:
		return nil, false
	}
}

func ProfileDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".lenv", "profiles"), nil
}

func LoadProfile(name string) (*ProfileFile, error) {
	if p, ok := BuiltInProfile(name); ok {
		return p, nil
	}
	dir, err := ProfileDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name+".toml")
	var pf ProfileFile
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	if _, err := toml.DecodeFile(path, &pf); err != nil {
		return nil, fmt.Errorf("parse profile %q: %w", name, err)
	}
	if strings.TrimSpace(pf.Profile.Name) == "" {
		pf.Profile.Name = name
	}
	return &pf, nil
}

func ApplyProfiles(cfg *Config, names []string) error {
	cfg.Profiles = append([]string{}, names...)
	for _, name := range names {
		pf, err := LoadProfile(name)
		if err != nil {
			return err
		}
		cfg.ExtraQEMUArgs = append(cfg.ExtraQEMUArgs, pf.QEMU.ExtraArgs...)
		cfg.Packages = mergeUnique(cfg.Packages, pf.Packages.Install)
		cfg.KernelConfig = append(cfg.KernelConfig, pf.Kernel.Config...)
	}
	return nil
}

func mergeUnique(base, add []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range base {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	for _, v := range add {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
