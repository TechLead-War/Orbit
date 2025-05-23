CREATE TABLE students (
    id           SERIAL PRIMARY KEY,
    student_id   VARCHAR(50) UNIQUE NOT NULL,
    name         VARCHAR(50) NOT NULL,
    email        VARCHAR(50) UNIQUE NOT NULL,
    leetcode_id  VARCHAR(50) UNIQUE NOT NULL,
    passing_year INT NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
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

CREATE TABLE contest_histories (
    id                  BIGSERIAL PRIMARY KEY,
    student_id          BIGINT,
    contest_title       TEXT,
    rating              NUMERIC,
    ranking             BIGINT,
    problems_solved     BIGINT,
    finish_time_seconds BIGINT,
    contest_date        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ
);

CREATE TABLE file_uploads (
    id                SERIAL PRIMARY KEY,
    file_name         VARCHAR(255) NOT NULL,
    original_name     VARCHAR(255) NOT NULL,
    file_type         VARCHAR(50) NOT NULL,
    file_size         BIGINT NOT NULL,
    storage_path      TEXT NOT NULL,
    total_records     INT NOT NULL,
    successful_records INT NOT NULL,
    failed_records    INT NOT NULL,
    status            VARCHAR(50) NOT NULL,
    error_details     TEXT,
    uploaded_by       INT REFERENCES students(id),
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);