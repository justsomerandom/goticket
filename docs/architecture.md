# Architecture

This document will describe GoTicket's API, worker, domain packages, storage boundaries, ingestion flow, and notification delivery model as implementation begins.

GoTicket is a modular monolith.

Rules:
- domain layer must not depend on HTTP or database code
- repositories expose interfaces used by services
- HTTP handlers remain thin
- PostgreSQL is the source of truth
- background work must be idempotent
- avoid global mutable state
- context.Context flows through all I/O boundaries
- errors should preserve cause and domain meaning
- prefer explicit code over unnecessary abstraction