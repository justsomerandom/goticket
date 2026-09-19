# GoTicket

GoTicket is planned as a production-oriented backend for ticket and support workflows, including multiple ingestion methods, asynchronous notifications, automation, SLA handling, outbound webhooks, and auditability.

This repository provides a deployable first ticketing backend: a REST API, PostgreSQL persistence, an audit trail, and a durable PostgreSQL-backed notification worker.

## Goals

- Build a maintainable support-ticket backend around clear domain packages.
- Demonstrate idiomatic Go API, worker, persistence, observability, and background-processing patterns.
- Explore reliable webhook ingestion and notification delivery with idempotency, retries, and audit history.
- Keep the first version focused on backend correctness rather than UI breadth.

## Planned Features

- Ticket CRUD.
- Multiple inbound ticket sources.
- Generic webhook ingestion.
- Email-provider webhook ingestion later.
- Internal and public comments.
- Assignment, tags, priority, and custom fields.
- SLA tracking.
- Notification queues.
- Outbound webhooks with signing.
- Retry policies, exponential backoff, and dead-letter handling.
- Idempotency for inbound and outbound workflows.
- Ticket merging.
- Audit history.
- Automation and routing rules.

## Architecture

The intended architecture has an API process for synchronous request handling and a worker process for asynchronous jobs. Domain packages own ticketing, ingestion, notification, SLA, automation, and webhook behavior. Storage and observability packages provide shared infrastructure without becoming global business-logic layers.

```mermaid
flowchart LR
    Client[API Clients] --> Api[cmd/api]
    Provider[Inbound Webhooks] --> Api
    Api --> Storage[(PostgreSQL)]
    Api --> Queue[(Redis or job queue - planned)]
    Worker[cmd/worker] --> Queue
    Worker --> Storage
    Worker --> Webhooks[Outbound Webhooks]
    Worker --> Notifications[Notifications]
```

## Repository Structure

- `cmd/api` - planned HTTP API entry point.
- `cmd/worker` - planned background worker entry point.
- `internal/ticket` - ticket domain behavior.
- `internal/ingest` - inbound source handling.
- `internal/webhook` - webhook signing, dispatch, and receipt concepts.
- `internal/notification` - planned notification queue and delivery behavior.
- `internal/automation` - planned routing and automation rules.
- `internal/sla` - SLA policy and tracking behavior.
- `internal/storage` - database access boundaries.
- `internal/observability` - logging, tracing, and metrics setup.
- `api/openapi` - future API specifications.
- `docs` - architecture and event documentation.

## Technology Stack

- Go - primary language.
- Chi or another lightweight router - planned.
- pgx - planned PostgreSQL access.
- PostgreSQL - planned persistence layer.
- Redis - planned only where it clearly supports queues or coordination.
- OpenTelemetry - planned observability.
- REST API - planned public interface.

The API uses Chi and pgx/pgxpool. PostgreSQL is also the job queue, so Redis is not required for this version.

## Development

Set `DATABASE_URL` and apply the migration before starting either process:

```powershell
psql $env:DATABASE_URL -f migrations/000001_initial.sql
go run ./cmd/api
go run ./cmd/worker
```

`HTTP_ADDR` defaults to `:8080`. The API exposes `GET /healthz` and `GET /readyz`.

Ticket endpoints are under `/v1`: create organizations and users, create/list tickets, fetch a ticket, update assignment/status/priority, create/list comments, and list ticket audit events. All JSON errors include an error code, message, and request ID.

## Roadmap

- [ ] Define ticket, user, organization, and comment domain models.
- [ ] Establish HTTP API routing and validation.
- [ ] Add PostgreSQL migrations and storage interfaces.
- [ ] Implement inbound webhook ingestion with idempotency.
- [ ] Add worker-driven notification delivery.
- [ ] Add outbound webhook signing and retry handling.
- [ ] Add SLA tracking and automation rules.
- [ ] Publish OpenAPI documentation.

## License

MIT. See `LICENSE`.

GoTicket is under active development.
