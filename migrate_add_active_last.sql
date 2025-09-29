-- Migration to add active_last column to sessions table
-- Run this if you have an existing database
ALTER TABLE sessions ADD COLUMN active_last integer default 0;

-- Migration to add category1 and category2 columns to threads table
-- Run this if you have existing data in your database

ALTER TABLE threads ADD COLUMN category1 VARCHAR(255) DEFAULT '';
ALTER TABLE threads ADD COLUMN category2 VARCHAR(255) DEFAULT '';

-- Update existing threads to have empty categories (optional)
-- UPDATE threads SET category1 = '', category2 = '' WHERE category1 IS NULL OR category2 IS NULL;

-- mydb.db-shm is a temporary file created by SQLite when using Write-Ahead Logging (WAL) mode. 
-- It stands for "shared memory" and is used to coordinate access to the database among multiple connections. 
--It is safe to ignore or delete this file when the database is not in use.