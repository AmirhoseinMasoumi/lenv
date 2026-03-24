package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/fs"
)

func BuildArgs(cfg *config.Config, projectDir string, sshPort int) []string {
	cfg.Accel = DetectAccel()
	useVirtioFS := fs.Available() || runtime.GOOS != "windows"
	directKernelBoot := useDirectKernelBoot(cfg)
	args := []string{
		"-machine", "q35,accel=" + cfg.Accel,
		"-cpu", cpuModel(cfg.Accel),
		"-smp", strconv.Itoa(cfg.CPUs),
		"-m", cfg.Memory,
		"-nographic",
		"-serial", "none",
		"-monitor", "none",
		"-netdev", fmt.Sprintf("user,id=net0,hostfwd=tcp:127.0.0.1:%d-:22", sshPort),
		"-device", "virtio-net-pci,netdev=net0",
		"-drive", fmt.Sprintf("file=%s,if=virtio,format=qcow2", DiskPath(projectDir)),
	}
	if directKernelBoot {
		args = append(args, "-kernel", cfg.KernelPath, "-append", kernelCmdline(sshPort, cfg.Workspace, useVirtioFS))
		if initrd, ok := os.LookupEnv("LENV_INITRD_PATH"); ok && strings.TrimSpace(initrd) != "" {
			args = append(args, "-initrd", initrd)
		}
	}
	if runtime.GOOS != "windows" {
		args = append(args, "-daemonize", "-pidfile", PIDPath(projectDir))
	}
	if _, err := os.Stat(SeedISOPath(projectDir)); err == nil {
		args = append(args, "-drive", fmt.Sprintf("file=%s,if=virtio,media=cdrom,readonly=on", SeedISOPath(projectDir)))
	}
	if useVirtioFS {
		args = append(args,
			"-chardev", fmt.Sprintf("socket,id=char0,path=%s", fs.SocketPath(projectDir)),
			"-device", "vhost-user-fs-pci,chardev=char0,tag=workspace",
			"-object", fmt.Sprintf("memory-backend-file,id=mem,size=%s,mem-path=/dev/shm,share=on", cfg.Memory),
			"-numa", "node,memdev=mem",
		)
	} else if virtfsSupported() {
		hostPath := strings.ReplaceAll(projectDir, `\`, `/`)
		args = append(args,
			"-virtfs", fmt.Sprintf("local,path=%s,mount_tag=workspace,security_model=none,multidevs=remap,id=workspace", hostPath),
		)
	}
	return args
}

func kernelCmdline(sshPort int, workspace string, useVirtioFS bool) string {
	base := "root=/dev/vda rw console=ttyS0 quiet lenv.sshport=" + strconv.Itoa(sshPort)
	if useVirtioFS {
		return base + " virtiofs.tag=workspace virtiofs.mount=" + workspace
	}
	return base + " lenv.mount=9p lenv.workspace=" + workspace
}

func cpuModel(accel string) string {
	if strings.EqualFold(accel, "tcg") {
		return "max"
	}
	return "host"
}

func virtfsSupported() bool {
	v, ok := os.LookupEnv("LENV_VIRTFS_SUPPORTED")
	if !ok {
		return false
	}
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes"
}

func StateDir(projectDir string) string { return filepath.Join(projectDir, ".lenv") }
func PIDPath(projectDir string) string  { return filepath.Join(StateDir(projectDir), "pid") }
func DiskPath(projectDir string) string {
	if override := os.Getenv("LENV_DISK_PATH"); override != "" {
		return override
	}
	return filepath.Join(StateDir(projectDir), "disk.qcow2")
}
func PortPath(projectDir string) string   { return filepath.Join(StateDir(projectDir), "ssh_port") }
func ConfigPath(projectDir string) string { return filepath.Join(StateDir(projectDir), "config.toml") }
func SeedISOPath(projectDir string) string {
	return filepath.Join(StateDir(projectDir), "seed.iso")
}

