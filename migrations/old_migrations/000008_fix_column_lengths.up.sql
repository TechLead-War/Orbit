-- Increase status column length in file_uploads
ALTER TABLE file_uploads ALTER COLUMN status TYPE VARCHAR(50);

-- Add student_id column to students table if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name='students' AND column_name='student_id') THEN
        ALTER TABLE students ADD COLUMN student_id VARCHAR(50) UNIQUE;
    END IF;
END $$; 