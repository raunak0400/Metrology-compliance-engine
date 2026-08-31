# Legal Metrology Packaged Commodities Compliance Checker — Backend

Backend skeleton for a system that checks compliance of **packaged commodities**
under the **Legal Metrology Act, 2009** and the **Legal Metrology (Packaged
Commodities) Rules, 2011**.

> **Domain note:** This is a **Department of Consumer Affairs / Legal Metrology**
> system. It is **not** an FSSAI system. It verifies mandatory declarations on
> packages (MRP, net quantity, manufacturer/packer/importer, dates, consumer care,
> country of origin, etc.), not food-safety licensing.

This repository is the **Node.js backend (modular monolith)**. Computer vision /
OCR / model inference lives in a **separate Python ML service** (see `ml-service/`
placeholder and the `ml-service` branch).

---

## Core design principle

The architecture strictly separates *extraction* (AI/ML) from *decision*
(deterministic rules). **An ML model or LLM never decides legal compliance.**

```
IMAGE
  ↓  (Python ML service: preprocess, skew/perspective, denoise, contrast,
  ↓   label/panel detection, OCR, NER/classification)
RAW ML RESULT            → stored in MongoDB (evidence trail)
  ↓  NORMALIZATION        (deterministic, Node)
STRUCTURED DECLARATIONS + MEASUREMENTS   → stored in PostgreSQL
  ↓  DETERMINISTIC LEGAL RULE ENGINE     (pure, testable, Node)
COMPLIANCE CHECKS
  ↓
VIOLATIONS + EVIDENCE
  ↓
MANUAL REVIEW            (low-confidence / ambiguous cases only)
  ↓
FINAL REPORT             (async, PDF/DOCX → object storage → signed URL)
```

Outcomes: `COMPLIANT`, `NON_COMPLIANT`, `MANUAL_REVIEW`.

---

## Tech stack

| Concern            | Choice                                             |
| ------------------ | -------------------------------------------------- |
| Runtime / language | Node.js + TypeScript                               |
| HTTP framework     | Express.js (REST)                                  |
| Relational DB      | PostgreSQL via Prisma (source of truth)            |
| Document store     | MongoDB via Mongoose (raw OCR/CV/ML output)        |
| Cache / queues     | Redis + BullMQ (jobs, cache, locks, rate limiting) |
| Object storage     | S3-compatible — MinIO (dev), AWS S3 (prod)         |
| ML service         | Separate Python HTTP service (OpenCV/PaddleOCR/…)  |
| Auth               | JWT access + rotating refresh, Argon2 hashing      |
| Validation         | Zod                                                |
| Logging            | Pino (structured, request-scoped)                  |

## Responsibilities by store

- **PostgreSQL (authoritative):** users, RBAC, products, inspections, normalized
  declarations, legal rules + requirements, compliance checks, violations,
  evidence references, ML job state, reports, datasets, annotations, audit logs.
- **MongoDB (flexible):** raw OCR text/blocks/bboxes/confidence, detected regions,
  preprocessing metadata, model predictions, field-extraction results. Never the
  authoritative compliance result.
- **Object storage:** the actual binary files (original/label/cropped images,
  evidence, generated reports, dataset artifacts). PostgreSQL stores only keys.

---

## Project structure

```
src/
  app.ts                 Express app (middleware, routes, error handling)
  server.ts              HTTP bootstrap + graceful shutdown
  routes.ts              API router aggregation (mounts all modules)
  config/                env, database (Prisma), redis, mongodb, storage, queues, logger
  common/                errors, constants, types, utils
  middleware/            auth (JWT+RBAC), error, validation, rate-limit, upload
  modules/
    auth/ users/ roles/ products/ inspections/ scans/ declarations/
    compliance/          rule-engine/ (pure), measurement.service, compliance.service
    rules/ violations/ evidence/ reports/ datasets/ ml/ dashboard/ audit/
  workers/               scan / ocr / compliance / report BullMQ workers + index
prisma/
  schema.prisma          22 entities + enums
  seed.ts                roles, categories, example rules, dev admin
tests/
  unit/                  rule engine, normalization, measurement
  integration/           app health (opt-in)
ml-service/              Python ML service placeholder (FastAPI stub)
```

---

## Prerequisites

- Node.js ≥ 20 and npm
- Docker + Docker Compose (for PostgreSQL, MongoDB, Redis, MinIO, ML stub)

## Quick start (Docker for infra, app on host)

```bash
# 1. Install dependencies
npm install

# 2. Configure environment
cp .env.example .env        # then edit secrets as needed

# 3. Start infrastructure (Postgres, Mongo, Redis, MinIO + bucket, ML stub)
docker compose up -d postgres mongodb redis minio minio-setup ml-service

# 4. Generate the Prisma client, run migrations, seed
npm run prisma:generate
npm run prisma:migrate      # creates prisma/migrations/<ts>_init and applies it
npm run db:seed             # roles, categories, example rules, dev admin

# 5. Run the API and the workers (two terminals)
npm run dev                 # API   → http://localhost:4000/api/v1
npm run dev:workers         # BullMQ workers (scan/ocr/compliance/report)
```

Dev admin (from seed): `admin@legalmetrology.local` / `ChangeMe123!`
(override with `ADMIN_EMAIL` / `ADMIN_PASSWORD` before seeding).

### Run everything in Docker

```bash
docker compose up --build   # brings up infra + api + workers + ml stub
```

MinIO console: http://localhost:9001 (`minioadmin` / `minioadmin`).

---

## Environment variables

See `.env.example`. Key ones:

