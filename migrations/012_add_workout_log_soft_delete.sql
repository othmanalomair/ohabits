-- Add soft-delete flag to workout_logs. Previously the sync/push handler silently
-- ignored the client's is_deleted=true for workout logs (it only accepted upserts),
-- so deletions from the native app were never honored and the log reappeared on
-- every pull. Adding the column + honoring it in the handler makes 'change workout'
-- durable across syncs.

ALTER TABLE workout_logs
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
