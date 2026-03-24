package config

type LenvToml struct {
	Env      EnvConfig     `toml:"env"`
	Packages PackageConfig `toml:"packages"`
	Mount    MountConfig   `toml:"mount"`
}

type EnvConfig struct {
	Distro   string   `toml:"distro"`
	Version  string   `toml:"version"`
	CPUs     int      `toml:"cpus"`
	Memory   string   `toml:"memory"`
	Profiles []string `toml:"profiles"`
}

type PackageConfig struct {
	Install []string `toml:"install"`
}

type MountConfig struct {
	Workspace string `toml:"workspace"`
}

type Config struct {
	Distro            string
	Version           string
	CPUs              int
	Memory            string
	Workspace         string
	Packages          []string
	InstalledPackages []string
	Profiles          []string
	ExtraQEMUArgs     []string
	KernelConfig      []string
	Accel             string
	KernelPath        string
	PkgManager        string
}

type ProfileFile struct {
	Profile  ProfileMeta   `toml:"profile"`
	QEMU     ProfileQEMU   `toml:"qemu"`
	Kernel   ProfileKernel `toml:"kernel"`
	Packages ProfilePkgs   `toml:"packages"`
}

type ProfileMeta struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Author  string `toml:"author"`
}

type ProfileQEMU struct {
	ExtraArgs []string `toml:"extra_args"`
}

type ProfileKernel struct {
	Config []string `toml:"config"`
}

type ProfilePkgs struct {
	Install []string `toml:"install"`
}
