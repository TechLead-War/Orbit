-- Revert status column length in file_uploads
ALTER TABLE file_uploads ALTER COLUMN status TYPE VARCHAR(20);

-- Remove student_id column from students table if it exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
              WHERE table_name='students' AND column_name='student_id') THEN
        ALTER TABLE students DROP COLUMN student_id;
    END IF;
END $$; 