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