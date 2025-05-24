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
    total_solved    INT NOT NULL DEFAULT 0,
    easy_solved     INT NOT NULL DEFAULT 0,
    medium_solved   INT NOT NULL DEFAULT 0,
    hard_solved     INT NOT NULL DEFAULT 0,
    acceptance_rate DECIMAL(5,2),
    contribution_points INT DEFAULT 0,
    reputation      INT DEFAULT 0,
    global_ranking  INT,
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id)  -- One latest record per student
);

-- Historical record of daily progress
CREATE TABLE daily_progress (
    id              BIGSERIAL PRIMARY KEY,
    student_id      BIGINT REFERENCES students(id) ON DELETE CASCADE,
    problems_solved INT NOT NULL DEFAULT 0,
    easy_solved     INT NOT NULL DEFAULT 0,
    medium_solved   INT NOT NULL DEFAULT 0,
    hard_solved     INT NOT NULL DEFAULT 0,
    date           DATE NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, date)  -- One record per student per day
);

-- Contest participation history
CREATE TABLE contest_history (
    id                  BIGSERIAL PRIMARY KEY,
    student_id          BIGINT REFERENCES students(id) ON DELETE CASCADE,
    contest_name        VARCHAR(200) NOT NULL,
    contest_date        TIMESTAMPTZ NOT NULL,
    rating             INT NOT NULL,
    ranking            INT NOT NULL,
    problems_solved    INT NOT NULL DEFAULT 0,
    finish_time        INT,  -- in seconds
    rating_change      INT,  -- can be positive or negative
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, contest_name)  -- One record per student per contest
);

-- Weekly performance tracking
CREATE TABLE weekly_stats (
    id                  BIGSERIAL PRIMARY KEY,
    student_id          BIGINT REFERENCES students(id) ON DELETE CASCADE,
    week_start_date     DATE NOT NULL,
    week_end_date       DATE NOT NULL,
    problems_solved     INT NOT NULL DEFAULT 0,
    easy_solved         INT NOT NULL DEFAULT 0,
    medium_solved       INT NOT NULL DEFAULT 0,
    hard_solved         INT NOT NULL DEFAULT 0,
    contests_participated INT NOT NULL DEFAULT 0,
    avg_contest_rating   DECIMAL(10,2),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, week_start_date)  -- One record per student per week
);

-- Badges and achievements
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

-- Indexes for performance
CREATE INDEX idx_students_batch ON students(batch);
CREATE INDEX idx_students_department ON students(department);
CREATE INDEX idx_daily_progress_date ON daily_progress(date);
CREATE INDEX idx_contest_history_date ON contest_history(contest_date);
CREATE INDEX idx_weekly_stats_date ON weekly_stats(week_start_date);

-- Partitioning for historical tables (optional, for very large datasets)
-- Partition daily_progress and contest_history by date ranges
