-- Drop triggers
DROP TRIGGER IF EXISTS update_system_stats_updated_at ON system_stats;
DROP TRIGGER IF EXISTS update_department_stats_updated_at ON department_stats;
DROP TRIGGER IF EXISTS update_batch_stats_updated_at ON batch_stats;
DROP TRIGGER IF EXISTS update_file_uploads_updated_at ON file_uploads;
DROP TRIGGER IF EXISTS update_weekly_stats_updated_at ON weekly_stats;
DROP TRIGGER IF EXISTS update_daily_progress_updated_at ON daily_progress;
DROP TRIGGER IF EXISTS update_contest_history_updated_at ON contest_history;
DROP TRIGGER IF EXISTS update_leetcode_stats_updated_at ON leetcode_stats;
DROP TRIGGER IF EXISTS update_students_updated_at ON students;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_students_batch;
DROP INDEX IF EXISTS idx_students_department;
DROP INDEX IF EXISTS idx_students_passing_year;
DROP INDEX IF EXISTS idx_leetcode_stats_rating;
DROP INDEX IF EXISTS idx_leetcode_stats_total_solved;
DROP INDEX IF EXISTS idx_contest_history_contest_date;
DROP INDEX IF EXISTS idx_contest_history_rating;
DROP INDEX IF EXISTS idx_daily_progress_date;
DROP INDEX IF EXISTS idx_weekly_stats_week_start;

-- Drop tables in reverse order of creation (to handle foreign key constraints)
DROP TABLE IF EXISTS system_stats;
DROP TABLE IF EXISTS department_stats;
DROP TABLE IF EXISTS batch_stats;
DROP TABLE IF EXISTS file_uploads;
DROP TABLE IF EXISTS weekly_stats;
DROP TABLE IF EXISTS daily_progress;
DROP TABLE IF EXISTS contest_history;
DROP TABLE IF EXISTS leetcode_stats;
DROP TABLE IF EXISTS students;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp"; 