CREATE TABLE IF NOT EXISTS weekly_stats (
    id SERIAL PRIMARY KEY,
    student_id VARCHAR(50) NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    week_start_date TIMESTAMPTZ NOT NULL,
    week_end_date TIMESTAMPTZ NOT NULL,
    easy_count INT NOT NULL DEFAULT 0,
    medium_count INT NOT NULL DEFAULT 0,
    hard_count INT NOT NULL DEFAULT 0,
    total_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, week_start_date)
); 