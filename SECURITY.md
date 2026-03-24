# Security Policy

## Security posture

- Rootfs artifacts are checksum-verified.
- Managed runtime artifacts support checksum and manifest-signature policy hooks.
- Profiles support checksum verification and trusted source policy controls.

## Recommended production settings

```bash
LENV_RUNTIME_MANIFEST_REQUIRED=1
LENV_PROFILE_VERIFY=1
LENV_PROFILE_REQUIRE_CHECKSUM=1
```

Use a configured trust catalog:

```bash
lenv profile trust add github.com/your-org/
```

## Reporting a vulnerability

Please report vulnerabilities through the repository security advisory workflow when available.
