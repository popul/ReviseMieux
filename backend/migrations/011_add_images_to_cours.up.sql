-- Add images column to store saved image filenames for a course
ALTER TABLE cours ADD COLUMN images JSONB DEFAULT '[]'::jsonb;
