# Buster — security review workspace

This repository contains a reviewed working copy of [dessant/buster](https://github.com/dessant/buster), an open-source browser extension.

- Upstream: https://github.com/dessant/buster
- License: GNU GPL-3.0 (see `LICENSE`)
- Working branch: `security/hardening`
- Status: **partial manifest hardening applied; not yet production-ready or fully audited**.

## Current progress

- Imported the upstream source and preserved the GPL-3.0 license and attribution.
- Added [SECURITY_REPORT.md](SECURITY_REPORT.md) with the preliminary review, remediation notes, and remaining risks.
- Replaced broad `<all_urls>` host permissions with an explicit host allowlist across Chrome, Edge, Opera, and Firefox manifests.
- Scoped Manifest V3 web-accessible resources to the local setup origin and reCAPTCHA origins.
- Added background sender checks, reCAPTCHA-origin validation, a 10 MiB streamed audio-response limit, media-type/redirect checks, and native-client command validation.

## Important limitations

This is not a fully security-fixed release. The CSP still needs a complete destination inventory; Firefox Manifest V2 retains legacy broad web-accessible-resource declarations; custom IBM endpoints may need additional permissions; the new message/audio checks still need adversarial runtime testing; credential handling, native-client review, dependency scanning, browser builds, and runtime tests remain outstanding.

See [SECURITY_REPORT.md](SECURITY_REPORT.md) before using or distributing this branch. Preserve upstream attribution and comply with GPL-3.0 for derivative distributions.
