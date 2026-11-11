# Buster Security Hardening — Execution Status

Last updated: 2026-11-11

Legend: [x] implemented in source; [~] partially implemented or awaiting validation; [ ] not complete.

## Phase 1 — Filesystem and installation security
- [x] Validate install paths against the exact server-derived browser location.
- [x] Reject symbolic links in existing install path components.
- [x] Reject Windows reparse points/junctions in existing install path components.
- [x] Reject symlink/reparse manifest and embedded-asset targets.
- [x] Add path containment and Windows junction regression tests.
- [ ] Eliminate check-then-use filesystem TOCTOU races with handle-relative/no-follow operations.
- [ ] Validate Windows junction tests on a real Windows Actions runner.

## Phase 2 — Setup API security
- [x] Enforce exact Host and loopback peer.
- [x] Require trusted Origin for state-changing requests.
- [x] Restrict methods and POST content type.
- [x] Cap setup request bodies at 16 KiB.
- [x] Add no-store, nosniff, referrer, and restrictive CSP headers.
- [x] Add request-guard regression tests.
- [ ] Run Chrome and Firefox setup-flow compatibility tests.
- [ ] Run adversarial CSRF/session lifecycle tests against the built installer.

## Phase 3 — Native messaging and input
- [x] Allowlist native commands and validate key identifiers.
- [x] Enforce UTF-8, NUL, 4 KiB text, and coordinate limits.
- [x] Propagate RobotGo key-toggle and mouse-movement error results.
- [~] Convert panics from RobotGo text/click operations to errors.
- [ ] Replace RobotGo text/click methods with APIs that expose OS delivery status.
- [ ] Verify real input behavior on Windows, Linux, and macOS.

## Phase 4 — Updater
- [x] Resolve exact versioned GitHub release asset metadata.
- [x] Require SHA-256 digest and fail closed if absent.
- [x] Bound downloads and validate HTTP status/redirect destinations.
- [x] Stage updates under unpredictable names; preserve Windows .exe suffix.
- [x] Add digest, tampering, size, and HTTP status regression tests.
- [ ] Test updater replacement, rollback, interrupted downloads, and recovery on all OSes.
- [ ] Verify real release metadata contains SHA-256 digests for every supported asset.

## Phase 5 — Logging and secrets
- [x] Generate 256-bit random setup session tokens.
- [x] Compare session tokens in constant time.
- [x] Remove logging of typed text and key names.
- [x] Reject pre-existing symlinked log directories/targets and check opened file identity.
- [ ] Implement and verify restrictive Windows DACLs for log directories/files.
- [ ] Eliminate remaining platform-specific file-open race conditions.
- [ ] Add automated redaction and log-retention tests.

## Phase 6 — Build, dependency, and CI
- [x] Pin GitHub Actions to commit SHAs and restrict workflow token permissions.
- [x] Add Go helper tests/vet on Go 1.13.15 and stable.
- [x] Add setup test/compile jobs on Ubuntu, Windows, and macOS using CI-only asset stubs.
- [x] Add native-client build/test/vet jobs on Ubuntu, Windows, and macOS.
- [x] Add module checksum verification and govulncheck jobs.
- [x] Add extension pull-request build trigger and npm audit gate.
- [~] Await current GitHub Actions runs; queued runs are not passing tests.
- [ ] Make production setup asset generation reproducible and build the real installer.
- [ ] Resolve all dependency scanner findings.

## Phase 7 — Runtime QA and release
- [~] Earlier extension production build workflow passed on commit 72792ba794df74d6e2a4172a3b78cd6f10667068; latest npm audit/build run is pending.
- [ ] Complete Chrome, Edge, Firefox, and Opera runtime tests.
- [ ] Complete native messaging install/uninstall/update lifecycle tests.
- [ ] Complete OS input, updater rollback, and CSRF adversarial tests.
- [ ] Complete final independent security review.
- [ ] Publish release artifacts with reproducible builds and attestations.

## Current pull requests
- Extension hardening: https://github.com/judywallace020/buster-security-fixed/pull/1
- Native-client hardening: https://github.com/judywallace020/buster-security-fixed/pull/2

## Release gate
This repository is **not production-ready**. Keep both pull requests in draft until all unchecked release-critical items are validated. A queued CI run is not evidence of a passing build.
