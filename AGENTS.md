# AGENTS.md

## Project Purpose

GoTicket is intended to become a backend-focused ticket/support system with reliable ingestion, automation, SLA tracking, notifications, webhooks, and auditability.

## Engineering Priorities

- Correct domain behavior around ticket state, comments, SLA rules, and audit history.
- Maintainable idiomatic Go package design.
- Reliable idempotent processing for webhooks and background work.
- Security for public ingestion endpoints, webhook signing, and authorization.
- Observability for API requests, worker jobs, retries, and failures.
- Testability without requiring external providers.

## Architecture Rules

- Group packages by domain responsibility instead of global controller/service/repository layers.
- Keep API transport concerns in `cmd/api` and package-level HTTP handlers when introduced.
- Keep worker orchestration in `cmd/worker` and domain packages.
- Keep persistence details behind storage boundaries.
- Keep webhook ingestion separate from outbound webhook dispatch.
- Avoid package cycles and hidden global state.

## Coding Guidelines

- Use idiomatic Go with small interfaces at package boundaries.
- Prefer explicit errors and clear context propagation.
- Do not introduce unnecessary abstractions.
- Do not silently change architecture or package ownership.
- Do not add technologies merely for resume value.
- Prefer well-maintained libraries.
- Preserve backwards compatibility once public APIs exist.
- Validate external input from APIs, webhooks, and configuration.
- Keep secrets out of the repository.
- Avoid generated code unless justified and documented.
- Add tests with meaningful behavior changes.
- Document non-obvious design decisions.

## Testing

Future tests should include domain unit tests, HTTP handler tests, storage integration tests, worker retry tests, idempotency tests, and contract tests for webhook signatures and event payloads.

## Documentation

Update `README.md`, `docs/`, and future OpenAPI files when architecture, event contracts, or user-visible behavior changes.

## Agent Workflow

Before making significant changes:

1. Inspect the existing architecture.
2. Understand relevant domain code.
3. Make the smallest coherent change.
4. Run relevant formatting, linting, and tests.
5. Summarize architectural consequences.

`AGENTS.md` may be expanded as this project matures.
