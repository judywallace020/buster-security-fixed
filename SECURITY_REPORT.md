# Preliminary Security Review — Buster

**Review status:** preliminary static review; remediation and validation are not complete.  
**Upstream reviewed:** https://github.com/dessant/buster  
**Upstream revision:** `main` tree `d65ff5c90269a92a8307b1efcb1bd14663c877ad`  
**License:** GPL-3.0; see [LICENSE](LICENSE).

## Important scope statement

This report records security concerns identified during a limited review of selected source files and manifests. It is not a penetration test, a complete audit, or evidence of an exploitable vulnerability. No security fixes are represented as completed in this commit. No dependency audit, browser runtime test, or production build was executed in this environment.

## Findings and recommended remediation

### 1. Broad extension host permissions
The Chrome and Firefox manifests request broad host access. Broad access increases the impact of a compromised content script or dependency.

**Remediation:** identify the exact hosts needed by each feature; use optional host permissions and request them only when the user enables the corresponding feature. Test all supported browsers after reducing permissions.

### 2. Broad Content Security Policy sources
The extension policy includes permissive source directives such as wildcard network/image sources and inline styles.

**Remediation:** inventory actual network destinations and assets, then replace wildcard sources with explicit origins. Remove inline style allowances where the build and UI architecture permit. Verify model downloads and supported speech providers still work.

### 3. Speech-provider credentials in extension storage
Provider API keys are stored in browser extension storage. Extension storage is not a secure secret vault and should not be treated as encrypted at rest.

**Remediation:** prefer short-lived, narrowly scoped credentials or an appropriately secured backend/native helper. Avoid shipping a static encryption key in extension code; that only obscures the secret. Document the risks of user-supplied provider keys and provide a way to clear them.

### 4. Packaged encrypted secrets are not a trustworthy secret boundary
A key derived from publicly distributed JavaScript cannot protect a long-lived secret from a user who can inspect the package.

**Remediation:** do not package production secrets. Rotate any real credentials that may have been included in a distributed build and move sensitive operations to a properly authenticated service.

### 5. Background message dispatch needs centralized validation
The background message handler dispatches operations based on message identifiers. Each operation should validate sender context, payload shape, field types, sizes, and authorization before acting.

**Remediation:** add a centralized allowlist/schema validator; reject unknown operations and unexpected sender origins; apply per-operation authorization and bounded payload sizes. Do not assume a content-script message is trusted merely because it reaches the background context.

### 6. Offscreen message lifecycle and correlation
The offscreen messaging path should be robust to concurrent requests, disconnects, late replies, and timeouts.

**Remediation:** use unique request IDs, correlate replies, enforce timeouts, handle disconnect/reconnect explicitly, and cap concurrent work and message sizes.

### 7. Audio processing resource limits
Audio download and decoding paths should not accept unbounded responses or unlimited concurrent work.

**Remediation:** enforce response-size and duration limits, validate expected media types, cancel aborted requests, and bound decoding concurrency. Prefer streaming size checks before buffering full audio.

### 8. Local setup/native helper trust boundary
The setup UI communicates with a local helper. The helper implementation is a separate project and was not audited as part of this review.

**Remediation:** review the helper separately. Require short-lived, single-use session tokens; validate Origin and extension identity; protect against CSRF; constrain filesystem paths and commands; and reject requests from untrusted origins.

### 9. Configurable speech endpoints and token exposure
User-configurable service endpoints can create a credential-exfiltration risk if a provider key is sent to an untrusted host. Tokens in WebSocket query strings may also be captured in logs.

**Remediation:** require HTTPS, use a strict official endpoint allowlist by default, make custom endpoints an explicit trust decision, and avoid putting credentials in URLs. Review logging and redaction.

### 10. Dependency and CI security coverage
A complete dependency/security scan and runtime test suite were not executed during this review.

**Remediation:** run `npm ci`, `npm audit`, secret scanning, CodeQL or equivalent static analysis, and browser-specific build/runtime tests in CI. Review and update dependencies deliberately, with lockfile changes reviewed.

## Validation performed

- Retrieved the upstream repository metadata and recursive source tree.
- Imported the text source files and GPL-3.0 license into this review workspace.
- Application changes are limited to the background message/audio validation described above and the manifest permission changes; the broader remediation plan remains incomplete.
- No build, unit test, browser integration test, dependency audit, or penetration test was run.

## Suggested verification commands

```bash
npm ci
npm audit
npm run build:chrome
npm run build:edge
npm run build:firefox
npm run build:opera
npm run build:prod:chrome
npm run build:prod:zip:chrome
```

Run these in a clean checkout and record tool versions, output, and any exceptions. A successful build alone does not establish that the security findings are remediated.

