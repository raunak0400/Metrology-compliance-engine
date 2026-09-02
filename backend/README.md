# Legal Metrology Packaged Commodities Compliance Checker — Go Backend

Go rewrite of the Node.js modular monolith. Same domain, same architecture,
same API surface — different runtime. See [`MIGRATION.md`](./MIGRATION.md)
for the detailed Node → Go mapping.

> **Domain note:** This is a **Department of Consumer Affairs / Legal
> Metrology** system. It is **not** an FSSAI system. It verifies mandatory
> declarations on packages (MRP, net quantity, manufacturer/packer/importer,
> dates, consumer care, country of origin, etc.), not food-safety licensing.

Computer vision / OCR / model inference lives in a **separate Python ML
service** (`../ml-service`, unchanged by this migration).

---

## Core design principle

The architecture strictly separates *extraction* (AI/ML) from *decision*
(deterministic rules). **An ML model or LLM never decides legal compliance.**

```
IMAGE
  ↓  (Python ML service: preprocess, skew/perspective, denoise, contrast,
  ↓   label/panel detection, OCR, NER/classification)
RAW ML RESULT            → stored in MongoDB (evidence trail)
  ↓  NORMALIZATION        (deterministic, Go)
STRUCTURED DECLARATIONS + MEASUREMENTS   → stored in PostgreSQL
  ↓  DETERMINISTIC LEGAL RULE ENGINE     (pure, testable, Go)
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

| Concern            | Choice                                              |
| ------------------ | ---------------------------------------------------- |
| Runtime / language | Go 1.23+                                              |
| HTTP framework     | Gin (REST)                                            |
| Relational DB      | PostgreSQL via pgx/v5 + sqlc (source of truth)        |
| Document store     | MongoDB (raw OCR/CV/ML output)                        |
| Cache / queues     | Redis + Asynq (jobs, cache, rate limiting)            |
| Object storage     | S3-compatible — MinIO (dev), AWS S3 (prod)            |
| ML service         | Separate Python HTTP service (OpenCV/PaddleOCR/…)     |
| Auth               | JWT access + rotating refresh, Argon2id hashing       |
| Validation         | go-playground/validator (via Gin `binding` tags)      |
| Logging            | `log/slog` (structured, request-scoped)               |

## Responsibilities by store

- **PostgreSQL (authoritative):** users, RBAC, products, inspections,
  normalized declarations, legal rules + requirements, compliance checks,
  violations, evidence references, ML job state, reports, datasets,
  annotations, audit logs.
- **MongoDB (flexible):** raw OCR text/blocks/bboxes/confidence, detected
  regions, preprocessing metadata, model predictions, field-extraction
  results. Never the authoritative compliance result.
- **Object storage:** the actual binary files (original/label/cropped images,
  evidence, generated reports, dataset artifacts). PostgreSQL stores only keys.

---

## Project structure

```
backend/
  cmd/
    api/        HTTP server bootstrap (Gin app, middleware, graceful shutdown)
    worker/     Asynq worker fleet bootstrap
    seed/       One-off DB seed (roles, categories, example rules, dev admin)
  internal/
    app/                shared Deps container injected into every module
    authtoken/          JWT sign/verify (split out to avoid an import cycle)
    auth/ users/ roles/ products/ inspections/ scans/ declarations/
    compliance/         engine/ (pure rule engine) + measurement + service
    rules/ violations/ evidence/ reports/ datasets/ ml/ dashboard/ audit/
    workers/            Asynq handler registration (dispatches to services above)
    taskqueue/          Asynq task/queue name + payload contracts (shared)
    middleware/         auth (JWT+RBAC), error, rate-limit, request logging, upload
    db/                 sqlc-generated PostgreSQL access (DO NOT EDIT by hand)
    apperror/ response/ pagination/ dbutil/ httpapi/ constants/ config/
  pkg/
    postgres/ mongodb/ redisclient/ storage/ logger/    infra clients
  migrations/           SQL migrations (source of truth for the schema)
  queries/              sqlc query definitions (source for internal/db)
  sqlc.yaml
  ../ml-service/        Python ML service (unchanged, sibling directory)
