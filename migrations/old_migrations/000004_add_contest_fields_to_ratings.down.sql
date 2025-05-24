ALTER TABLE ratings
ALTER COLUMN contest_rating TYPE INTEGER,
DROP COLUMN contest_count,
DROP COLUMN top_percentage; 