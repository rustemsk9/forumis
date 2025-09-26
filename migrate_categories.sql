-- Migration to add category1 and category2 columns to threads table
-- Run this if you have existing data in your database

ALTER TABLE threads ADD COLUMN category1 VARCHAR(255) DEFAULT '';
ALTER TABLE threads ADD COLUMN category2 VARCHAR(255) DEFAULT '';

-- Update existing threads to have empty categories (optional)
-- UPDATE threads SET category1 = '', category2 = '' WHERE category1 IS NULL OR category2 IS NULL;