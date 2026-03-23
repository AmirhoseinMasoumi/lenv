# GitHub Copilot Instructions — lenv

## Project overview

`lenv` is a CLI tool written in Go that spins up instant, per-project Linux VMs scoped to the current directory. It uses QEMU for virtualization and virtio-fs for native file sharing. The user runs `lenv init` in any project folder and gets a Linux shell with their files mounted at `/workspace` in under 2 seconds.

**Target platforms:** Windows 10/11 (WHPX acceleration), macOS Intel + Apple Silicon (HVF), Linux (KVM).  
**Language:** Go 1.22+  
**CLI framework:** Cobra  
**Key dependencies:** QEMU (subprocess), virtiofsd (subprocess), golang.org/x/crypto/ssh, github.com/BurntSushi/toml

---

## Repository structure

```
lenv/
├── main.go
├── cmd/
│   ├── root.go         # Cobra root command, global flags
│   ├── init.go         # lenv init
│   ├── shell.go        # lenv shell
│   ├── run.go          # lenv run <cmd>
│   ├── install.go      # lenv install <packages>
│   ├── snapshot.go     # lenv snapshot save/restore
│   ├── destroy.go      # lenv destroy
│   └── status.go       # lenv status
├── vm/
│   ├── manager.go      # spawn/stop/query QEMU process
│   ├── qemu.go         # build QEMU command-line args
│   ├── accel.go        # detect WHPX / HVF / KVM availability
│   └── ports.go        # find free localhost SSH port
├── fs/
│   ├── virtiofsd.go    # spawn/stop virtiofsd process
│   └── winfsp.go       # check WinFsp install on Windows
├── distro/
│   ├── registry.go     # map distro name → image URL + kernel blob
│   ├── fetch.go        # download + verify rootfs image
│   └── images.go       # embedded minimal kernel blobs (go:embed)
├── config/
│   ├── lenv_toml.go    # parse lenv.toml
│   └── defaults.go     # default values per distro
├── ssh/
│   ├── client.go       # establish SSH connection to VM
│   ├── shell.go        # interactive shell session
│   └── exec.go         # non-interactive command execution
└── internal/
    ├── ui.go           # progress bars, colored output (charmbracelet/bubbles)
    └── platform.go     # runtime.GOOS helpers
```

---

## Core concepts

### VM lifecycle

Each project directory gets exactly one VM instance. The instance name is derived from the absolute path of the directory (SHA256 prefix):

```go
// vm/manager.go
func InstanceName(projectDir string) string {
    hash := sha256.Sum256([]byte(projectDir))
    base := filepath.Base(projectDir)
    return fmt.Sprintf("lenv-%s-%x", base, hash[:4])
}
```

VM state is stored in `<project-dir>/.lenv/`:

```
.lenv/
├── disk.qcow2       # VM root disk (distro rootfs)
├── pid              # QEMU PID when running
├── ssh_port         # localhost port for SSH
├── virtiofsd.pid    # virtiofsd PID when running
└── config.toml      # resolved config (merged lenv.toml + defaults)
```

### QEMU invocation

Build the QEMU command from the vm package. Always use `qemu-system-x86_64`. Key flags:

```go
// vm/qemu.go
func BuildArgs(cfg *config.Config, projectDir string, sshPort int) []string {
    args := []string{
        "-machine", "q35,accel=" + cfg.Accel,   // accel = whpx, hvf, or kvm
        "-cpu", "host",
        "-smp", strconv.Itoa(cfg.CPUs),
        "-m", cfg.Memory,
        "-nographic",
        "-serial", "none",
        "-monitor", "none",
        // SSH port forward — how lenv connects to the VM
        "-netdev", fmt.Sprintf("user,id=net0,hostfwd=tcp:127.0.0.1:%d-:22", sshPort),
        "-device", "virtio-net-pci,netdev=net0",
        // virtio-fs share — project directory
        "-chardev", fmt.Sprintf("socket,id=char0,path=%s", virtiofsdSocket(projectDir)),
        "-device", "vhost-user-fs-pci,chardev=char0,tag=workspace",
        // memory backend required for vhost-user
        "-object", fmt.Sprintf("memory-backend-file,id=mem,size=%s,mem-path=/dev/shm,share=on", cfg.Memory),
        "-numa", "node,memdev=mem",
        // boot disk
        "-drive", fmt.Sprintf("file=%s,if=virtio,format=qcow2", diskPath(projectDir)),
        "-kernel", cfg.KernelPath,
        "-append", kernelCmdline(sshPort),
        "-daemonize",
        "-pidfile", pidPath(projectDir),
    }
    return args
}

func kernelCmdline(sshPort int) string {
    return "root=/dev/vda rw console=ttyS0 quiet " +
        "virtiofs.tag=workspace virtiofs.mount=/workspace " +
        "lenv.sshport=" + strconv.Itoa(sshPort)
}
```

