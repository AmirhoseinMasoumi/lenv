# Autonomous Implementation Decisions

## Context

This document records implementation choices made autonomously for v0.2, v0.3, and v0.4.

## Decisions

1. **Snapshot optimization scope (v0.2)**  
   Implemented as qcow2 copy-on-write base snapshots (`disk.base.qcow2`) instead of VM memory snapshots, because this is cross-platform and robust in CLI workflows.

2. **Acceleration detection behavior (v0.2)**  
   Implemented runtime capability checks:
   - Windows: `whpx` only if `qemu-system-x86_64 -accel help` reports it
   - Linux: `kvm` only if `/dev/kvm` exists and qemu supports it
   - macOS: `hvf` only if qemu supports it  
   Falls back to `tcg` with warning.

3. **Benchmark implementation (v0.2)**  
   Added `lenv benchmark` using built-in lifecycle command functions to avoid shelling to external binaries and to capture precise timings.

4. **Cloud image checksums (v0.3)**  
   Used official checksum sources where available:
   - Ubuntu: `SHA256SUMS`
   - Debian: `SHA512SUMS`
   - Alpine: per-image `.sha512` files  
   Arch currently does not provide a stable checksum feed in this path; handled with explicit TODO.

5. **Feature feasibility in this environment (v0.4)**  
   For self-update and GitHub release API checks, implemented direct HTTP polling via Go; in-place replacement is implemented with platform-safe temp-file swap logic.

6. **Completions and dynamic names (v0.4)**  
   Implemented shell completion command plus dynamic completion functions for distros, snapshots, and instance names using Cobra completion APIs.

7. **Snapshot restore strategy (v0.2)**  
   Because `LENV_DISK_PATH` may point to a user-owned image, automatic boot snapshot restore/save is skipped when disk path override is set.

8. **v0.3 checksum source for Arch**  
   Arch cloud image checksum endpoint is not stable in this environment. Implemented distro entry with explicit TODO marker and checksum skipped for Arch only.

## TODOs left intentionally

- Arch checksum verification in distro metadata is marked TODO until a stable official checksum endpoint is integrated.

9. **Profile extensibility model (v0.5 groundwork)**  
   Implemented profile stacking via `--profile` and `[env].profiles`, merging extra QEMU args and package installs without changing base VM lifecycle.

10. **Community profile distribution**  
    Added `lenv profile install` supporting GitHub repo convention (`github.com/<owner>/lenv-profile-<name>` -> `raw/main/profile.toml`) and local TOML installs.
