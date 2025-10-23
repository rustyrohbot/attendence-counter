package handlers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"attendance-tracker/internal/db"
	"attendance-tracker/internal/util"
)

type AttendanceHandler struct {
	queries *db.Queries
	db      *sql.DB
}

func NewAttendanceHandler(queries *db.Queries, db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{
		queries: queries,
		db:      db,
	}
}

func (h *AttendanceHandler) List(w http.ResponseWriter, r *http.Request) {
	attendance, err := h.queries.ListAttendance(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch attendance", http.StatusInternalServerError)
		return
	}

	// Format dates for display
	for i := range attendance {
		attendance[i].Date = util.FormatDate(attendance[i].Date)
	}

	// Check for import results
	imported := r.URL.Query().Get("imported")
	skipped := r.URL.Query().Get("skipped")
	updated := r.URL.Query().Get("updated")

	data := struct {
		Attendance []db.Attendance
		Imported   string
		Skipped    string
		Updated    string
	}{
		Attendance: attendance,
		Imported:   imported,
		Skipped:    skipped,
		Updated:    updated,
	}

	tmpl := template.Must(template.ParseFiles("internal/templates/layout.html", "internal/templates/attendance/list.html"))
	tmpl.Execute(w, data)
}

func (h *AttendanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	date := util.ParseDate(r.FormValue("date"))
	workLocation := strings.TrimSpace(r.FormValue("work_location"))
	workCity := strings.TrimSpace(r.FormValue("work_city"))
	day := strings.TrimSpace(r.FormValue("day"))
	yearWeekStr := r.FormValue("year_week")
	notes := strings.TrimSpace(r.FormValue("notes"))
	blockers := strings.TrimSpace(r.FormValue("blockers"))
	inFlight := strings.TrimSpace(r.FormValue("in_flight"))

	yearWeek, _ := strconv.Atoi(yearWeekStr)

	// Calculate rolling count
	rollingCount, err := h.calculateRollingInOfficeCount(r.Context(), date)
	if err != nil {
		http.Error(w, "Failed to calculate rolling count", http.StatusInternalServerError)
		return
	}

	createdAttendance, err := h.queries.CreateAttendance(r.Context(), db.CreateAttendanceParams{
		Date:                 date,
		WorkLocation:         sql.NullString{String: workLocation, Valid: workLocation != ""},
		WorkCity:             sql.NullString{String: workCity, Valid: workCity != ""},
		Day:                  sql.NullString{String: day, Valid: day != ""},
		YearWeek:             sql.NullInt64{Int64: int64(yearWeek), Valid: yearWeek > 0},
		Notes:                sql.NullString{String: notes, Valid: notes != ""},
		Blockers:             sql.NullString{String: blockers, Valid: blockers != ""},
		InFlight:             sql.NullString{String: inFlight, Valid: inFlight != ""},
		RollingInOfficeCount: sql.NullInt64{Int64: int64(rollingCount), Valid: true},
	})
	if err != nil {
		http.Error(w, "Failed to create attendance", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Format date for display
		createdAttendance.Date = util.FormatDate(createdAttendance.Date)
		// Return the new row
		data := struct {
			Attendance db.Attendance
		}{
			Attendance: createdAttendance,
		}
		w.WriteHeader(http.StatusCreated)
		tmpl := template.Must(template.ParseFiles("internal/templates/attendance/row.html"))
		tmpl.Execute(w, data)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AttendanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/attendance/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.queries.DeleteAttendance(r.Context(), int64(id))
	if err != nil {
		http.Error(w, "Failed to delete attendance", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AttendanceHandler) Edit(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/attendance/")
	idStr = strings.TrimSuffix(idStr, "/edit")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	attendance, err := h.queries.GetAttendance(r.Context(), int64(id))
	if err != nil {
		http.Error(w, "Attendance record not found", http.StatusNotFound)
		return
	}

	// Check if this is an HTMX request for inline editing
	if r.Header.Get("HX-Request") == "true" {
		data := struct {
			Attendance db.Attendance
		}{
			Attendance: attendance,
		}
		tmpl := template.Must(template.ParseFiles("internal/templates/attendance/edit_inline.html"))
		tmpl.Execute(w, data)
		return
	}

	// Regular page request
	data := struct {
		Attendance db.Attendance
	}{
		Attendance: attendance,
	}

	tmpl := template.Must(template.ParseFiles("internal/templates/layout.html", "internal/templates/attendance/edit.html"))
	tmpl.Execute(w, data)
}

func (h *AttendanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/attendance/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	date := util.ParseDate(r.FormValue("date"))
	workLocation := strings.TrimSpace(r.FormValue("work_location"))
	workCity := strings.TrimSpace(r.FormValue("work_city"))
	day := strings.TrimSpace(r.FormValue("day"))
	yearWeekStr := r.FormValue("year_week")
	notes := strings.TrimSpace(r.FormValue("notes"))
	blockers := strings.TrimSpace(r.FormValue("blockers"))
	inFlight := strings.TrimSpace(r.FormValue("in_flight"))

	yearWeek, _ := strconv.Atoi(yearWeekStr)

	// Calculate rolling count
	rollingCount, err := h.calculateRollingInOfficeCount(r.Context(), date)
	if err != nil {
		http.Error(w, "Failed to calculate rolling count", http.StatusInternalServerError)
		return
	}

	updatedAttendance, err := h.queries.UpdateAttendance(r.Context(), db.UpdateAttendanceParams{
		Date:                 date,
		WorkLocation:         sql.NullString{String: workLocation, Valid: workLocation != ""},
		WorkCity:             sql.NullString{String: workCity, Valid: workCity != ""},
		Day:                  sql.NullString{String: day, Valid: day != ""},
		YearWeek:             sql.NullInt64{Int64: int64(yearWeek), Valid: yearWeek > 0},
		Notes:                sql.NullString{String: notes, Valid: notes != ""},
		Blockers:             sql.NullString{String: blockers, Valid: blockers != ""},
		InFlight:             sql.NullString{String: inFlight, Valid: inFlight != ""},
		RollingInOfficeCount: sql.NullInt64{Int64: int64(rollingCount), Valid: true},
		ID:                   int64(id),
	})
	if err != nil {
		http.Error(w, "Failed to update attendance", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Format date for display
		updatedAttendance.Date = util.FormatDate(updatedAttendance.Date)
		// Return the updated row
		data := struct {
			Attendance db.Attendance
		}{
			Attendance: updatedAttendance,
		}
		tmpl := template.Must(template.ParseFiles("internal/templates/attendance/row.html"))
		tmpl.Execute(w, data)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AttendanceHandler) GetRow(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/attendance/")
	idStr = strings.TrimSuffix(idStr, "/row")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	attendance, err := h.queries.GetAttendance(r.Context(), int64(id))
	if err != nil {
		http.Error(w, "Attendance record not found", http.StatusNotFound)
		return
	}

	// Format date for display
	attendance.Date = util.FormatDate(attendance.Date)

	data := struct {
		Attendance db.Attendance
	}{
		Attendance: attendance,
	}

	tmpl := template.Must(template.ParseFiles("internal/templates/attendance/row.html"))
	tmpl.Execute(w, data)
}

func (h *AttendanceHandler) CancelNew(w http.ResponseWriter, r *http.Request) {
	// Return empty string to remove the row
	w.Write([]byte(""))
}

func (h *AttendanceHandler) NewInline(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("internal/templates/attendance/new_inline.html"))
	tmpl.Execute(w, nil)
}

func (h *AttendanceHandler) Import(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to get CSV file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV", http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		http.Error(w, "CSV must have at least a header row and one data row", http.StatusBadRequest)
		return
	}

	// Skip header row
	dataRows := records[1:]
	imported := 0
	skipped := 0
	updated := 0
	overwrite := r.FormValue("overwrite") == "true"

	for _, row := range dataRows {
		if len(row) < 8 {
			continue // Skip malformed rows
		}

		// Parse date
		dateStr := strings.Trim(row[0], `"`)
		date := util.ParseDate(dateStr)
		if date == "" {
			continue // Skip invalid dates
		}

		// Parse other fields (ignore rollingInOfficeCount at index 8)
		workLocation := strings.TrimSpace(row[1])
		workCity := strings.TrimSpace(row[2])
		day := strings.TrimSpace(row[3])
		yearWeekStr := strings.TrimSpace(row[4])
		notes := strings.TrimSpace(row[5])
		blockers := strings.TrimSpace(row[6])
		inFlight := strings.TrimSpace(row[7])

		yearWeek, _ := strconv.Atoi(yearWeekStr)

		// Calculate rolling count
		rollingCount, err := h.calculateRollingInOfficeCount(r.Context(), date)
		if err != nil {
			http.Error(w, "Failed to calculate rolling count", http.StatusInternalServerError)
			return
		}

		// Check if date already exists
		existing, err := h.queries.GetAttendanceByDate(r.Context(), date)
		if err == nil {
			// Date exists
			if overwrite {
				// Update existing record
				_, err = h.queries.UpdateAttendance(r.Context(), db.UpdateAttendanceParams{
					Date:                 date,
					WorkLocation:         sql.NullString{String: workLocation, Valid: workLocation != ""},
					WorkCity:             sql.NullString{String: workCity, Valid: workCity != ""},
					Day:                  sql.NullString{String: day, Valid: day != ""},
					YearWeek:             sql.NullInt64{Int64: int64(yearWeek), Valid: yearWeek > 0},
					Notes:                sql.NullString{String: notes, Valid: notes != ""},
					Blockers:             sql.NullString{String: blockers, Valid: blockers != ""},
					InFlight:             sql.NullString{String: inFlight, Valid: inFlight != ""},
					RollingInOfficeCount: sql.NullInt64{Int64: int64(rollingCount), Valid: true},
					ID:                   existing.ID,
				})
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to update record: %v", err), http.StatusInternalServerError)
					return
				}
				updated++
				continue
			} else {
				// Skip existing record
				skipped++
				continue
			}
		}
		if err != sql.ErrNoRows {
			// Real error (not "no rows found")
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		// Insert record
		_, err = h.queries.CreateAttendance(r.Context(), db.CreateAttendanceParams{
			Date:                 date,
			WorkLocation:         sql.NullString{String: workLocation, Valid: workLocation != ""},
			WorkCity:             sql.NullString{String: workCity, Valid: workCity != ""},
			Day:                  sql.NullString{String: day, Valid: day != ""},
			YearWeek:             sql.NullInt64{Int64: int64(yearWeek), Valid: yearWeek > 0},
			Notes:                sql.NullString{String: notes, Valid: notes != ""},
			Blockers:             sql.NullString{String: blockers, Valid: blockers != ""},
			InFlight:             sql.NullString{String: inFlight, Valid: inFlight != ""},
			RollingInOfficeCount: sql.NullInt64{Int64: int64(rollingCount), Valid: true},
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to import record: %v", err), http.StatusInternalServerError)
			return
		}

		imported++
	}

	// Redirect with success message
	http.Redirect(w, r, "/?imported="+strconv.Itoa(imported)+"&skipped="+strconv.Itoa(skipped)+"&updated="+strconv.Itoa(updated), http.StatusSeeOther)
}

func (h *AttendanceHandler) calculateRollingInOfficeCount(ctx context.Context, currentDate string) (int, error) {
	count, err := h.queries.GetRollingInOfficeCount(ctx, db.GetRollingInOfficeCountParams{
		Date:   currentDate,
		Date_2: currentDate,
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