## Partial remediation applied in branch `security/hardening`

The following limited changes have now been applied to the Chrome, Edge, Opera, and Firefox manifests:

- Replaced the broad `<all_urls>` host permission with an explicit list for reCAPTCHA, the supported speech APIs, and Hugging Face model delivery.
- For Manifest V3 builds (Chrome, Edge, Opera), scoped the setup page web-accessible resource to the local setup origin and scoped reset/style resources to reCAPTCHA origins.
- Kept Firefox Manifest V2's web-accessible-resource declaration in its required legacy form; this does not provide the same per-origin scoping as the Manifest V3 declarations.

### Additional application-layer hardening applied

The background service worker now:

- Rejects malformed runtime message envelopes and messages whose sender ID is not the extension itself.
- Restricts CAPTCHA-sensitive operations (audio transcription, reset, frame-position/scale queries, and native-client control) to messages originating from the supported reCAPTCHA challenge origins.
- Validates the audio URL, language tag, final redirect URL, and response media type before transcription.
- Streams audio responses with a 10 MiB maximum instead of buffering an unbounded response.
- Validates native-client command names and argument types/ranges before forwarding messages.

### Validation of these partial changes

- Parsed all four edited manifest files as JSON successfully.
- Confirmed none of the four manifests retains the `<all_urls>` host permission.
- Confirmed the three Manifest V3 manifests have split, origin-scoped web-accessible-resource entries.
- Static source checks confirmed all native command names used by the content script are represented in the background allowlist.
- Static source checks confirmed the sender checks, audio size cap, final-origin validation, and media-type validation are present.

### Compatibility caveats and remaining work

- Custom, non-standard IBM speech endpoints may no longer be reachable unless their host is added through a deliberate permission flow. Standard Watson Speech-to-Text and IAM endpoints remain listed.
- The Content Security Policy still contains broad directives; it has not been tightened because the complete runtime destination inventory and compatibility tests have not yet been completed.
- Firefox Manifest V2 web-accessible resources remain broadly exposed under the legacy manifest format.
- The message and audio protections are partial; a complete per-operation schema review and adversarial tests are still required. Credential handling, native-client security review, dependency scanning, browser builds, and runtime tests remain outstanding.
- No full build or end-to-end test has been run. Do not treat this branch as production-ready.

## Attribution

