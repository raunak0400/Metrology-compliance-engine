# Metrology Compliance Engine

> Automated compliance checking for **packaged commodities** under India's
> **Legal Metrology Act, 2009** and the **Legal Metrology (Packaged
> Commodities) Rules, 2011** (incl. 2017 Amendments).

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](./backend)
[![Python](https://img.shields.io/badge/Python-3.9%2B-3776AB?logo=python&logoColor=white)](./rule_engine)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=nextdotjs&logoColor=white)](./frontend)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](./CONTRIBUTING.md)

The Metrology Compliance Engine takes a photograph of a packaged product and
determines whether its mandatory legal declarations — **MRP, net quantity,
manufacturer / packer / importer, manufacturing & expiry dates, consumer-care
details, country of origin**, etc. — comply with Indian Legal Metrology law.
Each verdict cites the exact legal clause it is based on, producing a fully
auditable compliance report.

> **Domain note.** This is a **Department of Consumer Affairs / Legal
> Metrology** system. It is **not** an FSSAI (food-safety) system. It verifies
> mandatory *declarations* on packages, not food-safety licensing.

---

## Table of contents

- [Core design principle](#core-design-principle)
- [How it works](#how-it-works)
- [Repository layout](#repository-layout)
- [Components](#components)
- [Tech stack](#tech-stack)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Testing](#testing)
- [Legal basis](#legal-basis)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

---

## Core design principle

The architecture **strictly separates *extraction* (AI / ML / OCR) from
*decision* (deterministic rules)**:

> **An ML model or LLM never decides legal compliance.**

Machine learning is used only to *read* the label (OCR, field detection,
entity recognition). The actual compliance verdict is produced by a
**deterministic, versioned, unit-tested rule engine** that maps extracted
fields to specific rules of the Legal Metrology (Packaged Commodities) Rules,
2011. This keeps every decision explainable, reproducible, and legally
defensible.

## How it works

```
IMAGE
  │  Python ML service: preprocess (skew/perspective, denoise, contrast),
  │  label/panel detection, OCR, NER / field classification
  ▼
RAW ML RESULT ─────────────────────────────► stored in MongoDB (evidence trail)
  │  deterministic normalization (Go backend)
  ▼
STRUCTURED DECLARATIONS + MEASUREMENTS ─────► stored in PostgreSQL
  │  POST /validate
  ▼
DETERMINISTIC LEGAL RULE ENGINE (Python) ──► checks + violations, each citing
  │                                            its legal clause
  ▼
MANUAL REVIEW  (low-confidence / ambiguous cases only)
  ▼
FINAL REPORT  (async → PDF/DOCX → object storage → signed URL)
```

**Inspection outcomes:** `PENDING` · `COMPLIANT` · `NON_COMPLIANT` · `MANUAL_REVIEW`
**Per-check results:** `PASS` · `FAIL` · `NOT_APPLICABLE`

The Go backend is the orchestrator: it builds the `/validate` request from the
PostgreSQL declarations + measurements, calls the Python rule engine, and
persists exactly what comes back — it never computes a verdict itself.

---

## Repository layout

This is a **monorepo** integrating every component of the project.

| Directory         | Language / stack        | Description                                                       |
|-------------------|-------------------------|-------------------------------------------------------------------|
| `backend/`        | Go (Gin, sqlc, Asynq)   | Current API and pipeline orchestrator.                            |
| `rule_engine/`    | Python (FastAPI)        | Deterministic legal rule engine — the compliance decision maker.  |
| `ml-service/`     | Python (FastAPI)        | ML/OCR microservice contract (stub in this repo).                 |
| `frontend/`       | Next.js / React / TS    | Web dashboard ("Niyamit").                                        |
| `frontend-new/`   | Next.js / React / TS    | Alternate web dashboard.                                          |
| `ML Pipeline_v2/` | Go + Python + Gemini    | Standalone label-verification pipeline (Go gateway + FastAPI).    |
| `legacy-node/`    | Node.js / TS / Express  | Original backend, since migrated to the Go `backend/`.            |

---

## Components

### `backend/` — Go API & orchestrator
The production backend. A modular monolith built with **Gin** (REST),
**pgx/v5 + sqlc** over PostgreSQL (source of truth), **MongoDB** (raw OCR/CV
output), **Redis + Asynq** (jobs, cache, rate limiting), and **MinIO/S3** for
object storage. Entry points: `cmd/api` (HTTP), `cmd/worker` (async jobs),
`cmd/seed` (seed data). See [`backend/README.md`](./backend/README.md) and the
Node → Go [`backend/MIGRATION.md`](./backend/MIGRATION.md).

### `rule_engine/` — deterministic compliance engine
A standalone **FastAPI** microservice that receives already-extracted,
structured label data as JSON and returns a structured violation report.
**No OCR, no ML.** Ruleset `2011-r6-v1.0` enforces Rules 2(l), 6, 7, 9 and 11
of the 2011 Rules with 2017 Amendment provisions; every violation cites its
exact clause. Endpoints: `POST /validate`, `GET /rules`, `GET /health`
(Swagger at `/docs`). See [`rule_engine/README.md`](./rule_engine/README.md).

### `ml-service/` — ML/OCR microservice
The extraction service the backend calls for OCR, label/panel detection and
field NER. The copy in this repo is a **deterministic stub** that returns
mock structures matching the backend contract so the full pipeline can run
locally; the production service (OpenCV, PaddleOCR, detection & NER models) is
owned by the ML team.

### `frontend/` & `frontend-new/` — web dashboards
**Next.js 16 / React 19 / Tailwind CSS 4** single-page dashboards for
uploading label images, driving inspections, and viewing compliance reports
(charts via Recharts, forms via React Hook Form + Zod, camera capture via
react-webcam). `frontend-new/` additionally bundles `jspdf` for client-side
PDF export.

### `ML Pipeline_v2/` — standalone verification pipeline
A self-contained experiment: a **Go gateway** (`:8080`) serving a single-page
dashboard and forwarding uploads to a **FastAPI engine** (`:8000`) that calls
the **Gemini API** for label extraction and analysis. Useful as a reference /
prototype; it is independent of the main `backend/` pipeline.

### `legacy-node/` — original backend
The initial **Node.js + TypeScript + Express + Prisma** implementation. It has
been fully migrated to the Go `backend/` and is retained for reference and
history. Not required to run the current system.

---

## Tech stack

| Concern              | Technology                                              |
|----------------------|--------------------------------------------------------|
| API / orchestrator   | Go 1.23+, Gin, pgx/v5, sqlc, Asynq                     |
| Rule engine / ML     | Python 3.9+, FastAPI, Pydantic v2, uvicorn             |
| Frontend             | Next.js 16, React 19, TypeScript, Tailwind CSS 4       |
| Relational DB        | PostgreSQL (source of truth)                            |
| Document store       | MongoDB (raw OCR/CV/ML evidence)                        |
| Cache / queues       | Redis + Asynq                                           |
| Object storage       | MinIO / S3-compatible                                   |
| Containerization     | Docker / Docker Compose                                 |

---

## Getting started

### Prerequisites
- **Go** 1.23+ (backend) · **Python** 3.9+ (rule engine, ML service) ·
  **Node.js** 18+ (frontends)
- **PostgreSQL**, **MongoDB**, **Redis**, and an **S3-compatible** store
  (e.g. MinIO) for the full backend — or use `backend/docker-compose.yml`.

Each component is self-contained. Start the two Python services and the Go
backend, then a frontend.

### 1. Rule engine
```bash
cd rule_engine
pip install -r requirements.txt
uvicorn main:app --reload --port 8000     # Swagger UI at http://localhost:8000/docs
```

### 2. ML service (stub)
```bash
cd ml-service
pip install -r requirements.txt
uvicorn app:app --reload --port 8001
```

### 3. Backend (Go)
```bash
cd backend
cp .env.example .env                       # then fill in DB / service URLs
# point RULE_ENGINE_SERVICE_URL and the ML service URL at the services above
go run ./cmd/api
# or bring up the full stack:
docker compose up --build
```

### 4. Frontend
```bash
cd frontend        # or: cd frontend-new
npm install
npm run dev                                # http://localhost:3000
```

> `ML Pipeline_v2/` runs independently — see
> [`ML Pipeline_v2/README.md`](./ML%20Pipeline_v2/README.md).

## Configuration

Configuration is per component via environment variables. **Never commit real
secrets** — copy the provided `*.env.example` files and fill in local values.
Notable variables:

- `backend/` — database URLs (PostgreSQL, MongoDB, Redis), object-storage
  credentials, `RULE_ENGINE_SERVICE_URL`, and the ML service URL. See
  `backend/.env.example`.
- `ML Pipeline_v2/` — `GEMINI_API_KEY` (Google AI Studio).

See [`SECURITY.md`](./SECURITY.md) for secret-handling policy.

## Testing

```bash
# Rule engine
cd rule_engine && pytest tests/ -v

# Go backend
cd backend && go test ./...

# Legacy Node backend
cd legacy-node && npm test
```

---

## Legal basis

Compliance checks implement provisions of:

- **The Legal Metrology Act, 2009**
- **The Legal Metrology (Packaged Commodities) Rules, 2011** — Rules 2(l), 6,
  7, 9 and 11 — including the **2017 Amendments**.

The active ruleset version is surfaced by the rule engine (`GET /rules`) and is
currently `2011-r6-v1.0`. Every reported violation references the specific
clause it is based on, so reports are auditable end to end.

## Security

Please review the [Security Policy](./SECURITY.md) before deploying. If you
discover a vulnerability, **do not open a public issue** — follow the private
disclosure process described there. If you find a committed secret (e.g. an API
key) in history, report it and rotate the key immediately.

## Contributing

Contributions are welcome! Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md)
for the development workflow, coding standards, and commit/PR conventions, and
abide by our [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

Released under the [MIT License](./LICENSE). © 2026 raunak0400.
