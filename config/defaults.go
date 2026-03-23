package config

import (
"fmt"
)

type Defaults struct {
Version    string
CPUs       int
Memory     string
Workspace  string
KernelPath string
PkgManager string
}

var distroDefaults = map[string]Defaults{
"alpine": {Version: "3.19", CPUs: 2, Memory: "2G", Workspace: "/workspace", KernelPath: "vmlinuz-alpine-3.19", PkgManager: "apk"},
"ubuntu": {Version: "24.04", CPUs: 2, Memory: "2G", Workspace: "/workspace", KernelPath: "vmlinuz-ubuntu-24.04", PkgManager: "apt"},
"debian": {Version: "12", CPUs: 2, Memory: "2G", Workspace: "/workspace", KernelPath: "vmlinuz-debian-12", PkgManager: "apt"},
"arch":   {Version: "latest", CPUs: 2, Memory: "2G", Workspace: "/workspace", KernelPath: "vmlinuz-arch", PkgManager: "pacman"},
}

func Resolve(lt *LenvToml) (*Config, error) {
distro := lt.Env.Distro
if distro == "" {
distro = "alpine"
}
d, ok := distroDefaults[distro]
if !ok {
return nil, fmt.Errorf("unsupported distro %q", distro)
}
cfg := &Config{
Distro:     distro,
Version:    firstNonEmpty(lt.Env.Version, d.Version),
CPUs:       firstNonZero(lt.Env.CPUs, d.CPUs),
Memory:     firstNonEmpty(lt.Env.Memory, d.Memory),
Workspace:  firstNonEmpty(lt.Mount.Workspace, d.Workspace),
Packages:   lt.Packages.Install,
KernelPath: d.KernelPath,
PkgManager: d.PkgManager,
}
return cfg, nil
}

func firstNonEmpty(v, fallback string) string {
if v == "" {
return fallback
}
return v
}

func firstNonZero(v, fallback int) int {
if v == 0 {
return fallback
}
return v
}
