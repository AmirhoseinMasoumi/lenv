# Upgrade Guide

## Upgrading to v1.0-ready

1. Upgrade binary:

```bash
go install github.com/AmirhoseinMasoumi/lenv@latest
```

2. Review trust policy defaults:

- Profiles require checksum files by default.
- GitHub profile sources are restricted by trust policy unless `LENV_PROFILE_TRUST_MODE=permissive`.
- Managed runtime manifest requirement can be enabled with `LENV_RUNTIME_MANIFEST_REQUIRED=1`.

3. Optional policy customizations:

```bash
lenv profile trust add github.com/your-org/
lenv runtime provenance
lenv provenance
```

4. Kernel profile workflows:

If profiles include kernel config requirements, configure:

```bash
LENV_KERNEL_BUILD_CMD="your-build-command"
lenv kernel rebuild
```

## Backward compatibility notes

- Existing `lenv.toml` files remain compatible.
- Existing profile files remain compatible; checksum and trust policies may require new metadata in stricter setups.