### Acceleration detection

Detect the right accelerator at runtime — never hardcode:

```go
// vm/accel.go
func Detect() string {
    switch runtime.GOOS {
    case "windows":
        if whpxAvailable() { return "whpx" }
        return "tcg"   // software fallback — warn user it will be slow
    case "darwin":
        return "hvf"   // always available on macOS 10.15+
    case "linux":
        if _, err := os.Stat("/dev/kvm"); err == nil { return "kvm" }
        return "tcg"
    }
    return "tcg"
}

func whpxAvailable() bool {
    // check HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization
    // returns true if Windows Hypervisor Platform feature is enabled
}
```

### virtiofsd

virtiofsd is a separate process (written in Rust, ships as a prebuilt binary bundled with lenv releases). Spawn it before QEMU:

```go
// fs/virtiofsd.go
func Start(projectDir string) error {
    socket := virtiofsdSocket(projectDir)
    cmd := exec.Command(virtiofsdBin(),
        "--socket-path", socket,
        "--shared-dir", projectDir,
        "--cache", "auto",
    )
    // store PID in .lenv/virtiofsd.pid
}
```

On Windows, virtiofsd requires WinFsp. Check before starting:

```go
// fs/winfsp.go
func CheckInstalled() error {
    // check for winfsp.dll in %ProgramFiles(x86)%\WinFsp\bin
    // if missing, print a helpful install message and return error
}
```

### SSH connection

After QEMU boots, poll SSH until the VM is ready (max 30s), then connect:

```go
// ssh/client.go
func WaitAndConnect(port int, timeout time.Duration) (*ssh.Client, error) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        client, err := tryConnect(port)
        if err == nil { return client, nil }
        time.Sleep(200 * time.Millisecond)
    }
    return nil, fmt.Errorf("VM did not become ready within %s", timeout)
}

func tryConnect(port int) (*ssh.Client, error) {
    return ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), &ssh.ClientConfig{
        User:            "root",
        Auth:            []ssh.AuthMethod{ssh.Password("lenv")},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        Timeout:         500 * time.Millisecond,
    })
}
```

### lenv.toml parsing

```go
// config/lenv_toml.go
type LenvToml struct {
    Env      EnvConfig      `toml:"env"`
    Packages PackageConfig  `toml:"packages"`
    Mount    MountConfig    `toml:"mount"`
}

type EnvConfig struct {
    Distro  string `toml:"distro"`   // alpine, ubuntu, debian, arch
    Version string `toml:"version"`
    CPUs    int    `toml:"cpus"`
    Memory  string `toml:"memory"`   // e.g. "2G"
}

type PackageConfig struct {
    Install []string `toml:"install"`
}

type MountConfig struct {
    Workspace string `toml:"workspace"` // default: /workspace
}

func Load(projectDir string) (*LenvToml, error) {
    path := filepath.Join(projectDir, "lenv.toml")
    // if not found, return defaults — lenv.toml is optional
}
```

---

## Coding conventions

### Error handling

Always wrap errors with context. Use `fmt.Errorf("verb: %w", err)` pattern:

```go
// good
if err := vm.Start(cfg); err != nil {
    return fmt.Errorf("starting VM: %w", err)
}

// bad
if err := vm.Start(cfg); err != nil {
    return err
}
```

Surface user-facing errors with actionable messages. Never print raw Go errors to the user:

```go
// good
fmt.Fprintf(os.Stderr, "Error: QEMU not found.\nInstall it from https://www.qemu.org/download/\n")

// bad
fmt.Fprintf(os.Stderr, "exec: \"qemu-system-x86_64\": executable file not found in $PATH\n")
```

### Platform guards

