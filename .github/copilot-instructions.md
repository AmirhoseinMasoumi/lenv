# Copilot Instructions for `lenv`

## Build and test commands

- Build all packages: `go build ./...`
- Run all tests: `go test ./...`
- Run a single test function: `go test ./path/to/package -run '^TestName$'`
- Run tests in a single file pattern: `go test ./path/to/package -run 'TestFeature'`

Use `LENV_INTEGRATION=1` only when intentionally running integration tests that require QEMU/virtiofsd.

## High-level architecture

`lenv` is a Go CLI (Cobra) that orchestrates a per-project Linux VM rooted in the current directory.

- `cmd/` defines CLI entry points (`init`, `shell`, `run`, `install`, `snapshot`, `destroy`, `status`).
- `config/` resolves `lenv.toml` (optional) + defaults into runtime config.
- `distro/` maps distro/version to rootfs image metadata and kernel blob selection.
- `fs/` starts/stops `virtiofsd` and validates WinFsp on Windows.
- `vm/` detects acceleration (WHPX/HVF/KVM), allocates ports, and builds/starts QEMU args.
- `ssh/` waits for VM readiness and executes interactive/non-interactive commands over localhost SSH.
- `internal/ui` centralizes terminal output/progress behavior.

Runtime flow for `lenv init`/`run`: load config -> ensure distro/image assets -> start `virtiofsd` -> start QEMU -> wait for SSH -> execute shell/command with `/workspace` mounted via virtio-fs.

## Key repository conventions

- Treat `.lenv/` as the only writable area inside the user project for runtime state (disk, pid, ssh port, resolved config).
- Preserve idempotency: running `lenv init` repeatedly in the same directory should detect/reuse an existing instance.
- Keep `lenv destroy` scoped to `.lenv/` artifacts only; never remove user project files.
- Propagate command exit status for `lenv run` from the in-VM command to the host process.
- Keep cross-platform logic runtime-based (`runtime.GOOS`) and preserve accelerator detection/fallback behavior.
- Prefer actionable user-facing errors (missing QEMU/WinFsp, etc.) rather than raw low-level subprocess errors.
- In command handlers, route user output through `internal/ui` helpers for consistent UX.

## Existing project guidance to retain

The repo already contains detailed guidance in:

- `lenv-README.md`
- `lenv-copilot-instructions.md`
- `.github/lenv-copilot-instructions.md`

When changing behavior around VM lifecycle, process cleanup, distro metadata, or SSH bootstrap, align with those documents unless the user explicitly requests a behavior change.
