-- Drop the foreign key constraint
ALTER TABLE file_uploads DROP CONSTRAINT IF EXISTS file_uploads_uploaded_by_fkey;

-- Add back the original foreign key constraint
ALTER TABLE file_uploads 
    ALTER COLUMN uploaded_by SET NOT NULL,
    ADD CONSTRAINT file_uploads_uploaded_by_fkey 
    FOREIGN KEY (uploaded_by) 
    REFERENCES students(id); 