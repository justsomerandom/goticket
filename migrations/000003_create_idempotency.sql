-- Ticket and comment creation can be safely retried without duplicating their
-- audit events or notification jobs. A key is scoped to the parent resource.
ALTER TABLE tickets ADD COLUMN idempotency_key text;
ALTER TABLE tickets ADD COLUMN idempotency_hash bytea;
CREATE UNIQUE INDEX tickets_organization_idempotency_key_idx
    ON tickets(organization_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

ALTER TABLE comments ADD COLUMN idempotency_key text;
ALTER TABLE comments ADD COLUMN idempotency_hash bytea;
CREATE UNIQUE INDEX comments_ticket_idempotency_key_idx
    ON comments(ticket_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
