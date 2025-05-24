-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Core Tables
CREATE TABLE students (
    id                BIGSERIAL PRIMARY KEY,
    student_id        VARCHAR(50) UNIQUE NOT NULL,  -- College ID
    name             VARCHAR(100) NOT NULL,
    email            VARCHAR(100) UNIQUE NOT NULL,
    leetcode_id      VARCHAR(50) UNIQUE NOT NULL,
    passing_year     INT NOT NULL,
    batch           VARCHAR(20) NOT NULL,           -- For batch-wise analytics
    department      VARCHAR(50) NOT NULL,           -- For department-wise tracking
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- Latest snapshot of student's LeetCode stats
CREATE TABLE leetcode_stats (
    id              BIGSERIAL PRIMARY KEY,
    student_id      BIGINT REFERENCES students(id) ON DELETE CASCADE,
    rating          INT NOT NULL DEFAULT 0,
    global_ranking  INT NOT NULL DEFAULT 0,
    total_solved    INT NOT NULL DEFAULT 0,
    easy_solved     INT NOT NULL DEFAULT 0,
    medium_solved   INT NOT NULL DEFAULT 0,
    hard_solved     INT NOT NULL DEFAULT 0,
    contests_participated INT NOT NULL DEFAULT 0,
    contest_rating  FLOAT NOT NULL DEFAULT 0,
    contest_global_ranking INT NOT NULL DEFAULT 0,
    reputation      INT NOT NULL DEFAULT 0,
    contribution_points INT NOT NULL DEFAULT 0,
    last_solved_at  TIMESTAMPTZ,
    initial_rating  INT NOT NULL DEFAULT 0,
    problems_solved_last_week INT NOT NULL DEFAULT 0,
    acceptance_rate DECIMAL(5,2),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id)  -- One latest record per student
);

-- Historical record of daily progress (partitioned by month)
CREATE TABLE daily_progress (
    id              BIGSERIAL PRIMARY KEY,
    student_id      BIGINT REFERENCES students(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    problems_solved INT NOT NULL DEFAULT 0,
    easy_solved     INT NOT NULL DEFAULT 0,
    medium_solved   INT NOT NULL DEFAULT 0,
    hard_solved     INT NOT NULL DEFAULT 0,
    time_spent      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, date)
) PARTITION BY RANGE (date);

-- Create partitions for the current year by month
CREATE TABLE daily_progress_y2024m01 PARTITION OF daily_progress
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE daily_progress_y2024m02 PARTITION OF daily_progress
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
-- Add more partitions as needed

-- Contest participation history (partitioned by year)
CREATE TABLE contest_history (
    id                  BIGSERIAL PRIMARY KEY,
    student_id          BIGINT REFERENCES students(id) ON DELETE CASCADE,
    contest_name        VARCHAR(100) NOT NULL,
    contest_date        TIMESTAMPTZ NOT NULL,
    rating             INT NOT NULL DEFAULT 0,
    ranking            INT NOT NULL DEFAULT 0,
    problems_attempted INT NOT NULL DEFAULT 0,
    problems_solved    INT NOT NULL DEFAULT 0,
    finish_time        BIGINT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (contest_date);

-- Create partitions for contest history
CREATE TABLE contest_history_2024 PARTITION OF contest_history
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
-- Add more partitions as needed

-- Weekly performance tracking
CREATE TABLE weekly_stats (
    id                  BIGSERIAL PRIMARY KEY,
    student_id          BIGINT REFERENCES students(id) ON DELETE CASCADE,
    week_start         DATE NOT NULL,
    week_end           DATE NOT NULL,
    problems_solved     INT NOT NULL DEFAULT 0,
    easy_solved         INT NOT NULL DEFAULT 0,
    medium_solved       INT NOT NULL DEFAULT 0,
    hard_solved         INT NOT NULL DEFAULT 0,
    time_spent          INT NOT NULL DEFAULT 0,
    contests_attended   INT NOT NULL DEFAULT 0,
    average_rating      FLOAT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, week_start)  -- One record per student per week
);

-- Achievements system
CREATE TABLE achievements (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL,
    criteria        JSONB NOT NULL,  -- Flexible criteria storage
    badge_image_url VARCHAR(255),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE student_achievements (
    id              BIGSERIAL PRIMARY KEY,
    student_id      BIGINT REFERENCES students(id) ON DELETE CASCADE,
    achievement_id  BIGINT REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, achievement_id)  -- Each achievement once per student
);

-- Create file_uploads table
CREATE TABLE file_uploads (
    id BIGSERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create batch_stats table
CREATE TABLE batch_stats (
    id BIGSERIAL PRIMARY KEY,
    batch VARCHAR(20) UNIQUE NOT NULL,
    total_students INT NOT NULL DEFAULT 0,
    active_students INT NOT NULL DEFAULT 0,
    average_rating FLOAT NOT NULL DEFAULT 0,
    highest_rating INT NOT NULL DEFAULT 0,
    total_problems_solved INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create department_stats table
CREATE TABLE department_stats (
    id BIGSERIAL PRIMARY KEY,
    department VARCHAR(50) UNIQUE NOT NULL,
    total_students INT NOT NULL DEFAULT 0,
    active_students INT NOT NULL DEFAULT 0,
    average_rating FLOAT NOT NULL DEFAULT 0,
    highest_rating INT NOT NULL DEFAULT 0,
    total_problems_solved INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create system_stats table
CREATE TABLE system_stats (
    id BIGSERIAL PRIMARY KEY,
    total_students INT NOT NULL DEFAULT 0,
    active_students INT NOT NULL DEFAULT 0,
    total_problems_solved INT NOT NULL DEFAULT 0,
    total_contests INT NOT NULL DEFAULT 0,
    average_student_rating FLOAT NOT NULL DEFAULT 0,
    highest_student_rating INT NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_leetcode_stats_rating ON leetcode_stats(rating);
CREATE INDEX idx_leetcode_stats_total_solved ON leetcode_stats(total_solved);
CREATE INDEX idx_daily_progress_date ON daily_progress(date);
CREATE INDEX idx_contest_history_contest_date ON contest_history(contest_date);
CREATE INDEX idx_contest_history_rating ON contest_history(rating);
CREATE INDEX idx_weekly_stats_week_start ON weekly_stats(week_start);
CREATE INDEX idx_students_batch ON students(batch);
CREATE INDEX idx_students_department ON students(department);
CREATE INDEX idx_students_passing_year ON students(passing_year);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_students_updated_at
    BEFORE UPDATE ON students
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_leetcode_stats_updated_at
    BEFORE UPDATE ON leetcode_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contest_history_updated_at
    BEFORE UPDATE ON contest_history
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_daily_progress_updated_at
    BEFORE UPDATE ON daily_progress
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_weekly_stats_updated_at
    BEFORE UPDATE ON weekly_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_uploads_updated_at
    BEFORE UPDATE ON file_uploads
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_batch_stats_updated_at
    BEFORE UPDATE ON batch_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_department_stats_updated_at
    BEFORE UPDATE ON department_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_stats_updated_at
    BEFORE UPDATE ON system_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 