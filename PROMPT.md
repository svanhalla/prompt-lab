# Prompt for Amazon Q Chat: Build a Production-Quality Go CLI + API + Web UI

You are an expert senior Go (Golang) engineer. Create a **production-quality** project scaffold and implementation that meets the following requirements.

## High-Level Goal

Build a **Go 1.25.1** application that is both:

1) A **CLI** (using **Cobra** + **Viper**) with subcommands:
- `health` — returns application health information.
- `hello` — prints a friendly greeting (optionally accepts a `--name` flag).
- `set message` — stores a message that the API and Web UI will serve (persisted to disk as JSON).  
   - Expose this via the CLI and also as an **HTTP POST** endpoint in the API.

2) An **HTTP API + Web server** (using **Echo**) that:
- Serves endpoints for `health`, `hello`, `GET /message`, `POST /message`.
- Renders an HTML page (Go **html/template**) that shows the currently stored message and a form to update it.
- Provides a page to **view the application logs nicely**, styled with **Tailwind CSS**.
- Persists:
  - **Message data** to a JSON file on disk.
  - **Application settings/config** to a JSON file on disk; if missing, create with **sensible defaults**.

The project must be **well-structured**, **git-versioned**, documented, and ready to push to GitHub.

---

## Technical Requirements

### Languages, Frameworks, and Tooling
- **Go**: version **1.25.1** (ensure `go.mod` enforces this).
- **CLI**: **Cobra** for commands, **Viper** for configuration (env vars + flags + config file).
- **Web server**: **Echo**.
- **Templates**: Go `html/template` for server-side rendering.
- **Styling**: **Tailwind CSS** (set up a minimal build pipeline).
- **Logging**: **logrus** (JSON formatter in production; human-readable in dev).
- **Linting**: **golangci-lint** with a sensible, strict config that still allows velocity.
- **Tests**: Unit tests with table-driven style; include coverage targets.
- **OpenAPI**: Provide a complete **OpenAPI 3.1** spec for the API.

### CLI Design
- Root command name: suggest a **sane, short, memorable** name (e.g., `greetd`, `gomsg`, `sayso`, or similar). Document your choice.
- Subcommands:
  - `health`: prints JSON health output (status, uptime, version, git commit, build time).
  - `hello [--name NAME]`: prints “Hello, NAME!” or “Hello, World!” if omitted.
  - `set message <text>`: persists the message to disk (JSON).
  - `api [--host 0.0.0.0 --port 8080]`: starts the API + Web server.
- Config precedence: **flags > env vars > config file defaults**.
- Add `--log-level`, `--log-format` (json|text), `--config` path flag.
- Add `version` subcommand (prints version, commit, build date).
- All log messages and comments in **English**.

### API Endpoints (Echo)
- `GET /health` → JSON with health info (status, uptime, version, commit, build time).
- `GET /hello?name=...` → JSON with greeting.
- `GET /message` → JSON `{ "message": "..." }`.
- `POST /message` → Accepts JSON `{ "message": "..." }`; validates, persists to disk, returns updated value.
- `GET /logs` → Returns an HTML page that nicely renders recent application logs (no streaming required; read from a file with safe tailing).
- `GET /ui` → HTML page (Go templates) that:
  - Shows current message.
  - Has a form to update message (POSTs to `/message`).
  - Includes Tailwind styling; keep the page clean and modern.

### Persistence
- Default data directory: `~/.<appname>/`
  - `config.json` — application settings.
  - `message.json` — stored message document.
  - `app.log` — rolling log file (size-based rotate is nice-to-have; implement if simple).
- If `config.json` is missing, create it with defaults on first run. Include sensible defaults for host, port, data paths, log level, and log format.

### Configuration
- Use **Viper** to read:
  - Config file (`config.json` in the data directory by default; allow `--config` override).
  - Environment variables with a clear prefix (e.g., `APP_`).
  - CLI flags on each command.
- Provide a **default configuration** struct and a function to load/validate it.
- Include schema/comments in the README for each config field.

### Logging
- **logrus** with:
  - `--log-level` (debug, info, warn, error).
  - `--log-format` (json|text).
  - Log file at `app.log` in data directory; also log to stdout.
