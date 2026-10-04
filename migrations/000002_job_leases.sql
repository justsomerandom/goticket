-- A worker owns a claimed job only for the duration of its processing lease.
-- This lets another worker recover work after a process crashes mid-job.
ALTER TABLE jobs ADD COLUMN lock_token uuid;
CREATE INDEX jobs_running_locked_idx ON jobs(locked_at) WHERE state = 'running';