```

---

## Prerequisites

- Go ≥ 1.23
- Docker + Docker Compose (for PostgreSQL, MongoDB, Redis, MinIO, ML stub)
- [sqlc](https://sqlc.dev) (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`) — only needed if you change `queries/` or `migrations/`
- A Postgres migration tool of your choice — commands below use plain `psql`; [golang-migrate](https://github.com/golang-migrate/migrate) works too if you prefer versioned up/down migrations

## Quick start (Docker for infra, app on host)

```bash
# 1. Configure environment
cp .env.example .env        # then edit secrets as needed

# 2. Start infrastructure (Postgres, Mongo, Redis, MinIO + bucket, ML stub)
docker compose up -d postgres mongodb redis minio minio-setup ml-service

# 3. Apply migrations (tracked in a schema_migrations table via golang-migrate
#    — safe to re-run; already-applied versions are skipped, not re-executed)
go run ./cmd/migrate up

# 4. (Only if you changed queries/ or migrations/) regenerate sqlc code
sqlc generate

# 5. Seed roles, categories, dev admin (set SEED_EXAMPLE_RULES=true to also
#    seed illustrative EXAMPLE legal rule templates for local dev — see
#    cmd/seed; leave unset for any environment used for real compliance work)
go run ./cmd/seed

# 6. Run the API and the worker (two terminals)
go run ./cmd/api          # API   → http://localhost:4000/api/v1
go run ./cmd/worker       # Asynq workers (scan/ocr/compliance/report)
```

Dev admin (from seed): `admin@legalmetrology.local` / `ChangeMe123!`
(override with `ADMIN_EMAIL` / `ADMIN_PASSWORD` before seeding).

### Run everything in Docker

```bash
docker compose up --build   # infra -> migrate (one-shot) -> api + worker + ml stub
```

`api` and `worker` wait on `migrate` (applies `migrations/*.sql`, then exits)
and `minio-setup` (creates the bucket, then exits) completing successfully,
which themselves wait on Postgres/Mongo/Redis/MinIO healthchecks — so
`docker compose up` brings the whole stack up in the right order with no
manual steps.

MinIO console: http://localhost:9001 (`minioadmin` / `minioadmin`).

---

## Environment variables

See `.env.example`. Key ones:

| Variable                                   | Purpose                          |
| ------------------------------------------- | -------------------------------- |
| `DATABASE_URL`                             | PostgreSQL connection (pgx)       |
| `MONGODB_URI`                              | MongoDB connection                |
| `REDIS_URL`                                | Redis (Asynq, cache, rate limits) |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | JWT signing secrets               |
| `S3_ENDPOINT` / `S3_REGION` / `S3_*`       | Object storage (MinIO/S3)         |
| `ML_SERVICE_URL`                           | Python ML service base URL        |
| `PORT` / `GO_ENV`                          | Server port / environment         |

**Never commit `.env`.** Secrets are environment-based only.

---

## Database migrations & sqlc

The SQL migration in `migrations/0001_init.up.sql` mirrors the original
Prisma schema table-for-table (see `MIGRATION.md`). It is the schema source
of truth; `queries/*.sql` define every query; `sqlc generate` produces the
type-safe Go bindings in `internal/db` (checked in, but regenerate after
editing `queries/` or `migrations/`).

`cmd/migrate` (backed by [golang-migrate](https://github.com/golang-migrate/migrate))
applies `migrations/*.up.sql`/`*.down.sql` and tracks the applied version in a
`schema_migrations` table, so `up` is safe to run repeatedly — already-applied
versions are skipped, never re-executed:

```bash
go run ./cmd/migrate up            # apply all pending migrations
go run ./cmd/migrate down          # roll back one migration (dev only)
go run ./cmd/migrate version       # print current schema version
sqlc generate                      # regenerate internal/db after editing queries/ or migrations/
```

> **Legal rules are seeded as EXAMPLE templates only** (see `cmd/seed`). They
> are not an authoritative statement of the law. Populate/verify the
> legal-rule repository from the authoritative Act, Rules and amendments
> before any real use.

---

## API surface (prefix `/api/v1`)

Identical to the original Node service — see `MIGRATION.md` for the full
route table. All responses use a consistent envelope:

```json
{ "success": true, "data": {} }
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "…", "details": [] } }
```

### Async processing flow

Uploads and expensive work are **never** done synchronously in the request:

```
POST /inspections/:id/images  → store in S3 → create InspectionImage + MLJob
                              → enqueue Asynq task → return { jobId }
scan worker → Python ML → save raw output to MongoDB → normalize declarations
           → enqueue compliance-check → compliance worker → checks + violations
```

Queues: `scan-processing`, `ocr-processing`, `compliance-check`,
`report-generation` (see `internal/taskqueue`).

---

## The ML service contract

Go talks to the Python service (only from `internal/ml/client.go`) via:

```
POST /ml/ocr             POST /ml/detect-label
POST /ml/extract-fields  POST /ml/analyze
GET  /health
```

`../ml-service/` contains a FastAPI **stub** returning deterministic mock
structures so the full pipeline runs locally — unchanged by this migration.
Replace it with the real service; the Go backend only consumes structured
results and never contains ML logic.

---

## Authentication & RBAC

- JWT access token + rotating refresh token (refresh hashes stored in
  Postgres).
- Passwords hashed with Argon2id (`internal/auth/password.go`).
- Roles are **extensible** (seeded: `ADMIN`, `INSPECTOR`, `REVIEWER`,
  `SUPERVISOR`).
- Authorization is via reusable middleware/policies (`middleware.Authorize`,
  `middleware.RequirePermission`) — **not** hard-coded in handlers.

## Manual review

Low-confidence OCR and ambiguous measurement cases are routed to
`MANUAL_REVIEW` and are **never** auto-converted into legal violations.
Reviewers inspect the original/cropped image, OCR text and bounding boxes,
edit extracted values (`PATCH /declarations/:id`), and confirm/reject
violations (`PATCH /violations/:id`).

## Measurements

`internal/compliance/measurement.go` estimates real-world sizes (e.g. text
height) from bounding boxes — **only** when a valid scale reference exists.
Without one it returns `hasValidScale: false`; the system never pretends
pixel→mm conversion is accurate. Measurement logic is kept separate from OCR
and from the rule engine.

---

## Testing

```bash
go test ./...              # unit tests (rule engine, normalization, measurement, auth, tokens, config)
go vet ./...
gofmt -l .                 # should print nothing
```

Unit tests cover the deterministic core (pure, infrastructure-free):
`internal/compliance/engine`, `internal/compliance` (measurement),
`internal/declarations` (normalization), `internal/authtoken`, `internal/auth`
(password hashing), `internal/pagination`, `internal/config` (env validation).
`cmd/api` has an infra-free handler test for `/health` and the 404 envelope.

### Infrastructure connectivity tests

`pkg/postgres`, `pkg/mongodb`, `pkg/redisclient`, `pkg/storage` and
`cmd/migrate` each have a live-connectivity test that's skipped by default and
only runs when its `TEST_*` env var is set (see `.env.example`) — point these
at disposable/scratch services, never at dev data:

```bash
export TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/legal_metrology_test?sslmode=disable
export TEST_MONGODB_URI=mongodb://localhost:27017/legal_metrology_ml_test
export TEST_REDIS_URL=redis://localhost:6379/1
export TEST_S3_ENDPOINT=localhost:9000 TEST_S3_ACCESS_KEY=minioadmin TEST_S3_SECRET_KEY=minioadmin TEST_S3_BUCKET=legal-metrology-test
go test ./pkg/... ./cmd/migrate/...
```

The `cmd/migrate` test additionally proves migrations apply cleanly to a
fresh database, are idempotent on re-run, and roll back cleanly — it calls
`Drop()` on the target database, so it must point at a scratch database.

---

## Security

Gin recovery + CORS, Redis-backed rate limiting, `binding`-tag request
validation, Argon2id password hashing, JWT validation, RBAC, file
MIME/size limits, signed object-storage URLs, environment-based secrets (none
committed), and no internal error details leaked in production responses.

---

## Health checks

- `GET /health` — liveness only (process is up).
- `GET /ready` — readiness: pings PostgreSQL, Redis, MongoDB and object
  storage (bucket existence). Returns `503` with a structured error envelope
  if any dependency isn't reachable.

---

## Status

Feature-complete parity with the original Node skeleton, including its
acknowledged stubs:

- `../ml-service/` is a mock; wire the real Python service.
- Report rendering (`internal/reports/renderer.go`) emits placeholder HTML —
  plug in a real PDF/DOCX renderer (e.g. `gofpdf`, a headless-Chromium print,
  or `unidoc`/`unioffice` for DOCX).
- Legal rules are example templates — replace with authoritative rules.
- Expand test coverage (repository/handler tests need a live Postgres —
  consider `testcontainers-go` for CI).
