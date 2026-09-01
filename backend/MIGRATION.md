# Node.js → Go Migration Mapping

This documents how every piece of the original Node.js/Express/Prisma backend
(`../src`, `../prisma`) maps onto the new Go/Gin/sqlc backend (`./`). The
business/domain model, API surface and pipeline architecture are preserved
exactly; only the implementation language and a few infrastructure choices
changed (BullMQ → Asynq, Prisma → sqlc, Mongoose → the official Mongo driver).

## Module → package mapping

| Node module (`src/`)                     | Go package (`backend/`)                              |
| ----------------------------------------- | ----------------------------------------------------- |
| `app.ts`, `server.ts`, `routes.ts`        | `cmd/api/main.go` (router assembly + bootstrap)        |
| `workers/index.ts`                        | `cmd/worker/main.go`                                   |
| `prisma/seed.ts`                          | `cmd/seed/main.go`                                     |
| `config/env.ts`                           | `internal/config`                                      |
| `config/database.ts` (Prisma client)      | `pkg/postgres` + `internal/db` (sqlc-generated)         |
| `config/mongodb.ts`                       | `pkg/mongodb`                                          |
| `config/redis.ts`                         | `pkg/redisclient`                                      |
| `config/storage.ts`                       | `pkg/storage`                                          |
| `config/queues.ts`                        | `internal/taskqueue`                                   |
| `config/logger.ts`                        | `pkg/logger` (`log/slog`)                               |
| `common/errors/*`                         | `internal/apperror`                                     |
| `common/utils/response.ts`                | `internal/response`                                     |
| `common/utils/pagination.ts`              | `internal/pagination`                                   |
| `common/constants/*`                      | `internal/constants`                                    |
| `common/utils/http-client.ts`             | `internal/ml/client.go` (inlined; only one caller)      |
| `common/utils/crypto.ts`                  | `internal/auth/crypto.go`, `internal/dbutil.GenerateCode` |
| `middleware/auth.middleware.ts`           | `internal/middleware/auth.go`                           |
| `middleware/error.middleware.ts`          | `internal/middleware/error.go`                          |
| `middleware/rate-limit.middleware.ts`     | `internal/middleware/ratelimit.go`                      |
| `middleware/upload.middleware.ts`         | `internal/middleware/upload.go`                         |
| `middleware/validation.middleware.ts`     | `internal/httpapi/bind.go` + Gin `binding` struct tags   |
| `modules/auth/*`                          | `internal/auth` + `internal/authtoken` (JWT split out to avoid an import cycle with `internal/app`) |
| `modules/users/*`                         | `internal/users`                                        |
| `modules/roles/*`                         | `internal/roles`                                        |
| `modules/products/*`                      | `internal/products`                                     |
| `modules/inspections/*`                   | `internal/inspections`                                  |
| `modules/scans/*`                         | `internal/scans` (service.go = API-side enqueue; pipeline.go = worker-side orchestration, was `scan.pipeline.ts`) |
| `modules/declarations/*`                  | `internal/declarations` (service.go + normalization.go) |
| `modules/compliance/rule-engine/*`        | `internal/compliance/engine` (pure, zero dependencies, ported line-for-line) |
| `modules/compliance/measurement.service.ts` | `internal/compliance/measurement.go`                  |
| `modules/compliance/compliance.service.ts`  | `internal/compliance/service.go`                       |
| `modules/rules/*`                         | `internal/rules`                                        |
| `modules/violations/*`                    | `internal/violations`                                   |
| `modules/evidence/*`                      | `internal/evidence`                                     |
| `modules/reports/*`                       | `internal/reports` (service.go + renderer.go)            |
| `modules/datasets/*`                      | `internal/datasets`                                      |
| `modules/ml/ml.client.ts`                 | `internal/ml/client.go`                                  |
| `modules/ml/ml.types.ts`                  | `internal/ml/types.go`                                   |
| `modules/ml/ml.mongo.ts`, `ml.repository.ts` | `internal/ml/mongo.go`                                |
| `modules/dashboard/*`                     | `internal/dashboard`                                     |
| `modules/audit/*`                         | `internal/audit`                                         |
| `workers/scan.worker.ts`, `ocr.worker.ts`, `compliance.worker.ts`, `report.worker.ts` | `internal/workers/server.go` (one `asynq.ServeMux` with 4 handlers, dispatching into the packages above) |
| *(new)* shared HTTP dependency injection | `internal/app` (Deps struct — Node had no equivalent; DI was implicit via module-level singletons like `prisma`) |
| *(new)* shared DTO/UUID conversion helpers | `internal/dbutil`, `internal/httpapi` (Node didn't need these; Prisma/Zod handled coercion implicitly) |

## Prisma model → PostgreSQL table

Every Prisma model became a table in `migrations/0001_init.up.sql` with the
same name (snake_cased) and the same columns, indexes and foreign keys. No
entity was dropped. Prisma enums became native Postgres `ENUM` types.

| Prisma model            | Postgres table            |
| ------------------------ | -------------------------- |
| `User`                   | `users`                     |
| `Role`                   | `roles`                     |
| `UserRole`               | `user_roles`                |
| `RefreshToken`           | `refresh_tokens`            |
| `ProductCategory`        | `product_categories`        |
| `Manufacturer`           | `manufacturers`             |
| `Product`                | `products`                  |
| `ProductMedia`           | `product_media`             |
| `Inspection`             | `inspections`               |
| `InspectionImage`        | `inspection_images`         |
| `ExtractedDeclaration`   | `extracted_declarations`    |
| `LegalRule`              | `legal_rules`                |
| `RuleRequirement`        | `rule_requirements`          |
| `RuleCategory`           | `rule_categories`            |
| `ComplianceCheck`        | `compliance_checks`          |
| `Violation`              | `violations`                 |
| `Evidence`               | `evidence`                   |
| `MLJob`                  | `ml_jobs`                    |
| `Report`                 | `reports`                    |
| `Dataset`                | `datasets`                   |
| `DatasetVersion`         | `dataset_versions`           |
| `DatasetItem`            | `dataset_items`              |
| `DatasetAnnotation`      | `dataset_annotations`        |
| `AuditLog`               | `audit_logs`                 |

`ProductMedia` and `Manufacturer` had no dedicated Node controller either
(only referenced via `Product`); the tables and FKs exist for schema parity,
CRUD is exposed only where the original API exposed it (rely on the FK
constraint + centralized error mapping for referential-integrity errors,
same as Prisma's `P2003` → `409` mapping in `error.middleware.ts`).

## Prisma query → sqlc query

Every Prisma call (`prisma.<model>.findMany/create/update/...`, often with
`include`/`select`) became one or more named queries in `queries/*.sql`,
compiled by `sqlc generate` into typed methods on `*db.Queries` in
`internal/db`. Patterns used throughout:

- **Optional filters** (Prisma's conditional `where` spread): `sqlc.narg('x')`
  combined with `($n::type IS NULL OR column = $n)`.
- **Partial updates** (Prisma's "only spread defined fields"): `COALESCE($n,
  column)` for every optional column, with an explicit `::type` cast whenever
  the fallback is a literal (`COALESCE($n::violation_severity, 'MEDIUM')`) —
  without the cast, sqlc can't infer the parameter type and falls back to
  `interface{}`.
- **Prisma `include`** (nested relations in one round trip): denormalized
  single queries with `LEFT JOIN` + scalar subselects for counts (see
  `GetInspectionDetail`, `GetProductByID`, `ListLegalRules`), which is a
  single round trip just like Prisma's generated SQL — arguably more
  transparent than Prisma's join-then-reshape.
- **Prisma transactions** (`prisma.$transaction`): `pgxpool.Pool.Begin` +
  `queries.WithTx(tx)`, used in `internal/users` (role reassignment) and
  `internal/rules` (requirement/category replacement).
- **`skipDuplicates`**: `ON CONFLICT ... DO NOTHING`.
- **Prisma `Decimal`**: mapped to `pgtype.Numeric` by sqlc (no override
  applies cleanly to `numeric` columns); converted to `*float64` via
  `db.NumericToFloat64`/`db.Float64ToNumeric` at the service boundary —
  matching the Node code's own `Number(decimal)` conversion (neither
  preserves arbitrary precision; the domain doesn't need it).
- **Prisma `Json`**: mapped to `[]byte` (raw JSONB) via an sqlc override;
  marshaled/unmarshaled with `internal/dbutil.JSONBytes` and
  `encoding/json` at the boundary, same as Node's `as never`-cast JSON fields.

## BullMQ job → Asynq task

| BullMQ queue (`config/queues.ts`) | Asynq queue (`internal/taskqueue`) | Task type              |
| ----------------------------------- | ------------------------------------ | ----------------------- |
| `scan-processing`                   | `QueueScan`                          | `TaskScanProcess`        |
| `ocr-processing`                    | `QueueOCR`                           | `TaskOCRProcess`         |
| `ml-inference`                      | *(dropped — see below)*              | —                        |
| `compliance-check`                  | `QueueCompliance`                    | `TaskComplianceCheck`    |
| `report-generation`                 | `QueueReport`                        | `TaskReportGenerate`     |

The Node `ml-inference` queue was declared in `config/queues.ts` but never
actually used by any producer or worker (the scan pipeline calls `mlClient`
directly, not via that queue) — dropped as dead code rather than carried
over. Job payloads became typed structs (`taskqueue.ScanPayload`,
`OCRPayload`, `CompliancePayload`, `ReportPayload`) serialized as JSON task
payloads, same shape as the original `JobData` interfaces.

## Node middleware → Go middleware

| Node (Express)                         | Go (Gin)                                                        |
| ---------------------------------------- | ------------------------------------------------------------------ |
| `authenticate` / `optionalAuth`          | `middleware.Authenticate` / `middleware.OptionalAuth`                |
| `authorize(...roles)`                    | `middleware.Authorize(...roles)`                                     |
| `requirePermission(...perms)`            | `middleware.RequirePermission(...perms)`                             |
| `errorHandler` / `notFoundHandler`       | `middleware.ErrorHandler` / `middleware.NotFoundHandler`              |
| `globalRateLimiter` / `authRateLimiter`  | `middleware.RateLimit` (parameterized, Redis fixed-window)            |
| `uploadSingleImage` (Multer, in-memory)  | `middleware.RequireSingleImage` (Gin multipart, in-memory)            |
| `validate({ body, query, params })` (Zod) | Gin `binding` struct tags + `httpapi.BindJSON/BindQuery/Bind`; UUID path params via `httpapi.ParseUUIDParam` |
| `helmet()`                                | Not ported 1:1 — Gin has no direct equivalent; CORS + no-stack-traces-in-prod cover the same threat model. Add `gin-contrib/secure` if header-hardening parity is required. |
| `pino-http` (request logging)             | `middleware.RequestLogger` (`log/slog`, same request-id + duration + status fields) |

## Node service → Go service, Node controller → Go handler

Every `<module>.service.ts` became a `Service` struct in
`internal/<module>/service.go` with the same method names translated to
Go idiom (`listUsers` → `List`, `getInspection` → `Get`, etc.), taking
`context.Context` first and returning `(result, error)`. Every
`<module>.controller.ts` + `<module>.routes.ts` pair became a single
`internal/<module>/routes.go` with a `RegisterRoutes(rg *gin.RouterGroup, d
*app.Deps, ...)` function — Gin closures replace Express handler functions,
but the request flow (validate → call service → audit → respond) is
identical at every call site.

Cross-module dependencies (e.g. `scans` needing `inspections.ensureInspection`)
are passed explicitly as constructor/function parameters instead of Node's
implicit "import the other module's service" pattern — Go's stricter import
graph forced this to be explicit, which surfaced two import cycles during the
port that Node's structural typing let pass silently:

1. `internal/app` (shared Deps, holds a token verifier for middleware) vs.
   `internal/auth` (business logic, which itself registers routes via
   `internal/app`) → resolved by extracting `internal/authtoken` (just the
   JWT type) so both can depend on it without depending on each other.
2. `internal/scans` (HTTP-side producer, needs task/queue names) vs.
   `internal/workers` (consumes those constants AND imports `internal/scans`
   for the pipeline) → resolved by extracting `internal/taskqueue` (task
   names + payload structs only) as the shared leaf dependency.

## What did NOT carry over

- **BullMQ itself** — replaced by Asynq per the migration brief. The
  `ml-inference` queue (dead code, see above) was dropped entirely rather
  than ported.
- **Prisma** — replaced by sqlc + pgx. No Prisma-style runtime query builder
  exists in the Go backend; all queries are static SQL reviewed at build time.
- **Mongoose schemas** (`ml.mongo.ts`) — replaced by plain structs +
  the official `go.mongodb.org/mongo-driver`. The Go repository
  (`internal/ml/mongo.go`) additionally round-trips the ML payload through
  `encoding/json` before insert so the stored document keeps the exact
  camelCase field names the ML service returned, instead of the Mongo
  driver's default lowercased-struct-field naming — a deliberate fidelity
  fix, not a behavior change.
- **`helmet()`** header hardening — no direct Gin equivalent was added (see
  middleware table above); flagged rather than silently dropped.
- **Zod schemas** — replaced by Gin `binding` struct tags backed by
  `go-playground/validator`. Validation error *shape* (`{ field, message,
  code }`) is preserved by `internal/httpapi.translateBindErr`.
- **`pino`/`pino-http`** — replaced by the standard library's `log/slog`,
  per the target stack (avoids an extra structured-logging dependency).
- **The Node process itself** — after this migration, `../src`, `../prisma`,
  and the Node `package.json`/`Dockerfile`/`docker-compose.yml` at the repo
  root are no longer required to run the application. They have been left in
  place for historical reference rather than deleted; deleting them is a
  separate, explicit decision for whoever owns the repository.

## What is unchanged by design

- **The Python ML service** (`../ml-service`) — untouched. The Go backend
  talks to it exactly as Node did: one HTTP client (`internal/ml/client.go`),
  same four endpoints, same JSON contract (`internal/ml/types.go` is a
  field-for-field port of `ml.types.ts`).
- **The deterministic rule engine** (`internal/compliance/engine`) — ported
  operator-for-operator, including the `MANUAL_REVIEW:`-prefixed reason
  convention used internally to distinguish "not a definite failure" from a
  real `FAIL`, and the low-confidence-routes-to-manual-review behavior. Unit
  tests in `engine_test.go` mirror `tests/unit/rule-engine.test.ts` case for
  case.
- **The API route table** — identical paths, methods and permission scopes;
  see `README.md` for the full list.
- **The response envelope** (`{ success, data }` / `{ success, error }`) and
  audit-log shape.
