package distro

// Guest image requirement for lenv-managed SSH:
// - /etc/ssh/sshd_config contains:
//   PermitRootLogin yes
//   PasswordAuthentication yes
// - sshd is enabled and running on boot.
type Distro struct {
	Name         string
	Version      string
	RootFSURL    string
	ImageSHA256  string
	KernelBlob   string
	DefaultUser  string
	PkgManager   string
	GuestSSHNote string
}

var Registry = map[string]Distro{
	"alpine": {
		Name:        "alpine",
		Version:     "3.19",
		RootFSURL:   "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/cloud/nocloud_alpine-3.19.1-x86_64-bios-cloudinit-r0.qcow2",
		KernelBlob:  "vmlinuz-alpine-3.19",
		DefaultUser: "root",
		PkgManager:  "apk",
		GuestSSHNote: "Requires OpenSSH enabled plus /etc/ssh/sshd_config entries: " +
			"PermitRootLogin yes and PasswordAuthentication yes",
	},
	"ubuntu": {
		Name:         "ubuntu",
		Version:      "24.04",
		RootFSURL:    "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
		KernelBlob:   "vmlinuz-ubuntu-24.04",
		DefaultUser:  "ubuntu",
		PkgManager:   "apt",
		GuestSSHNote: "Root password SSH login is disabled by default; configure credentials before first run.",
	},
	"debian": {
		Name:         "debian",
		Version:      "12",
		RootFSURL:    "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2",
		KernelBlob:   "vmlinuz-debian-12",
		DefaultUser:  "root",
		PkgManager:   "apt",
		GuestSSHNote: "Requires SSH service enabled with explicit login policy for automation.",
	},
	"arch": {
		Name:         "arch",
		Version:      "latest",
		RootFSURL:    "https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2",
		KernelBlob:   "vmlinuz-arch",
		DefaultUser:  "root",
		PkgManager:   "pacman",
		GuestSSHNote: "Requires SSH service enabled with password auth for lenv SSH bootstrap.",
	},
}
