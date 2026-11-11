# Native Client Hardening Work

This directory contains security helpers prepared for integration into the upstream
Buster native client. It is intentionally separate from the browser-extension source
and does not replace the upstream native client yet.

## Helpers

- `security.DownloadVerified`: bounded HTTPS download to a temporary file, SHA-256
  verification against a digest supplied by trusted release metadata, and atomic
  replacement. The caller must obtain the digest from authenticated release metadata
  and must fail closed if no digest is available. Never trust a checksum downloaded
  from the same asset URL as the binary.
- `security.NewSessionToken`: cryptographically random 256-bit setup token and
  constant-time comparison helper.
- `security.NewPrivateLogger`: per-user cache directory, restrictive file modes,
  and regular-file check. Callers must also stop logging typed text, key values,
  session tokens, and other user content.

## Integration requirements

1. Replace the native client's updater download path with release metadata lookup,
   validate the exact release tag and asset name, require a valid SHA-256 digest,
   then call `DownloadVerified`. Missing digest must stop the update; do not fall
   back to unverified download.
2. Replace ignored UUID errors with `NewSessionToken`; abort setup startup on error.
   Use `SessionTokenEqual` for session comparisons.
3. Replace fixed files in the shared OS temp directory with `NewPrivateLogger`.
   Remove all logs of message.Data and other user-entered content.
4. Recompute install paths server-side from the supported browser and current user.
   Reject caller-supplied paths that do not exactly match the computed paths after
   canonicalization; validate browser/target combinations and extension identifiers.
5. Add tests for invalid/missing digests, oversized responses, redirect host changes,
   path traversal, symlink attempts, invalid extension IDs, token-generation errors,
   and log redaction.

## Status

These are helper implementations only. They have not yet been wired into the full
native-client source, compiled, cross-built, or tested against Chrome/Firefox native
messaging. Do not distribute them as a finished client or claim the native client is
security-fixed until integration and platform tests pass.

Upstream project: https://github.com/dessant/buster-client
License: GPL-3.0-only; retain the upstream LICENSE and attribution when integrating.