Use `runtime.GOOS` for platform-specific logic. Never use build tags for small platform differences — keep it in the same file:

```go
func virtiofsdBin() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(lenvBinDir(), "virtiofsd.exe")
    default:
        return filepath.Join(lenvBinDir(), "virtiofsd")
    }
}
```

### Process management

Always store PIDs and clean them up. A crash should not leave orphaned QEMU processes:

```go
// write PID immediately after spawn
os.WriteFile(pidPath(projectDir), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)

// lenv destroy reads the PID file and kills the process
func Stop(projectDir string) error {
    pid, err := readPID(pidPath(projectDir))
    if err != nil { return nil } // already stopped
    proc, err := os.FindProcess(pid)
    if err != nil { return nil }
    return proc.Signal(syscall.SIGTERM)
}
```

### CLI output

Use the `internal/ui` package for all terminal output. Never use `fmt.Println` directly in cmd/ files:

```go
// internal/ui.go — available functions:
ui.Info("message")       // blue  →  info text
ui.Success("message")    // green →  done / ready
ui.Warn("message")       // amber →  non-fatal warning
ui.Error("message")      // red   →  fatal error (before os.Exit)
ui.Step("Booting Linux") // progress step with spinner
ui.Done("1.8s")          // completes the current spinner with timing
```

---

## Distro registry

Each distro entry defines where to fetch the rootfs and which kernel blob to use:

```go
// distro/registry.go
type Distro struct {
    Name       string
    Version    string
    ImageURL   string   // cloud image download URL
    ImageSHA256 string
    KernelBlob string   // name of embedded kernel blob in distro/blobs/
    DefaultUser string
    PkgManager string   // apk, apt, pacman
}

var Registry = map[string]Distro{
    "alpine": {
        Name:       "alpine",
        Version:    "3.19",
        ImageURL:   "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-virt-3.19.0-x86_64.iso",
        KernelBlob: "vmlinuz-alpine-3.19",
        DefaultUser: "root",
        PkgManager: "apk",
    },
    "ubuntu": {
        Name:       "ubuntu",
        Version:    "24.04",
        ImageURL:   "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
        KernelBlob: "vmlinuz-ubuntu-24.04",
        DefaultUser: "ubuntu",
        PkgManager: "apt",
    },
    "debian": {
        Name:       "debian",
        Version:    "12",
        ImageURL:   "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2",
        KernelBlob: "vmlinuz-debian-12",
        DefaultUser: "root",
        PkgManager: "apt",
    },
    "arch": {
        Name:       "arch",
        Version:    "latest",
        ImageURL:   "https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2",
        KernelBlob: "vmlinuz-arch",
        DefaultUser: "root",
        PkgManager: "pacman",
    },
}
```

---

## Key behaviours to preserve

- `lenv init` must be idempotent — running it twice in the same directory should detect the existing instance and report its status, not create a duplicate
- `lenv destroy` must never delete any files outside `.lenv/` — only the VM disk and config are removed
- `lenv run` must exit with the same exit code as the command that ran inside the VM
- If QEMU is not installed, print a clear install URL for the current platform and exit — do not silently fall back to software emulation without warning
- If `lenv.toml` is missing, that is not an error — all values fall back to defaults
- `.lenv/` directory must be created with a `.gitignore` inside it that ignores everything (`*`) so users don't accidentally commit their VM disk

---

## Testing

Unit tests live alongside source files as `_test.go`. Integration tests that require QEMU live in `tests/integration/` and are skipped in CI unless `LENV_INTEGRATION=1` is set.

When writing tests for `vm/` and `fs/` packages, use interfaces so QEMU and virtiofsd can be mocked:

```go
type VMRunner interface {
    Start(args []string) error
    Stop(pid int) error
    IsRunning(pid int) bool
}
```

---

## What Copilot should never do

- Never hardcode paths like `C:\Program Files\QEMU\qemu-system-x86_64.exe` — always search PATH first, then fall back to known install locations
- Never use `os.Exit` inside packages — only in `cmd/` after printing a user-facing error
- Never ignore errors from process spawning or file writes
- Never write to the user's project directory except inside `.lenv/`
- Never use `ssh.InsecureIgnoreHostKey()` in production SSH connections to non-localhost hosts — this is acceptable only for the localhost VM connection
