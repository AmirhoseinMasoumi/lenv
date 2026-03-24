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
	ChecksumAlgo string
	ChecksumURL  string
	ChecksumFile string
	KernelBlob   string
	DefaultUser  string
	PkgManager   string
	GuestSSHNote string
}

var Registry = map[string]Distro{
	"alpine": {
		Name:         "alpine",
		Version:      "3.19",
		RootFSURL:    "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/cloud/nocloud_alpine-3.19.1-x86_64-bios-cloudinit-r0.qcow2",
		ChecksumAlgo: "sha512",
		ChecksumURL:  "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/cloud/nocloud_alpine-3.19.1-x86_64-bios-cloudinit-r0.qcow2.sha512",
		KernelBlob:   "vmlinuz-alpine-3.19",
		DefaultUser:  "root",
		PkgManager:   "apk",
		GuestSSHNote: "Cloud-init should enforce OpenSSH and /etc/ssh/sshd_config entries: " +
			"PermitRootLogin yes and PasswordAuthentication yes",
	},
	"ubuntu": {
		Name:         "ubuntu",
		Version:      "24.04",
		RootFSURL:    "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
		ChecksumAlgo: "sha256",
		ChecksumURL:  "https://cloud-images.ubuntu.com/noble/current/SHA256SUMS",
		ChecksumFile: "noble-server-cloudimg-amd64.img",
		KernelBlob:   "vmlinuz-ubuntu-24.04",
		DefaultUser:  "ubuntu",
		PkgManager:   "apt",
		GuestSSHNote: "Cloud-init should configure root SSH policy for automation on first boot.",
	},
	"debian": {
		Name:         "debian",
		Version:      "12",
		RootFSURL:    "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2",
		ChecksumAlgo: "sha512",
		ChecksumURL:  "https://cloud.debian.org/images/cloud/bookworm/latest/SHA512SUMS",
		ChecksumFile: "debian-12-genericcloud-amd64.qcow2",
		KernelBlob:   "vmlinuz-debian-12",
		DefaultUser:  "root",
		PkgManager:   "apt",
		GuestSSHNote: "Cloud-init should configure SSH service and login policy for automation.",
	},
	"arch": {
		Name:         "arch",
		Version:      "latest",
		RootFSURL:    "https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2",
		ChecksumAlgo: "sha256",
		ChecksumURL:  "https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2.SHA256",
		ChecksumFile: "Arch-Linux-x86_64-cloudimg.qcow2",
		KernelBlob:   "vmlinuz-arch",
		DefaultUser:  "root",
		PkgManager:   "pacman",
		GuestSSHNote: "Cloud-init should configure SSH policy and service for first boot automation.",
	},
}
