CREATE TABLE students (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(50) NOT NULL,
    email          VARCHAR(50) UNIQUE NOT NULL,
    profile_link   TEXT,
    current_rating INT,
    passing_year  INT,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE weekly_progress (
    id                SERIAL PRIMARY KEY,
    student_id        INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    week_start        DATE NOT NULL,
    questions_solved  INT DEFAULT 0,
    submissions_count INT DEFAULT 0,
    rating_delta      INT DEFAULT 0,
    suspected_cheat   BOOLEAN DEFAULT FALSE,
    UNIQUE(student_id, week_start)
);

CREATE TABLE contest_participation (
    id            SERIAL PRIMARY KEY,
    student_id    INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    contest_name  VARCHAR(100),
    contest_date  DATE NOT NULL,
    rating_before INT,
    rating_after  INT,
    rating_change INT
);