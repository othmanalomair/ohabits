-- Add unique constraint on mood_ratings for (user_id, date)
-- This allows using ON CONFLICT for upsert operations

ALTER TABLE mood_ratings
ADD CONSTRAINT mood_ratings_user_date_unique UNIQUE (user_id, date);
