-- Daily images table for storing photos with notes
CREATE TABLE daily_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    original_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for quick lookup by user and date
CREATE INDEX idx_daily_images_user_date ON daily_images(user_id, date);
