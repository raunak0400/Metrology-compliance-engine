# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project documentation: root `README.md`, `SECURITY.md`, `CONTRIBUTING.md`,
  `CODE_OF_CONDUCT.md`, `LICENSE` (MIT), and GitHub issue/PR templates.

### Changed
- **Monorepo integration.** Combined the previously separate development
  branches (`backend`, `frontend`, `New`, `ml-service`) into a single `main`
  tree, each under its own directory:
  - `backend/` — Go backend service (Gin/sqlc/Asynq)
  - `legacy-node/` — legacy Node.js/TS service
  - `ml-service/` — Python ML microservice (OCR/NER)
  - `frontend/` — Next.js web app ("Niyamit")
  - `frontend-new/` — alternate Next.js web app
  - `ML Pipeline_v2/` — ML label-verification pipeline
  - `rule_engine/` — Python compliance rule engine
- Added a combined root `.gitignore`.

### Security
- Committed secrets (`.env`) and Python bytecode (`__pycache__`) from the
  source branches are intentionally excluded from `main`. A `GEMINI_API_KEY`
  previously committed on the `ml-service` branch must be treated as
  compromised and rotated — see [`SECURITY.md`](./SECURITY.md).