Buster is originally developed by [dessant](https://github.com/dessant). This repository is a derivative review workspace. Preserve upstream notices and comply with GPL-3.0 when distributing modified copies.


## Native client audit — upstream `dessant/buster-client`

**Reviewed repository:** https://github.com/dessant/buster-client  
**Reviewed branch:** `master`  
**Review type:** static source inspection of the Go native-messaging client, setup server, input simulation packages, and module manifests. The upstream repository reports its last push as 2022-12-15. No executable was built or run and no issue below has been validated by dynamic testing.

### N1. Unverified updater download and execution — high priority

In `cmd/client/utils.go`, `downloadFile` performs an HTTP GET, does not check the response status, and copies the response body directly to a file opened with mode `0755`. In `cmd/client/main.go`, the downloaded setup executable is then run with `--update`.

The URL is HTTPS and points to a GitHub release, which is useful transport protection, but the updater does not verify a published checksum, signature, or trusted release manifest before execution. An HTTP error body or compromised release asset path can also be written as an executable because the status is not checked.

**Recommended fix:** require a 2xx response; reject unexpected redirects/hosts; cap download size; write to a safely created temporary file; verify a pinned publisher signature or a checksum obtained through a separately authenticated release manifest; fsync/close; then atomically replace the target. Never execute an unverified download.

### N2. Setup API accepts caller-supplied filesystem paths and extension identifiers — high priority

In `cmd/setup/main.go`, the authenticated `/api/v1/setup/install` handler reads `appDir`, `manifestDir`, `browser`, `targetEnv`, and `extension` directly from form values and passes them to `install`. The install path subsequently creates directories, writes a native-messaging manifest, restores an executable, and on Windows writes a registry value. The code shown does not constrain these paths to the expected per-user application/manifest directories or validate the extension identifier against a known format/allowlist.

The random session token is a useful mitigation, but it does not replace path validation and request-origin protections. The practical exploitability of a hostile local web page depends on browser protections, token/port discovery, and the setup page lifecycle; this has not been tested.

**Recommended fix:** derive install paths server-side from the selected supported browser and current user; never accept arbitrary absolute paths from HTTP form fields; validate browser/target combinations and extension IDs; enforce path containment after canonicalization; reject symlinks where relevant; and validate the request Origin/Host against the expected local setup origin. Add CSRF-resistant request handling and tests for traversal, symlink, malformed IDs, and unexpected origins.

### N3. Sensitive input data is written to a predictable shared-temp log — high priority

The native client logs every command and logs the complete `Data` value for key and text commands in `cmd/client/main.go`. `InitLogger` in `lib/utils/utils.go` writes to a fixed filename under the system temp directory using `O_CREATE|O_TRUNC|O_WRONLY` and mode `0666`. This can disclose text typed through the native helper and creates a predictable-file/symlink concern in shared temporary directories.

**Recommended fix:** do not log typed text, key values, tokens, or other user content; use structured redacted diagnostics; create per-user log directories with restrictive permissions (for example, 0700 directory and 0600 file on Unix); use exclusive creation or safe platform-specific temp-file APIs; and implement rotation/retention controls.

### N4. Setup session token generation ignores errors — medium priority

`newToken` in `cmd/setup/utils.go` discards the error from `uuid.NewV4()`. If secure random generation fails, the setup server may continue with an empty/invalid token rather than failing closed.

**Recommended fix:** return an error from token generation and abort setup startup if secure token creation fails. Use a cryptographically secure token with sufficient entropy and compare secrets using a constant-time comparison where applicable.

### N5. Local setup server does not visibly enforce Origin/Host policy — medium priority

The setup server listens on loopback with an ephemeral port and uses a session token, but the inspected handlers authorize requests only by comparing a form/query session value. No explicit Origin or Host validation is visible in the handlers.

**Recommended fix:** enforce exact expected local origin/host and content type, reject cross-origin requests, use strict method/path handling, and add browser-based CSRF tests. Treat the session token as defense in depth, not the sole trust boundary.

### N6. Native input command execution lacks result/error propagation — medium priority

The client dispatches keyboard and mouse actions through `robotgo`. The wrapper functions in `pkg/input` and their keyboard/mouse implementations do not return execution errors to the native-messaging response. This weakens observability and makes it difficult for the extension to distinguish successful input simulation from failed or unsupported operations.

**Recommended fix:** propagate errors through every input wrapper, validate key names and text lengths again inside the native client (not only in the extension), and return explicit failure responses. Keep mouse coordinates bounded and test multi-monitor/negative-coordinate behavior.

### N7. Old Go toolchain/dependency baseline — medium priority

The inspected Go modules declare Go 1.13 and pin older dependencies, including RobotGo and related input/system packages from 2020. This is a maintenance and supply-chain risk, but the source inspection alone does not establish a known exploitable CVE.

**Recommended fix:** inventory modules with `go list -m all` and `govulncheck`; update dependencies in controlled steps; build on supported Go toolchains for each target OS/architecture; and add automated dependency scanning to CI.

### Native-client audit validation status

- Inspected the native messaging command loop, updater download path, setup HTTP handlers, install/update routines, log initialization, keyboard/mouse wrappers, Go module manifests, and build Makefile.
- Confirmed the updater does not check HTTP status or verify a signature/checksum in the inspected code.
- Confirmed setup install parameters are accepted from request form values and passed into filesystem/manifest installation logic.
- Confirmed key/text command data is logged and the log path is a fixed filename in the system temp directory.
- Confirmed UUID generation errors are ignored.
- Not run: Go tests, `go vet`, `govulncheck`, static analyzers, cross-platform builds, native-messaging integration tests, CSRF/path-traversal tests, or updater tampering tests.

**Important:** These are review findings and remediation recommendations, not proof of remote code execution or a confirmed exploit. The native client has not been modified by this audit. The extension hardening branch does not contain a hardened/rebuilt native-client binary; native-client remediation and end-to-end compatibility testing remain separate work.


## Native-client hardening helper implementation — branch `security/native-client-hardening`

A separate branch now contains an initial helper package under `native-client-hardening/security/`:

- `DownloadVerified` requires a trusted SHA-256 digest, limits downloads to 100 MiB by default, requires HTTPS GitHub release URLs, restricts redirect hosts, streams to a temporary file, verifies the digest, and refuses to overwrite an existing destination. The caller must stage to a new path and perform the platform-specific replacement only after verification.
- `NewSessionToken` creates a 256-bit cryptographically random token and provides constant-time comparison/format validation helpers.
- `NewPrivateLogger` writes under the current user's cache directory, applies restrictive directory/file modes, and rejects non-regular log targets.
- Basic unit tests cover token validation, rejection of untrusted updater URLs, missing digest rejection, and private logger file creation.

**Integration status:** these helpers are a standalone module and are not yet wired into the upstream `cmd/client` or `cmd/setup` programs. The original updater, setup request handling, and sensitive command logging therefore remain unchanged in the native client itself. No Go test or cross-platform build has been run yet. The helper package is not a replacement native-client binary and must not be distributed as a fixed client.
