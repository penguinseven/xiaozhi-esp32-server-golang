# Repository Guidelines

## Project Structure & Module Organization

This repository implements a voice-assistant backend for the XiaoZhi ESP32 device.  The top-level Go module at `go.mod` is the primary server; a separate Go module under `asr_server/` (a git submodule) provides sherpa-onnx-based speech recognition via CGo.

| Path | Purpose |
|------|---------|
| `cmd/server/` | Main server entry point (WebSocket, HTTP, MQTT) |
| `cmd/mqtt/` | Standalone MQTT server |
| `internal/` | Core packages: app lifecycle, config, domain logic, DB, pool, utilities |
| `manager/backend/` | Admin console API (separate Go module) |
| `manager/frontend/` | Admin console UI (Vue 3 + Vite + Element Plus) |
| `asr_server/` | ASR submodule (sherpa-onnx + onnxruntime CGo bindings) |
| `test/` | Integration tests and protocol automation suites |
| `doc/` | Architecture, deployment, and configuration documentation |
| `docker/` | Multi-platform Dockerfiles and Compose files |
| `build/` | Platform-specific runtime assets (DLLs, configs, scripts) |
| `config/` | Default configuration YAML and JSON files |

## Build, Test, and Development Commands

**Go backend** — run from the repository root:

```sh
# Build the main server binary
go build -o xiaozhi-server ./cmd/server/

# Run tests for a specific package
go test ./internal/util/...

# Run all tests
go test ./...

# Run with verbose output
go test -v -count=1 ./...
```

**Admin frontend** — run from `manager/frontend/`:

```sh
npm run dev      # Start Vite dev server
npm run build    # Production build (copies diagnose.js into dist/)
```

**Docker builds** — tagged releases are built via `docker/Dockerfile.windows`, `docker/Dockerfile.linux`, and `docker/Dockerfile.main`.  See `.github/workflows/build-release.yml` for the CI pipeline.

## Coding Style & Naming Conventions

This is a Go project.  Follow standard Go idioms (`go fmt`):

- **Formatting**: Use `gofmt` (or `go fmt ./...`) before committing.  No external linter config is checked in; `go vet ./...` is the minimum due-diligence step.
- **Naming**: Exported identifiers use PascalCase; unexported use camelCase.  Acronyms stay uppercase (`WS`, `HTTP`, `MQTT`).
- **Error handling**: Errors are propagated explicitly.  Use `fmt.Errorf("context: %w", err)` to wrap with context.  Avoid `_` for ignored errors unless the intent is obvious.
- **Logging**: Use the project's structured logger (`logger/`), built on `logrus`.  Avoid `fmt.Println` in production code.
- **Config**: Configuration is loaded from `config/config.yaml` and merged with environment variables.  Do not hardcode secrets.

## Testing Guidelines

- **Framework**: Standard `testing` package.  No third-party test frameworks are used.
- **Test files**: Co-located with source code as `*_test.go`.  Integration tests live under `test/` and may depend on running services (Redis, MQTT brokers, ASR models).
- **Test naming**: `TestFunctionName_Scenario` or `TestFunctionName` (standard Go conventions).
- **Coverage**: No strict coverage threshold is enforced, but critical domain logic (audio utilities, queues, sessions) should have unit tests.  See `internal/util/` for reference examples.
- **Running integration tests**: Some tests require a live Redis instance or external AI endpoints.  Check individual test files for build tags or environment preconditions.

## Commit & Pull Request Guidelines

Commits follow a lightweight [Conventional Commits](https://www.conventionalcommits.org/) style:

- `feat:` — new feature
- `fix:` — bug fix
- `test:` — adding or updating tests
- `refactor:` — code restructuring with no behavioral change
- `doc:` — documentation-only changes

Both English and Chinese descriptions are used in the existing history.  PR titles mirror the branch topic.  Squash merges are preferred (the project history shows one merge commit per PR).

When opening a pull request:

- Reference the associated issue or design document in the description.
- Include a summary of what changed and why.
- If the change affects configuration, MQTT/WebSocket protocol, or OTA behavior, note the compatibility implications.

## Agent-Specific Instructions

This section is for AI coding agents (Codex, etc.) working on this repository.

- **Read before editing**: Always check `go.mod`, `.gitmodules`, and existing `doc/` files for design context before making structural changes.  The `ai_doc/` directory contains prior proposals and review findings — consult it when modifying related subsystems.
- **Submodule awareness**: `asr_server/` is a git submodule.  Changes to its CGo bindings or model paths affect both the main server and the standalone ASR server build.  Verify compatibility across both.
- **Build tags**: The main `cmd/server/` uses build tags (`asr_enabled`/`!asr_enabled`) to conditionally compile ASR support.  Ensure new ASR-related code is gated behind the correct tag.
- **Agent ephemera**: The `.agents/` directory (if present) contains task-tracking context used by automated agents.  Do not commit agent ephemera to the main branch without cleaning.
- **Config drift**: The server reads config from `config/config.yaml`.  If you add a new config field, update the default file and the corresponding struct in `internal/config/`.

- **onnxruntime dylib (macOS)**: The ASR subsystem requires `libonnxruntime.1.21.0.dylib` (and a symlink) at `lib/ten-vad/lib/macOS/`. These are not tracked in Git due to size (~64MB).  If running the ASR server on macOS, download from the [ONNX Runtime releases](https://github.com/microsoft/onnxruntime/releases/tag/v1.21.0) page and place both files in that directory.
