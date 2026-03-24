# Changelog

All notable changes to this project are documented in this file.

## v1.0.0-ready (unreleased)

- Completed zero-dependency runtime fallback and runtime management commands.
- Added profile platform model with install/remove/trust controls.
- Added checksum and trust policy hardening for runtime and profiles.
- Added kernel profile rebuild command path and deterministic state tracking.
- Added provenance reporting commands for runtime and profiles.
- Added release notes helper command.

## v1.1.0

- Added global `--plain` output mode for script-friendly, no-ANSI command output.
- Added global `--compact` output mode for concise `key=value` output.
- Completed visual output consistency across runtime, snapshot, profile, kernel, and provenance commands.

## v0.5.2

- Added runtime status/verify/clean commands.
- Added profile remove command.
- Enforced default checksum-required policy for external profiles.

## v0.5.1

- Added managed runtime fallback and checksum verification.
- Added profile checksum verification support.
- Added kernel profile handling strict mode hooks.

## v0.5.0

- Added profile system foundation with built-ins and community install flow.
