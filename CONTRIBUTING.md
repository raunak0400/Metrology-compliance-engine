# Contributing to the Metrology Compliance Engine

Thanks for your interest in improving the project! This guide covers how to set
up your environment, our conventions, and how to submit changes. By
participating you agree to abide by our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Ground rules

- **Never commit secrets.** Use `.env` files (git-ignored) and commit only
  `*.env.example` templates. See [`SECURITY.md`](./SECURITY.md).
- **Keep the core principle intact:** extraction (ML/OCR) and decision
  (rule engine) stay separate. ML/LLM code must never make a compliance
  verdict — that belongs to `rule_engine/`.
- Every new compliance rule must cite the exact legal clause and ship with
  tests.

## Repository structure

This is a monorepo. Each top-level directory is an independent component with
its own tooling:

| Directory         | Stack                  | Tooling                     |
|-------------------|------------------------|-----------------------------|
| `backend/`        | Go                     | `go build` / `go test`      |
| `rule_engine/`    | Python / FastAPI       | `pip` / `pytest`            |
| `ml-service/`     | Python / FastAPI       | `pip`                       |
| `frontend/`       | Next.js / TypeScript   | `npm`                       |
| `frontend-new/`   | Next.js / TypeScript   | `npm`                       |
| `ML Pipeline_v2/` | Go + Python            | `go` / `pip`                |
| `legacy-node/`    | Node.js / TypeScript   | `npm` (reference only)      |

## Development setup

See the [Getting started](./README.md#getting-started) section of the README
for full run instructions. In brief:

```bash
# Rule engine
cd rule_engine && pip install -r requirements.txt && uvicorn main:app --reload

# ML service (stub)
cd ml-service && pip install -r requirements.txt && uvicorn app:app --reload --port 8001

# Backend
cd backend && cp .env.example .env && go run ./cmd/api

# Frontend
cd frontend && npm install && npm run dev
```

## Branching & workflow

1. Fork the repo (external contributors) or create a topic branch.
2. Branch off `main` using a descriptive name:
   `feat/…`, `fix/…`, `docs/…`, `refactor/…`, `test/…`, `chore/…`.
3. Make focused commits (see below).
4. Ensure the relevant component builds and its tests pass.
5. Open a Pull Request against `main` and fill in the PR template.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<optional scope>): <short summary>

<optional body explaining what and why>
```

Common types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`,
`build`, `ci`. Scope is usually the component, e.g.
`feat(rule_engine): add Rule 9 net-quantity font-size check`.

Write clear, self-contained messages. Keep the history readable — squash
noise before opening a PR.

## Coding standards

- **Go** (`backend/`, `ML Pipeline_v2/`): run `gofmt`/`go vet`; keep packages
  cohesive; mirror the existing `internal/` layout.
- **Python** (`rule_engine/`, `ml-service/`): target Python 3.9+, use type
  hints and Pydantic v2 models; keep rules pure and deterministic.
- **TypeScript** (`frontend*/`): follow the existing ESLint/TS config; prefer
  typed props and server components where appropriate.
- Match the style of the surrounding code; do not reformat unrelated lines.

## Testing

Run the tests for any component you touch and add tests for new behavior:

```bash
cd rule_engine && pytest tests/ -v
cd backend && go test ./...
cd legacy-node && npm test
```

Rule-engine changes **must** include tests that assert the correct clause and
severity for both compliant and non-compliant inputs.

## Pull request checklist

- [ ] Branch is up to date with `main`.
- [ ] Component builds and tests pass locally.
- [ ] No secrets or generated artifacts are committed.
- [ ] New/changed behavior is covered by tests.
- [ ] Docs updated (README / component README) where relevant.
- [ ] PR description explains the what and why and links any related issue.

## Reporting bugs & requesting features

Use the GitHub issue templates. For security issues, **do not** open a public
issue — follow [`SECURITY.md`](./SECURITY.md).

Thank you for contributing! 🎉
