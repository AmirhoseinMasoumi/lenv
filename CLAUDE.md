# Lenv runtime notes

- On Windows, port `2222` can be blocked by excluded ephemeral ranges. Do not hardcode `2222`; always use dynamic host port selection.
- Alpine guest images used by `lenv` must enable SSH password login for root:
  - `PermitRootLogin yes`
  - `PasswordAuthentication yes`
- Default SSH wait timeout is `120` seconds to accommodate slower TCG boot.
