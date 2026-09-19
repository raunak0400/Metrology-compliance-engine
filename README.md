# Metrology Compliance Engine

Legal Metrology Packaged Commodities Compliance Checker — compliance checking
under the Legal Metrology Act, 2009 and the Legal Metrology (Packaged
Commodities) Rules, 2011.

This repository is a monorepo that integrates all components of the project,
previously developed across separate branches.

## Repository layout

| Directory         | Description                                                        |
|-------------------|--------------------------------------------------------------------|
| `backend/`        | Go backend service (Gin / sqlc / Asynq) — current API.             |
| `legacy-node/`    | Legacy Node.js/TypeScript service (being migrated to the Go backend).|
| `ml-service/`     | Python ML microservice (OCR / NER) consumed by the backend.        |
| `frontend/`       | Next.js web application (finalized "Niyamit" frontend).            |
| `frontend-new/`   | Alternate Next.js web application.                                  |
| `ML Pipeline_v2/` | ML label-verification pipeline (OCR + NER + preprocessing).        |
| `rule_engine/`    | Python compliance rule engine (rules, registry, tests).            |

## Getting started

Each component is self-contained with its own tooling and README. See the
README inside each directory for build and run instructions:

- `backend/README.md` — Go backend setup and migration notes.
- `legacy-node/README.md` — legacy Node service.
- `ML Pipeline_v2/README.md` — ML pipeline.
- `rule_engine/README.md` — rule engine.
