-- +goose Up
-- Create attendance table
CREATE TABLE attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL UNIQUE,
    work_location TEXT,
    work_city TEXT,
    day TEXT,
    year_week INTEGER,
    notes TEXT,
    blockers TEXT,
    in_flight TEXT,
    rolling_in_office_count INTEGER DEFAULT 0
);

-- Create index on date for efficient queries
CREATE INDEX idx_attendance_date ON attendance(date);

-- +goose Down
DROP INDEX idx_attendance_date;
DROP TABLE attendance;