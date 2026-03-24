package config

type LenvToml struct {
Env      EnvConfig     `toml:"env"`
Packages PackageConfig `toml:"packages"`
Mount    MountConfig   `toml:"mount"`
}

type EnvConfig struct {
Distro  string `toml:"distro"`
Version string `toml:"version"`
CPUs    int    `toml:"cpus"`
Memory  string `toml:"memory"`
}

type PackageConfig struct {
Install []string `toml:"install"`
}

type MountConfig struct {
Workspace string `toml:"workspace"`
}

type Config struct {
Distro     string
Version    string
CPUs       int
Memory     string
Workspace  string
Packages   []string
Accel      string
KernelPath string
PkgManager string
}

