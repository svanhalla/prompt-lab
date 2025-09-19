# Prompt for Amazon Q Chat: Build a Production-Quality Go CLI + API + Web UI

You are an expert senior Go (Golang) engineer. Create a **production-quality** project scaffold and implementation that meets the following requirements.

## High-Level Goal

Build a **Go 1.25** application that is both:

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
- **Go**: version **1.25** (ensure `go.mod` enforces this).
- **CLI**: **Cobra** for commands, **Viper** for configuration (env vars + flags + config file).
- **Web server**: **Echo**.
- **Templates**: Go `html/template` for server-side rendering.
  - **Template Structure**: Store templates in `internal/web/templates/` directory with proper file organization
  - **Embedded Templates**: Use Go's `embed` directive to embed template files in the binary
  - **Template Manager**: Create `internal/web/templates.go` to load and manage all templates
  - **Hot Reload Development**: Implement filesystem fallback - if template files exist on disk, use them (allowing live editing); otherwise use embedded templates (production)
  - **Development Mode**: Auto-detect development mode by checking if template files exist in filesystem
  - **No Inline HTML**: Avoid inline HTML strings in Go code; use proper template files for maintainability
- **Styling**: **Tailwind CSS** (set up a minimal build pipeline).
- **Logging**: **logrus** (JSON formatter in production; human-readable in dev).
- **Linting**: **golangci-lint v2** with a sensible, strict config that still allows velocity.
  - **Config Format**: Use `version: "2"` in `.golangci.yml` (required for v2 compatibility)
  - **Linters vs Formatters**: Separate `linters:` and `formatters:` sections (v2 breaking change)
  - **Go Version**: Specify Go version in `run.go` field to match go.mod version
  - **Essential Setup**: Enable govet, ineffassign, misspell, unused, staticcheck linters; gofmt, goimports formatters
- **Tests**: Unit tests with table-driven style; include coverage targets (60% threshold).
  - **Error Handling**: Functions returning errors must be properly handled in tests (avoid "assignment mismatch" compilation errors)
  - **Test Isolation**: Use temporary directories and ephemeral ports for test isolation
  - **Coverage Focus**: Business logic coverage; main functions and CLI commands are acceptable to exclude
- **OpenAPI**: Provide a complete **OpenAPI 3.1** spec for the API.

### CLI Design
- Root command name: suggest a **sane, short, memorable** name (e.g., `greetd`, `gomsg`, `sayso`, or similar). Document your choice.
- Subcommands:
  - `health`: prints JSON health output (status, uptime, version, git commit, build time).
  - `hello [--name NAME]`: prints "Hello, NAME!" or "Hello, World!" if omitted.
  - `set message <text>`: persists the message to disk (JSON).
  - `api [--host 0.0.0.0 --port 8080]`: starts the API + Web server.
- Config precedence: **flags > env vars > config file defaults**.
- Add `--log-level`, `--log-format` (json|text), `--config` path flag.
- Add `version` subcommand (prints version, commit, build date).
- All log messages and comments in **English**.

### API Endpoints (Echo)

### API Documentation (Swagger + Redoc)
- Provide **Swagger UI** at `/swagger/*` serving the OpenAPI spec (`api/openapi.yaml`).
- Provide **Redoc** documentation at `/docs` as a static HTML page, styled cleanly, also served from the same OpenAPI spec.
- Ensure both Swagger UI and Redoc remain in sync with the OpenAPI definition.
- **Important**: Swagger/Redoc endpoints must handle missing OpenAPI spec files gracefully (return 404 with clear error message).
- **Route Ordering**: Define specific routes (like `/swagger/openapi.yaml`) before wildcard routes (`/swagger/*`) to avoid conflicts.

- Provide a **Swagger UI** served at `/swagger/*` path, exposing and serving the generated OpenAPI spec (`api/openapi.yaml`).
- Ensure the UI is accessible when the API server is started, styled cleanly, and kept in sync with the spec.

- `GET /` → Redirect to `/ui` for better user experience when hitting the root path.
- `GET /health` → JSON with health info (status, uptime, version, commit, build time).
- `GET /hello?name=...` → JSON with greeting.
- `GET /message` → JSON `{ "message": "..." }`.
- `POST /message` → Accepts JSON `{ "message": "..." }`; validates, persists to disk, returns updated value.
- `GET /logs` → Returns an HTML page that nicely renders recent application logs (no streaming required; read from a file with safe tailing).
- `GET /ui` → HTML page (Go templates) that:
  - Shows current message.
  - Has a form to update message (POSTs to `/message`).
  - Includes Tailwind styling; keep the page clean and modern.