| Variable                                   | Purpose                          |
| ------------------------------------------ | -------------------------------- |
| `DATABASE_URL`                             | PostgreSQL connection (Prisma)   |
| `MONGODB_URI`                              | MongoDB connection               |
| `REDIS_URL`                                | Redis (BullMQ, cache, limits)    |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | JWT signing secrets              |
| `S3_ENDPOINT` / `S3_REGION` / `S3_*`       | Object storage (MinIO/S3)        |
| `ML_SERVICE_URL`                           | Python ML service base URL       |
| `PORT` / `NODE_ENV`                        | Server port / environment        |

**Never commit `.env`.** Secrets are environment-based only.

---

## Prisma migrations & seed

```bash
npm run prisma:generate     # regenerate client after schema changes
npm run prisma:migrate      # dev: create + apply a migration
npm run prisma:deploy       # prod/CI: apply committed migrations
npm run db:seed             # idempotent seed
npm run prisma:studio       # inspect data
```

The initial `prisma/migrations` folder is empty by design — generate the first
migration locally against a running database, then commit it.

> **Legal rules are seeded as EXAMPLE templates only.** They are not an
> authoritative statement of the law. Populate/verify the legal-rule repository
> from the authoritative Act, Rules and amendments before any real use.

---

## API surface (prefix `/api/v1`)

| Group        | Endpoints                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------- |
| Auth         | `POST /auth/register` `/login` `/refresh` `/logout`                                         |
| Users        | `GET /users` · `GET /users/:id` · `PATCH /users/:id`                                        |
| Roles        | `GET /roles` · `POST /roles`                                                                 |
| Products     | `POST /products` · `GET /products` · `GET /products/:id` · `PATCH /products/:id`            |
| Inspections  | `POST /inspections` · `GET /inspections` · `GET /inspections/:id` · `PATCH /inspections/:id`|
| Scans        | `POST /inspections/:id/images` · `POST /inspections/:id/process` · `GET /inspections/:id/jobs` |
| Declarations | `GET /inspections/:id/declarations` · `PATCH /declarations/:id`                             |
| Compliance   | `POST /inspections/:id/check` · `GET /inspections/:id/checks` · `GET /inspections/:id/violations` |
| Rules        | `GET /rules` · `GET /rules/:id` · `POST /rules` · `PATCH /rules/:id`                        |
| Evidence     | `POST /inspections/:id/evidence` · `GET /inspections/:id/evidence`                          |
| Reports      | `POST /inspections/:id/reports` · `GET /inspections/:id/reports` · `GET /reports/:id`       |
| Datasets     | `POST /datasets` · `GET /datasets` · `POST /datasets/:id/items` · `GET /datasets/:id/items` · `POST /dataset-items/:id/annotations` |
| Dashboard    | `GET /dashboard/summary` · `/violations` · `/inspections`                                   |
| Audit        | `GET /audit-logs`                                                                            |

All responses use a consistent envelope:

```json
{ "success": true, "data": { } }
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "…", "details": [] } }
```

### Async processing flow

Uploads and expensive work are **never** done synchronously in the request:

```
POST /inspections/:id/images  → store in S3 → create InspectionImage + MLJob
                              → enqueue BullMQ job → return { jobId }
scan worker → Python ML → save raw output to MongoDB → normalize declarations
           → enqueue compliance-check → compliance worker → checks + violations
```

Queues: `scan-processing`, `ocr-processing`, `ml-inference`, `compliance-check`,
`report-generation`.

---

## The ML service contract

Node talks to the Python service (only from `src/modules/ml/ml.client.ts`) via:

```
POST /ml/ocr             POST /ml/detect-label
POST /ml/extract-fields  POST /ml/analyze
GET  /health
```

`ml-service/` contains a FastAPI **stub** returning deterministic mock structures
so the full pipeline runs locally. Replace it with the real service; the Node
backend only consumes structured results and never contains ML logic.

---

## Authentication & RBAC

- JWT access token + rotating refresh token (refresh hashes stored in Postgres).
- Passwords hashed with Argon2id.
- Roles are **extensible** (seeded: `ADMIN`, `INSPECTOR`, `REVIEWER`, `SUPERVISOR`).
- Authorization is via reusable middleware/policies (`authorize`,
  `requirePermission`) — **not** hard-coded in controllers.

## Manual review

Low-confidence OCR and ambiguous measurement cases are routed to `MANUAL_REVIEW`
and are **never** auto-converted into legal violations. Reviewers inspect the
original/cropped image, OCR text and bounding boxes, edit extracted values
(`PATCH /declarations/:id`), and confirm/reject violations (`PATCH /violations/:id`).

## Measurements

`measurement.service` estimates real-world sizes (e.g. text height) from bounding
boxes — **only** when a valid scale reference exists. Without one it returns
`hasValidScale: false`; the system never pretends pixel→mm conversion is accurate.
Measurement logic is kept separate from OCR and from the rule engine.

---

## Testing

```bash
npm test                    # unit tests (rule engine, normalization, measurement)
RUN_INTEGRATION=1 npm run test:integration   # opt-in; needs the local stack up
npm run typecheck           # tsc --noEmit
```

Unit tests cover the deterministic core (pure, infrastructure-free). Integration
tests are opt-in because they boot the app (Redis/DB clients).

---

## Security

Helmet, configurable CORS, Redis-backed rate limiting, Zod request validation,
Argon2 password hashing, JWT validation, RBAC, file MIME/type + size limits,
signed object-storage URLs, environment-based secrets (none committed), and no
stack traces in production responses.

---

## Status

This is an intentional **skeleton**, not a finished product. Notable stubs/TODOs:

- `ml-service/` is a mock; wire the real Python service.
- Report rendering (`report.renderer.ts`) emits placeholder HTML — plug in a real
  PDF/DOCX renderer.
- Legal rules are example templates — replace with authoritative rules.
- Add migrations (generated against a live DB) and expand test coverage.
```
