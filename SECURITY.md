# Security Policy

We take the security of the Metrology Compliance Engine seriously. This
document explains how to report vulnerabilities and how we handle secrets.

## Supported versions

The project is under active development. Security fixes are applied to the
`main` branch. Please always test against the latest `main` before reporting.

| Branch  | Supported          |
|---------|--------------------|
| `main`  | :white_check_mark: |
| others  | :x:                |

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Instead, report privately using one of:

1. **GitHub Security Advisories** — open a draft advisory via the repository's
   **Security → Report a vulnerability** tab (preferred).
2. **Email** — contact the maintainer at **raunakkumarjha233@gmail.com** with
   the subject line `SECURITY: <short summary>`.

Please include:

- A description of the vulnerability and its impact.
- Steps to reproduce (proof-of-concept if possible).
- Affected component(s) and version/commit.
- Any suggested remediation.

### What to expect

- **Acknowledgement** within 5 business days.
- An assessment and, where valid, a remediation plan with a target timeline.
- Credit in the release notes once a fix ships, if you wish.

Please act in good faith: give us reasonable time to remediate before any
public disclosure, and avoid privacy violations, data destruction, or service
disruption while researching.

## Handling of secrets

This project follows a **secrets-never-in-git** policy:

- Real credentials (database URLs, API keys, tokens) must live only in local
  `.env` files, which are ignored by `.gitignore`. Commit only `*.env.example`
  templates with placeholder values.
- `GEMINI_API_KEY` (used by `ML Pipeline_v2/`) and all storage/database
  credentials are examples of values that must **never** be committed.

### If you find a committed secret

1. **Report it privately** using the process above — do not post the secret in
   a public issue or PR.
2. **Rotate/revoke the exposed credential immediately** at its provider
   (e.g. Google AI Studio for a Gemini key). Treat any committed secret as
   compromised, even after removal, because it remains in git history.

> **Historical note.** An earlier `ml-service` development branch committed a
> `GEMINI_API_KEY` in `ML Pipeline_v2/.env`. That value must be considered
> compromised and rotated; it is intentionally excluded from `main` via
> `.gitignore`.

## Scope

The security policy covers the code in this repository. Third-party services
(PostgreSQL, MongoDB, Redis, MinIO/S3, Google Gemini) are governed by their
respective providers' security policies.
