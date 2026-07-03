# Repository Guidelines

## Project Structure & Module Organization

This repository contains the Xiaozhi ESP32 voice-assistant server. The main Go service starts in `cmd/server/`; MQTT-only tooling is in `cmd/mqtt/`. Core packages live under `internal/`, with shared helpers in `constants/`, `logger/`, and `lib/`. Configuration and runtime assets are in `config/` and `build/`. Docs and design notes are in `doc/` and `ai_doc/`; protocol clients and integration checks live under `test/`.

The management console has two modules: `manager/backend` is a Go API service, and `manager/frontend` is a Vue 3/Vite app.

## Build, Test, and Development Commands

- `make setup`: install dependencies and init submodules.
- `make build`: build the main server binary.
- `make run`: run the main server locally.
- `make dev`: start the server, manager backend, and frontend.
- `go test ./...`: run tests for the root Go module.
- `cd manager/backend && go test ./...`: run manager backend tests.
- `cd manager/frontend && npm ci`: install frontend dependencies.
- `cd manager/frontend && npm run dev`: start the Vite dev server.
- `cd manager/frontend && npm run build`: build the frontend.
- `cd docker/docker-composer && docker compose up -d`: run Compose.

## Coding Style & Naming Conventions

Format Go code with `gofmt` or `go fmt ./...`. Use short lowercase package names, PascalCase for exported identifiers, camelCase for private identifiers, and uppercase acronyms such as `HTTP`, `MQTT`, `ASR`, and `TTS`. Go tests use the `*_test.go` suffix. Vue components use PascalCase filenames; composables follow `useXxx.js`.

## Testing Guidelines

Use Go’s standard `testing` package. Keep unit tests beside the code they cover, especially for config parsing, protocols, queues, sessions, storage, and concurrency-sensitive logic. Integration tests in `test/` may require live services, model files, Redis, MQTT, or external AI endpoints; document prerequisites nearby. For frontend changes, run `npm run build` and manually verify affected console flows.

## Commit & Pull Request Guidelines

Recent history uses Conventional Commit-style subjects such as `fix:`, `feat:`, `build:`, and `docs(architecture):`. Keep subjects concise and imperative. Pull requests should explain what changed, why, validation commands, linked issues or design docs, and compatibility impacts for configuration, MQTT/WebSocket, OTA, or deployment. Include screenshots or recordings for visible UI changes.

## Security & Configuration Tips

Do not commit real API keys, tokens, passwords, private endpoints, or generated agent scratch files. When adding configuration, update the relevant structs, defaults in `config/`, and documentation in `doc/`.
