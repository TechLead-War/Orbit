CREATE TABLE students (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(50) NOT NULL,
    email      VARCHAR(50) UNIQUE NOT NULL,
    leetcode_id VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ratings (
    id             SERIAL PRIMARY KEY,
    student_id     INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    rating         INT NOT NULL,
    problems_count INT NOT NULL,
    easy_count     INT NOT NULL,
    medium_count   INT NOT NULL,
    hard_count     INT NOT NULL,
    global_rank    INT NOT NULL,
    recorded_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, recorded_at)
);

CREATE TABLE contest_history (
    id                  SERIAL PRIMARY KEY,
    student_id         INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    contest_title      VARCHAR(100) NOT NULL,
    rating             FLOAT NOT NULL,
    ranking           INT NOT NULL,
    problems_solved    INT NOT NULL,
    finish_time_seconds INT,
    contest_date       TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(student_id, contest_title)
);