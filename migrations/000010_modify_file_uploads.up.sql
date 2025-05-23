-- Drop the existing foreign key constraint
ALTER TABLE file_uploads DROP CONSTRAINT IF EXISTS file_uploads_uploaded_by_fkey;

-- Add the new foreign key constraint with ON DELETE SET NULL
ALTER TABLE file_uploads 
    ALTER COLUMN uploaded_by DROP NOT NULL,
    ADD CONSTRAINT file_uploads_uploaded_by_fkey 
    FOREIGN KEY (uploaded_by) 
    REFERENCES students(id) 
    ON DELETE SET NULL; 