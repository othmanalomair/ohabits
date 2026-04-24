-- Add client_upload_id to daily_images for idempotent uploads.
-- Prevents duplicate rows when the native client retries an upload
-- (e.g. network blip, app killed mid-upload). The client sends its local
-- UUID; server uses it as a natural dedup key per user.

ALTER TABLE daily_images
    ADD COLUMN IF NOT EXISTS client_upload_id UUID;

-- Allow NULL for legacy rows; new uploads will always send a value.
CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_images_user_client_upload_id
    ON daily_images (user_id, client_upload_id)
    WHERE client_upload_id IS NOT NULL;
