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
    status            VARCHAR(20) NOT NULL,
    error_details     TEXT,
    uploaded_by       INT REFERENCES students(id),
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
); 