- **404 Handler** → For browser requests, show a helpful HTML page with links to all valid endpoints instead of JSON error. For API requests (JSON Accept header), return JSON error response.

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
- The OpenAPI spec should be served via Swagger UI at `/swagger/*`.
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
    web/                   # embedded templates and assets
      templates/           # HTML template files (*.html)
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
- **Makefile** should also include `make docs` target that:
  - Validates the OpenAPI spec (`api/openapi.yaml`).
  - Serves or regenerates Swagger UI and Redoc assets so developers can preview API docs locally.

- **Makefile** targets:
  - `make deps`, `make build`, `make run`, `make lint`, `make test`, `make cover`, `make tailwind`, `make api`, `make cli`, `make docs`
    - `make docs` should: 
      - validate and bundle `api/openapi.yaml` (e.g., via `redocly lint/bundle` or `swagger-cli validate`),
      - (re)generate any embedded Swagger UI assets if needed,
      - build/update the **Redoc** static HTML page served at `/docs`,
      - ensure the running server serves **Swagger UI** at `/swagger/*` and Redoc at `/docs` from the latest spec.
  - `make deps`, `make build`, `make run`, `make lint`, `make test`, `make cover`, `make tailwind`, `make api`, `make cli`
- **Versioning**: Inject `version`, `commit`, `buildTime` via `-ldflags`.
- **Dockerfile**: Multi-stage, minimal final image (non-root user). Expose the API port.
- **GitHub Actions**:
  - `ci.yml` with steps: setup Go 1.25, cache, lint, test (with coverage), build.
  - Optional: Tailwind build step that runs on pushes (keep deterministic).
- **golangci-lint v2**: enable common linters (staticcheck, govet, ineffassign, misspell, unused) and formatters (gofmt, goimports). Use `version: "2"` config format with separate `linters:` and `formatters:` sections. Ensure `make lint` passes on the generated code.
- **Portability**: Ensure Makefile works on systems without external dependencies like `bc` command. Use portable shell commands (awk, sed) for coverage calculations.

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


### Build Verification and Testing
- Ensure that the application can be built and executed successfully on a clean environment with Go 1.25.1.
- Provide `make build` and `make run` targets that work out-of-the-box.
- Include **unit tests** (table-driven where appropriate) for all core logic (config load, storage, handlers).
- Add a `make test` target that runs all tests and generates a coverage report (`make cover` optional).
- CI pipeline must run linting, testing, and building to verify correctness before merge.
- **Test Environment Setup**: Tests should create necessary files (OpenAPI spec, config files) in temporary directories rather than relying on project files. This ensures tests are isolated and work in any environment.

### Developer Experience
- **API docs tooling**: include dev dependencies or instructions for `redocly` (or `swagger-cli`) in the README and Makefile so `make docs` works out-of-the-box.
- **Hot Reload Templates**: In development (when template files exist on filesystem), templates are reloaded on each request for instant feedback. In production (compiled binary), embedded templates are used for performance.

- **README.md** with:
  - Quick start.
  - CLI usage examples.
  - Config reference.
  - API docs pointer (OpenAPI).
  - Build/run with and without Docker.
  - **Clear endpoint documentation** with examples:
    ```
    Available endpoints:
    - http://localhost:8080/          (redirects to /ui)
    - http://localhost:8080/health    (health check)
    - http://localhost:8080/hello     (greeting API)
    - http://localhost:8080/message   (get/set message)
    - http://localhost:8080/ui        (web interface)
    - http://localhost:8080/logs      (view logs)
    - http://localhost:8080/swagger/  (Swagger UI)
    - http://localhost:8080/docs      (Redoc docs)
    ```
- **.env.example** with typical environment variables.
- **LICENSE** (MIT by default unless you prefer Apache-2.0).
- Comments and printed output in **English**.

### Nice-to-Haves (If Low Effort)
- Log file rotation (size-based using a small dependency like `lumberjack`).
- Simple CSRF protection on the message update form.
- Basic health details (goroutines, heap allocs) gated behind a `--debug` flag.

---


## Validation & Self-Testing

Before final delivery, **verify end-to-end** that the application builds, runs, and passes tests locally and in CI.

### Local Smoke Tests
- Run: `make deps && make lint && make test && make build` — all should succeed.
- CLI checks:
  - `myapp version` prints version/commit/build time.
  - `myapp hello --name Test` prints a greeting.
  - `myapp set message "Hello from tests"` persists to `message.json`.
