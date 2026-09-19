# Events

Ticket audit events are immutable rows: `created`, `assigned`, `status_changed`, `priority_changed`, and `comment_added`. They contain the ticket, optional actor, event type, JSON data, and creation timestamp.

`ticket.created` and `ticket.comment_added` enqueue durable jobs. The current worker delivers them to structured logs; replacing the sink does not alter queue semantics.
