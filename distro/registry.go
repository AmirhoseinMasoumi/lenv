package distro

// Guest image requirement for lenv-managed SSH:
// - /etc/ssh/sshd_config contains:
//   PermitRootLogin yes
//   PasswordAuthentication yes
// - sshd is enabled and running on boot.
type Distro struct {
Name        string
Version     string
ImageURL    string
ImageSHA256 string
KernelBlob  string
DefaultUser string
PkgManager  string
}

var Registry = map[string]Distro{
"alpine": {Name: "alpine", Version: "3.19", ImageURL: "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-virt-3.19.0-x86_64.iso", KernelBlob: "vmlinuz-alpine-3.19", DefaultUser: "root", PkgManager: "apk"},
"ubuntu": {Name: "ubuntu", Version: "24.04", ImageURL: "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img", KernelBlob: "vmlinuz-ubuntu-24.04", DefaultUser: "ubuntu", PkgManager: "apt"},
"debian": {Name: "debian", Version: "12", ImageURL: "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2", KernelBlob: "vmlinuz-debian-12", DefaultUser: "root", PkgManager: "apt"},
"arch":   {Name: "arch", Version: "latest", ImageURL: "https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2", KernelBlob: "vmlinuz-arch", DefaultUser: "root", PkgManager: "pacman"},
}
