-- Migration to add active_last column to sessions table
-- Run this if you have an existing database
ALTER TABLE sessions ADD COLUMN active_last integer default 0;