- API checks (start server): `myapp api --host 127.0.0.1 --port 8081`
  - `curl http://127.0.0.1:8081/health` returns JSON with status OK.
  - `curl "http://127.0.0.1:8081/hello?name=Test"` returns a greeting JSON.
  - `curl http://127.0.0.1:8081/message` returns the stored message.
  - `curl -X POST -H "Content-Type: application/json" -d '{"message":"Updated"}' http://127.0.0.1:8081/message` updates and returns the message.
  - Visit `/ui`, `/logs`, `/swagger/`, and `/docs` in a browser to confirm they render correctly.
  - **Root path test**: `curl http://127.0.0.1:8081/` should redirect to `/ui` (302 status).

### Unit Tests
### Runtime Verification
- During development and CI, verify that the **API and Web server start successfully** without errors.
- Confirm that all documented endpoints return the **expected responses**:
  - `/health` → status JSON with `ok` and version info.
  - `/hello?name=Test` → greeting JSON with `Hello, Test!`.
  - `/message` → returns latest stored message.
  - `POST /message` → updates and returns message JSON.
  - `/ui` → HTML page renders current message and update form with Tailwind styling.
  - `/logs` → HTML page renders recent logs in a human-friendly format.
  - `/swagger/` → Swagger UI loads and shows API documentation.
  - `/docs` → Redoc page renders cleanly and matches the OpenAPI spec.
  - `/` → redirects to `/ui` (302 status).
- Add an automated smoke/e2e test to validate at least `/health`, `/hello`, and `/message` endpoints at runtime.
- **Test Isolation**: Runtime verification tests should create mock OpenAPI specs and other dependencies in temporary directories to ensure they work in any environment.

- Provide **table-driven unit tests** for configuration loading/precedence, storage read/write, and `/message` + `/hello` handlers.
- Add coverage reporting (`make cover`) with a **minimum coverage threshold** (e.g., 70%). Fail CI if below threshold.
- Where feasible, include a simple e2e test that spins up the API on an ephemeral port and exercises core endpoints.

### CI Verification
- GitHub Actions workflow must run: setup Go 1.25 → lint → test (with coverage) → build → (optionally) `make docs`.
- CI should **fail** on lint errors, test failures, or coverage below threshold.


## Acceptance Criteria

- Running `make build` produces a working binary.
- `myapp version` prints version, commit, build time.
- `myapp set message "Hi"` stores to `message.json`.
- `myapp api --host 0.0.0.0 --port 8080` serves:
  - `GET /` redirects to `/ui` (302 status).
  - `GET /health` returns status + version info.
  - `GET /hello?name=Hans` returns greeting JSON.
  - `GET /message` returns the stored message.
  - `POST /message` updates the message on disk.
  - `GET /ui` shows the message + update form (Tailwind styled).
  - `GET /logs` shows recent logs with nice formatting.
  - `GET /swagger/` shows Swagger UI with API documentation.
  - `GET /docs` shows Redoc documentation.
  - **404 responses**: Browser requests to non-existent paths show helpful HTML page with links to valid endpoints.
- `golangci-lint v2` passes via `make lint` with proper config format.
- Unit tests pass with `make test` and coverage report is generated (60% threshold).
- OpenAPI 3.1 spec exists at `api/openapi.yaml` and matches implemented endpoints.
- Project is initialized as a **git** repo with a first commit and `.gitignore`.

---

- CI workflow runs lint, tests (with coverage), and build successfully on a clean checkout.

- Local smoke tests (CLI commands and API endpoints) behave as documented without manual tweaks.

- Test coverage meets the configured threshold (≥ 70% by default).

- API and Web server verified to start and return expected responses during local smoke tests and CI.

- **First-time success**: A user should be able to run `make build && ./myapp api` and immediately access all endpoints without errors.
- **Hot Reload Development**: When running from source directory, template changes should be visible immediately without recompilation. When running compiled binary elsewhere, embedded templates should work.

## Deliverables

1) Complete repository with source code, tests, OpenAPI spec, Makefile, Dockerfile, Tailwind config, GitHub Actions workflow, `.golangci.yml`, `.env.example`, README, LICENSE.  
2) Clear instructions in README to install, configure, run CLI and API, and build Tailwind assets.  
3) Justify the chosen **app name** and any architectural decisions in the README.
4) **User-friendly experience**: Root path `/` redirects to main UI, clear error messages, and comprehensive endpoint documentation.

> If any requirement is ambiguous or missing, make senior-level decisions and document them.
