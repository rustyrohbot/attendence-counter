-- name: ListAttendance :many
SELECT id, date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count
FROM attendance
ORDER BY date DESC;

-- name: GetAttendance :one
SELECT id, date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count
FROM attendance
WHERE id = ?;

-- name: CreateAttendance :one
INSERT INTO attendance (date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count;

-- name: UpdateAttendance :one
UPDATE attendance
SET date = ?, work_location = ?, work_city = ?, day = ?, year_week = ?, notes = ?, blockers = ?, in_flight = ?, rolling_in_office_count = ?
WHERE id = ?
RETURNING id, date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count;

-- name: DeleteAttendance :exec
DELETE FROM attendance WHERE id = ?;

-- name: GetAttendanceByDate :one
SELECT id, date, work_location, work_city, day, year_week, notes, blockers, in_flight, rolling_in_office_count
FROM attendance
WHERE date = ?;

-- name: GetRollingInOfficeCount :one
SELECT COUNT(*) as count
FROM attendance
WHERE work_location = 'OFFICE'
  AND date >= date(?, '-91 days')  -- 13 weeks * 7 days
  AND date < ?;