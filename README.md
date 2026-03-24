[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/AmirhoseinMasoumi/lenv/ci.yml?branch=master&label=build)](https://github.com/AmirhoseinMasoumi/lenv/actions)
[![Release](https://img.shields.io/github/v/release/AmirhoseinMasoumi/lenv)](https://github.com/AmirhoseinMasoumi/lenv/releases)
[![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/AmirhoseinMasoumi/lenv)

`lenv` is a production-ready CLI for reproducible, per-project Linux VM environments with zero-dependency runtime fallback and shareable project configuration.

> Demo GIF is generated locally from `demo.tape` (which runs `demo.ps1`) and will appear here once `demo.gif` is committed.
>
> Render command: `vhs demo.tape`

## Why teams use lenv

- **Per-project isolation by default**: each repository gets its own Linux runtime and state.
- **Zero-dependency default path**: if QEMU is not installed, `lenv` can provision a managed runtime automatically.
- **Shareable environments**: commit `lenv.toml` and optional profiles so every contributor runs the same environment.

## Quick Start

```bash
go install github.com/AmirhoseinMasoumi/lenv@latest
lenv init --distro alpine
lenv run "uname -a"
```

No manual QEMU installation is required in the default flow.

## Feature comparison

| Feature | lenv | WSL2 | Docker Desktop | Vagrant |
|---|---|---|---|---|
| Windows Home support | Yes | Limited | Yes | Yes |
| macOS support | Yes | No | Yes | Yes |
| Per-project VM isolation | Native | No | Container model | Manual |
| Zero-dependency default | Yes (managed runtime fallback) | No | No | No |
| Auto rootfs download | Yes | No | Image pull model | No |
| Shareable project config | `lenv.toml` | No | Different model | `Vagrantfile` |

## Installation

### Standard install

```bash
go install github.com/AmirhoseinMasoumi/lenv@latest
```

### Optional host-managed QEMU

If you prefer system QEMU over managed runtime fallback:

```powershell
# Windows
winget install QEMU.QEMU
```

```bash
# macOS
brew install qemu
```

```bash
# Debian/Ubuntu
sudo apt-get update && sudo apt-get install -y qemu-system-x86 qemu-utils
```

## Core commands

### Initialize environment

```bash
lenv init --distro ubuntu
```

### Run commands

```bash
lenv run "go test ./..."
lenv run --env CI=1 --env GOOS=linux "env | grep -E 'CI|GOOS'"
```

### Interactive shell

```bash
lenv shell
```

### Install guest packages

```bash
lenv install git curl build-base
```

### VM lifecycle

```bash
lenv status
lenv snapshot save baseline
lenv snapshot restore baseline
lenv destroy
```

## Project configuration (`lenv.toml`)

```toml
[env]
distro = "ubuntu"
version = "24.04"
cpus = 4
memory = "4G"
profiles = ["embedded", "usb"]

[packages]
install = ["git", "curl", "build-essential"]

[mount]
workspace = "/workspace"
```

## Profile system (platform model)

Profiles allow optional capabilities (USB, audio, GPU, embedded tooling) without bloating the default environment.

### Built-in profiles

- `minimal`
- `usb`
- `audio`
- `embedded`
- `gpu`
- `full`

### Profile operations

```bash
lenv profile list
lenv init --profile usb --profile audio
lenv profile install github.com/someone/lenv-profile-ros2
lenv profile remove ros2
lenv profile trust list
lenv profile trust add github.com/your-org/
lenv profile trust remove github.com/your-org/
lenv init --profile ros2
```

Trusted profile sources can also be managed with:

```text
~/.lenv/profiles/trusted-sources.txt
```

### Community profile format

```toml
[profile]
name = "usb"
version = "1.0.0"
author = "community-author"

[qemu]
extra_args = ["-device", "qemu-xhci"]

[kernel]
config = ["CONFIG_USB=y", "CONFIG_USB_XHCI_HCD=y"]

[packages]
install = ["usbutils", "libusb"]
```

Optional integrity file:

```text
profile.toml.sha256
```

When present, `lenv` verifies it automatically.
By default, community/local installed profiles are expected to include checksum files.

## Zero-dependency runtime mode

Runtime resolution order:

1. `LENV_QEMU_PATH`
2. `qemu-system-x86_64` from `PATH`
3. Managed runtime cache
4. Managed runtime auto-download + SHA256 verification

Managed runtime cache:

```text
~/.lenv/runtime/qemu/<os>-<arch>/
```

Related environment variables:

```bash
LENV_QEMU_PATH=/custom/path/qemu-system-x86_64
LENV_QEMU_IMG_PATH=/custom/path/qemu-img
LENV_QEMU_RUNTIME_URL=https://.../qemu-<os>-<arch>.zip
LENV_QEMU_RUNTIME_SHA256_URL=https://.../qemu-<os>-<arch>.zip.sha256
LENV_QEMU_RUNTIME_MANIFEST_URL=https://.../runtime.manifest.json
LENV_QEMU_RUNTIME_MANIFEST_SIG_URL=https://.../runtime.manifest.json.sig
LENV_RUNTIME_MANIFEST_PUBKEY=<base64-ed25519-public-key>
LENV_RUNTIME_MANIFEST_REQUIRED=1
LENV_PROFILE_VERIFY=0
LENV_PROFILE_REQUIRE_CHECKSUM=0
LENV_PROFILE_TRUST_MODE=permissive
LENV_KERNEL_REBUILD=1
LENV_KERNEL_BUILD_CMD="..."
```

## Runtime management commands

```bash
lenv runtime status
lenv runtime verify
lenv runtime clean
lenv runtime provenance
```

`lenv runtime verify` validates runtime completeness and trust policy requirements.

## Provenance

```bash
lenv provenance
```

Prints runtime source/trust policy inputs and installed profile source metadata.

## Release notes helper

```bash
lenv release-notes --version v0.6.0
```

## VS Code and shell completion

```bash
lenv vscode
lenv completion bash > /etc/bash_completion.d/lenv
lenv completion zsh > "${fpath[1]}/_lenv"
lenv completion fish > ~/.config/fish/completions/lenv.fish
lenv completion powershell > lenv.ps1
```

## CI/CD and release model

- CI runs `go build ./...`, `go test ./...`, and `go vet ./...` on push.
- Tagged releases produce binaries for Windows, Linux, and macOS targets.
- Versioned tags are the source of truth for released artifacts.

## Security and reliability notes

- Rootfs downloads are checksum-verified.
- Managed runtime artifacts are checksum-verified.
- Profile checksums are enforced when `.sha256` files are present.
- Trust policy can be tightened via runtime manifest requirements and profile source catalog.
- VM state is scoped to `.lenv/` and teardown does not remove project source files.

See also:

- `SECURITY.md`
- `SUPPORT.md`
- `UPGRADE.md`
- `CHANGELOG.md`

## Roadmap

- [x] v0.1: Init/run/destroy with rootfs auto-fetch
- [x] v0.2: Acceleration and boot optimization
- [x] v0.3: Team workflows and snapshot sharing
- [x] v0.4: DX commands, completion, CI pipeline
- [x] v0.5: Profile platform and zero-dependency runtime fallback
- [x] v0.5.2: Runtime management commands and profile lifecycle hardening
- [x] v0.5.3: Kernel rebuild command path for profile kernel config
- [x] v0.6: Signed manifest/trust policy foundations and provenance command

## Contributing

Contributions are welcome. Please open an issue or pull request with reproducible details.

```bash
go build ./...
go test ./...
go vet ./...
```

## License

MIT
