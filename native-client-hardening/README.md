# Native Client Hardening Work

This directory contains security helpers prepared for integration into the upstream
Buster native client. It is intentionally separate from the browser-extension source
and does not replace the upstream native client yet.

## Helpers

- `security.ReleaseAssetSHA256`: retrieves the exact asset URL and SHA-256 digest from GitHub's HTTPS release metadata API, validates the release tag, asset name, URL origin, and digest format, and fails closed when GitHub does not publish a digest. Uses a bounded response size and Go 1.13-compatible APIs.\n- `security.DownloadVerified`: bounded HTTPS download to a temporary file, SHA-256
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


## Upstream source import and first integration pass

The upstream `dessant/buster-client` source files and GPL-3.0 license have been copied under `upstream/` for a reviewable integration target. This is a source snapshot, not a built or signed release.

Integrated in the upstream client source:
- The updater requests the exact GitHub release metadata for the requested version, requires the expected asset's published SHA-256 digest, then downloads to a new staging path with bounded size and redirect validation before executing the verified setup binary.
- Client and setup logs now use the private per-user logger helper. Key presses, released keys, tapped keys, and typed text are no longer written to logs.
- Setup session tokens now use cryptographic random bytes, propagate token-generation failures, and compare in constant time.
- Setup HTTP requests are restricted to the exact loopback Host and a loopback peer address.
- Setup install requests validate supported browser/target identifiers, validate Chrome/Firefox extension identifier formats, and compare submitted install paths against paths recomputed for the current user.

## Not yet complete or verified

- Origin/CSRF policy still needs browser compatibility tests; the loopback Host/peer guard is not a substitute for a fully tested Origin policy.
- Setup now rejects symbolic links in existing install-directory components and rejects symlink/non-regular manifest files. Windows-specific reparse-point/junction handling still needs dedicated validation.
- Input wrappers still discard RobotGo errors; command input validation and bounds need to be added.
- Setup executable embedding/build generation and Windows/macOS builds have not yet been validated.
- CI now includes helper tests and a Linux native-client build matrix; the latest workflow result is pending.
- Do not distribute the source snapshot as a finished native-client installer until CI and platform integration tests pass.
