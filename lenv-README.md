# lenv

> Instant, per-project Linux environments. Works on Windows, macOS, and Linux.

```
$ cd C:\Projects\my-firmware
$ lenv init
  Fetching Alpine rootfs (31MB)...  done
  Building VM image...              done
  Booting Linux (QEMU/WHPX)...      done  1.8s
  Mounting /workspace...            done
  Ready. Run `lenv shell` or `lenv run <cmd>`

$ lenv run make all
  CC src/main.c
  LD firmware.elf

$ lenv shell
  root@my-firmware:/workspace# uname -a
  Linux my-firmware 6.6.0 #1 SMP x86_64 GNU/Linux
```

---

## What it is

`lenv` spins up a real Linux VM scoped to your current project directory. Your files are mounted natively inside the VM at `/workspace` — no copying, no syncing. Destroy it when you're done and your project files are untouched.

Unlike WSL2, `lenv` works on **Windows Home**, **macOS** (including Apple Silicon), and **Linux**. No Hyper-V required. No system-wide installation. No admin privileges.

---

## Why not WSL2?

| | WSL2 | lenv |
|---|---|---|
| Windows Home support | No (needs Hyper-V) | Yes |
| macOS support | No | Yes |
| Per-project isolation | No | Yes |
| Works without admin | No | Yes |
| Shareable env config | No | Yes (`lenv.toml`) |
| /mnt/c file I/O speed | Slow (9P protocol) | Fast (virtio-fs) |
| Cold start | ~3s | ~1.8s |

---

## Install

### Windows

```powershell
winget install lenv
```

Or download `lenv.exe` from [Releases](https://github.com/AmirhoseinMasoumi/lenv/releases) and add it to your PATH.

**Dependencies (one-time):**
- [QEMU for Windows](https://www.qemu.org/download/#windows)
- [WinFsp](https://winfsp.dev) — needed for virtio-fs directory sharing

`lenv` will check for these on first run and guide you if anything is missing.

### macOS

```bash
brew install lenv
```

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/AmirhoseinMasoumi/lenv/main/install.sh | sh
```

---

## Usage

### init

Initialize a Linux environment in the current directory:

```bash
lenv init
```

Choose a distro:

```bash
lenv init --distro ubuntu    # Ubuntu 24.04
lenv init --distro debian    # Debian 12
lenv init --distro arch      # Arch Linux
lenv init --distro alpine    # Alpine 3.19 (default, fastest)
```

This creates a `.lenv/` folder in your project directory containing the VM disk image and config. Add `.lenv/` to your `.gitignore`.

### shell

Drop into a bash shell inside the VM:

```bash
lenv shell
```

Your project directory is mounted at `/workspace`. Changes to files are reflected instantly on both sides — no sync delay.

### run

Run a single command inside the VM and stream output back:

```bash
lenv run make all
lenv run cargo build --release
lenv run pytest tests/
lenv run "apt install -y cmake && cmake .."
```

This is the core workflow — use Linux toolchains on your Windows/macOS project files without ever leaving your terminal.

### install

Install packages into the running VM:

```bash
lenv install cmake gcc gdb
lenv install python3 pip
```

Uses the distro's native package manager (apk, apt, pacman) automatically.

### snapshot

Save the current VM state as a reusable template:

```bash
lenv snapshot save my-embedded-toolchain
lenv snapshot restore my-embedded-toolchain
```

Snapshots are stored in `~/.lenv/snapshots/`. Share a snapshot with teammates so everyone starts with identical toolchains.

### destroy

Remove the VM and free disk space:

```bash
lenv destroy
```

Your project files are never touched. Only `.lenv/` is deleted.

### status

Check if a VM is running for the current directory:

```bash
lenv status
```

---

## lenv.toml

Commit an `lenv.toml` to your project and teammates get the exact same environment with `lenv init`:

```toml
[env]
distro  = "ubuntu"
version = "24.04"
cpus    = 2
memory  = "2G"

[packages]
install = [
  "cmake",
  "gcc",
  "gcc-arm-none-eabi",
  "gdb-multiarch",
  "libssl-dev",
]

[mount]
workspace = "/workspace"   # where your project appears inside the VM
```

This is the feature that makes `lenv` useful for teams — no more "works on my WSL" problems.

---

## How it works

`lenv` is a thin CLI that orchestrates three components:

**QEMU** runs a lightweight Linux VM with hardware acceleration:
- Windows: WHPX (Windows Hypervisor Platform — works on Home and Pro)
- macOS: HVF (Hypervisor.framework — works on Intel and Apple Silicon via Rosetta)
- Linux: KVM

**virtiofsd** shares your project directory into the VM with native filesystem semantics — no 9P slowness, no inotify breakage. The VM sees your files as if they were local.

**A custom minimal kernel** stripped of everything unnecessary (no GUI stack, no sound, no USB) boots in under 2 seconds. Each distro ships with a prebuilt kernel blob included in the `lenv` binary.

```
your terminal
    │
    ├── lenv run make
    │       │
    │       └── SSH → localhost:2222 → VM
    │                       │
    │               /workspace (virtio-fs)
    │                       │
    └───────────── C:\Projects\my-firmware  (your actual files)
```

---

## Performance

Tested on Windows 11, AMD Ryzen 7, 32GB RAM:

| Operation | Time |
|---|---|
| `lenv init` (Alpine, first time) | ~8s (download + first boot) |
| `lenv init` (cached rootfs) | ~2s |
| `lenv shell` (VM already running) | ~300ms |
| `lenv run make` (small project) | ~400ms overhead |
| File read throughput (/workspace) | ~1.2 GB/s (virtio-fs) |
| File read throughput (WSL2 /mnt/c) | ~120 MB/s (9P) |

---

## Supported platforms

| Platform | Acceleration | Status |
|---|---|---|
| Windows 10/11 (Home + Pro) | WHPX | Supported |
| macOS Intel | HVF | Supported |
| macOS Apple Silicon | HVF + Rosetta | Supported |
| Linux | KVM | Supported |
| Windows without virtualization | Software (slow) | Fallback |

---

## Compared to alternatives

**vs WSL2** — lenv works on Windows Home, macOS, and Linux. Per-project isolation. No admin required. Faster file I/O via virtio-fs. Shareable `lenv.toml`.

**vs Docker Desktop** — lenv gives you a full Linux environment, not just containers. No daemon running in the background. No image registry. Just a VM scoped to your project.

**vs Vagrant** — lenv is designed for the "I need Linux for this project right now" moment. No Vagrantfile boilerplate. `lenv init` and you're in.

**vs Quickemu** — Quickemu is for running full desktop VMs. lenv is for per-project developer toolchains with native file access.

---

## Roadmap

- [ ] `lenv exec` — run a command in an already-running VM by name from anywhere
- [ ] GPU passthrough for CUDA workloads
- [ ] VS Code Remote extension integration
- [ ] `lenv publish` — push a snapshot to a registry for team sharing
- [ ] Windows Terminal profile auto-registration

---

## Contributing

```bash
git clone https://github.com/AmirhoseinMasoumi/lenv
cd lenv
go build ./...
go test ./...
```

Requires Go 1.22+. QEMU and WinFsp are only needed at runtime, not build time.

The codebase is organized as:

```
lenv/
├── cmd/           # CLI commands (Cobra)
├── vm/            # QEMU subprocess management
├── fs/            # virtiofsd integration
├── distro/        # rootfs images and kernel blobs
├── config/        # lenv.toml parsing
└── ssh/           # SSH connection to VM
```

---

## License

MIT
