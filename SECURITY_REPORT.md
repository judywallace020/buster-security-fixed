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
