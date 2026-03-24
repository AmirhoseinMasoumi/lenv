[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/AmirhoseinMasoumi/Lenv/ci.yml?branch=master&label=build)](https://github.com/AmirhoseinMasoumi/Lenv/actions)
[![Release](https://img.shields.io/github/v/release/AmirhoseinMasoumi/Lenv)](https://github.com/AmirhoseinMasoumi/Lenv/releases)
[![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/AmirhoseinMasoumi/Lenv)

**`lenv` gives you instant per-project Linux VMs with auto rootfs download, zero global distro setup, and reproducible `lenv.toml` environments.**

<!-- DEMO GIF HERE -->

- `lenv` works per project, so you stop polluting one giant global Linux instance.
- `lenv` runs on Windows Home and non-Windows hosts; WSL2 does not.
- `lenv` bootstraps from config and image metadata automatically, not manual distro install steps.

## Quick Start

```bash
go install github.com/AmirhoseinMasoumi/Lenv@latest
lenv init --distro alpine
lenv run "uname -a"
```

## Why lenv

| Feature | lenv | WSL2 | Docker Desktop | Vagrant |
|---|---|---|---|---|
| Windows Home support | Yes | Limited / Hyper-V constraints | Yes | Yes |
| macOS support | Yes | No | Yes | Yes |
| Per-project isolation | Native | No | Container-scoped | Manual |
| No admin required | Typical flow | Often required | Often required | Varies |
| Auto rootfs download | Yes | No | Yes (images) | No |
| Shareable `lenv.toml` | Yes | No | Different model | Vagrantfile only |
| Cold start time | Fast (seconds) | Fast | Medium | Slow |
| File I/O speed | Fast local VM disk; shared fs depends on backend | Good in Linux FS, slower on mounted Windows paths | Varies by mount driver | Varies |

## Features

- **`lenv init`**: Creates and boots a Linux VM for the current project with automatic rootfs provisioning.
- **`lenv shell`**: Opens an interactive shell session directly inside your project VM.
- **`lenv run`**: Executes a command in Linux and streams output back to your host terminal.
- **`lenv install`**: Installs packages via `apk`, `apt`, or `pacman` based on distro.
- **`lenv snapshot`**: Saves and restores reusable VM disk snapshots.
- **`lenv destroy`**: Stops VM processes and removes `.lenv` state without touching project files.
- **`lenv status`**: Shows currently running `lenv` VM instances and connection info.

## Installation

### Windows

#### winget

```powershell
winget install AmirhoseinMasoumi.lenv
```

#### Manual

```powershell
# Download latest lenv.exe from Releases and place in PATH
```

#### QEMU

```powershell
winget install QEMU.QEMU
```

### macOS

#### Homebrew

```bash
brew tap amirhoseinmasoumi/tap
brew install lenv
```

#### Manual

```bash
# Download binary from Releases and move to /usr/local/bin or ~/.local/bin
```

#### QEMU

```bash
brew install qemu
```

### Linux

#### Install

```bash
curl -fsSL https://raw.githubusercontent.com/AmirhoseinMasoumi/Lenv/master/install.sh | sh
```

#### QEMU

```bash
# Debian/Ubuntu
sudo apt-get update && sudo apt-get install -y qemu-system-x86 qemu-utils
```

## Usage

### Initialize a project VM

```bash
lenv init --distro alpine
```

```text
[..] Fetching rootfs for alpine
[DONE] checksum verified
[DONE] rootfs ready
[..] Starting QEMU
[DONE] VM started
[..] Waiting for SSH
[DONE] VM ready
[OK] Ready. Run `lenv shell` or `lenv run <cmd>`
```

### Run a command inside Linux

```bash
lenv run "uname -a"
```

```text
[INFO] Running: uname -a
Linux lenv 6.x.x #1 SMP x86_64 Linux
```

### Open an interactive shell

```bash
lenv shell
```

```text
root@lenv:/workspace#
```

### Install packages

```bash
lenv install git curl build-base
```

```text
[INFO] Running package install in VM
```

### Snapshot the VM

```bash
lenv snapshot save base-toolchain
lenv snapshot restore base-toolchain
```

### Check status

```bash
lenv status
```

```text
Instance: lenv-myproject-a1b2c3d4
Project: /path/to/myproject
Running: true
PID: 12345
SSH Port: 3471
```

### Destroy the environment

```bash
lenv destroy
```

```text
[OK] Destroyed .lenv state
```

## `lenv.toml`

```toml
# Runtime environment settings
[env]
distro  = "alpine"    # alpine | ubuntu | debian | arch
version = "3.19"      # distro version hint
cpus    = 2           # VM vCPU count
memory  = "2G"        # VM memory

# Packages to install (optional workflow convention)
[packages]
install = [
  "git",
  "curl",
  "build-base",
]

# Workspace mount location inside guest
[mount]
workspace = "/workspace"
```

## How It Works

```text
+--------------------+      +------------------+      +-------------------+
| Windows/macOS/Linux| ---> | lenv CLI (Go)    | ---> | QEMU VM lifecycle |
| terminal           |      | command handlers |      | + port management |
+--------------------+      +------------------+      +---------+---------+
                                                                 |
                                                                 v
                                                      +---------------------+
                                                      | Linux guest VM      |
                                                      | cloud-init firstboot|
                                                      +---------+-----------+
                                                                |
                                                                v
                                                      +---------------------+
                                                      | /workspace mount     |
                                                      | project files host   |
                                                      +---------------------+
```

## Roadmap

- [x] **v0.1** Working `init/run/destroy`, rootfs auto-fetch, dynamic ports
- [ ] **v0.2** Better usability: hardened shell/install/status UX
- [ ] **v0.3** Shareable workflows: polished `lenv.toml` and distro matrix
- [ ] **v0.4** Performance: acceleration wiring and warm-start snapshots

## Built with

- [Go](https://go.dev/)
- [QEMU](https://www.qemu.org/)
- [cloud-init](https://cloud-init.io/)

## Contributing

Contributions are welcome. Open an issue, discuss your approach, then send a PR.

```bash
go build ./...
go test ./...
```

## License

MIT