- Add request logging middleware for Echo (without leaking PII).
- Include structured fields (request id, path, method, status, latency).

### OpenAPI Spec
- Author a **single source of truth** OpenAPI 3.1 YAML in `api/openapi.yaml` documenting:
  - `/health` (GET)
  - `/hello` (GET with `name` query)
  - `/message` (GET, POST with JSON body)
- Add examples, response schemas, and error models.

### Project Structure
Propose and implement a clean structure, for example:
```
<appname>/
  cmd/<appname>/           # Cobra root command, main.go
  internal/                # internal packages
    api/                   # Echo server setup, handlers, routes
    config/                # config types, load/validate defaults
    logging/               # logrus setup
    storage/               # persistence of config + message
    web/                   # templates, Tailwind assets, helpers
    version/               # version info (ldflags)
  api/
    openapi.yaml
  web/
    templates/             # *.tmpl
    static/                # compiled Tailwind CSS
    tailwind.config.js
    package.json
    postcss.config.js
  build/
    scripts/
  .github/workflows/
  .golangci.yml
  Makefile
  Dockerfile
  .env.example
  README.md
  LICENSE
```

### Build, Run, and CI/CD
- **Makefile** targets:
  - `make deps`, `make build`, `make run`, `make lint`, `make test`, `make cover`, `make tailwind`, `make api`, `make cli`
- **Versioning**: Inject `version`, `commit`, `buildTime` via `-ldflags`.
- **Dockerfile**: Multi-stage, minimal final image (non-root user). Expose the API port.
- **GitHub Actions**:
  - `ci.yml` with steps: setup Go 1.25.1, cache, lint, test (with coverage), build.
  - Optional: Tailwind build step that runs on pushes (keep deterministic).
- **golangci-lint**: enable common linters (staticcheck, revive, gofumpt, govet, errcheck, gocyclo with reasonable thresholds, mnd tuned to allow HTTP codes). Ensure `make lint` passes on the generated code.

### Security & Quality
- Input validation (for POST /message).
- Graceful shutdown (context + timeouts).
- Sensible defaults and clear error messages.
- Avoid global state; dependency injection where practical.
- Unit tests for:
  - Config load/merge precedence.
  - Storage read/write.
  - Handlers for `/message` and `/hello`.
- Simple e2e test (spin up server on ephemeral port, hit endpoints).

### Developer Experience
- **README.md** with:
  - Quick start.
  - CLI usage examples.
  - Config reference.
  - API docs pointer (OpenAPI).
  - Build/run with and without Docker.
- **.env.example** with typical environment variables.
- **LICENSE** (MIT by default unless you prefer Apache-2.0).
- Comments and printed output in **English**.

### Nice-to-Haves (If Low Effort)
- Log file rotation (size-based using a small dependency like `lumberjack`).
- Simple CSRF protection on the message update form.
- Basic health details (goroutines, heap allocs) gated behind a `--debug` flag.

---

## Acceptance Criteria

- Running `make build` produces a working binary.
- `myapp version` prints version, commit, build time.
- `myapp set message "Hi"` stores to `message.json`.
- `myapp api --host 0.0.0.0 --port 8080` serves:
  - `GET /health` returns status + version info.
  - `GET /hello?name=Hans` returns greeting JSON.
  - `GET /message` returns the stored message.
  - `POST /message` updates the message on disk.
  - `GET /ui` shows the message + update form (Tailwind styled).
  - `GET /logs` shows recent logs with nice formatting.
- `golangci-lint` passes via `make lint`.
- Unit tests pass with `make test` and coverage report is generated.
- OpenAPI 3.1 spec exists at `api/openapi.yaml` and matches implemented endpoints.
- Project is initialized as a **git** repo with a first commit and `.gitignore`.

---

## Deliverables

1) Complete repository with source code, tests, OpenAPI spec, Makefile, Dockerfile, Tailwind config, GitHub Actions workflow, `.golangci.yml`, `.env.example`, README, LICENSE.  
2) Clear instructions in README to install, configure, run CLI and API, and build Tailwind assets.  
3) Justify the chosen **app name** and any architectural decisions in the README.

> If any requirement is ambiguous or missing, make senior-level decisions and document them.
