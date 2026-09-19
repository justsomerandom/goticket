# Architecture

The API process (`cmd/api`) owns HTTP transport and calls `internal/application`. Domain types live in `internal/ticket`, `internal/organization`, and `internal/user`. PostgreSQL SQL is confined to `internal/storage/postgres`.

Ticket creation and comment creation insert their audit event and durable notification job in the same PostgreSQL transaction. The worker (`cmd/worker`) claims due jobs with `FOR UPDATE SKIP LOCKED`. Failed jobs retry with exponential backoff; after `max_attempts`, their state becomes `dead`.
