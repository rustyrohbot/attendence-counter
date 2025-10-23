package main

import (
	"log"
	"net/http"
	"strings"

	"attendance-tracker/internal/database"
	"attendance-tracker/internal/db"
	"attendance-tracker/internal/handlers"
)

func main() {
	// Initialize database
	dbConn, err := database.New("")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer dbConn.Close()

	// Create queries instance
	queries := db.New(dbConn)

	// Create handlers
	attendanceHandler := handlers.NewAttendanceHandler(queries, dbConn)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Attendance routes
	mux.HandleFunc("/", attendanceHandler.List)
	mux.HandleFunc("/attendance/new", attendanceHandler.New)
	mux.HandleFunc("/attendance/import", attendanceHandler.Import)
	mux.HandleFunc("/attendance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			attendanceHandler.Create(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/attendance/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/edit") {
			attendanceHandler.Edit(w, r)
		} else if r.Method == http.MethodPost {
			attendanceHandler.Update(w, r)
		} else if r.Method == http.MethodDelete {
			attendanceHandler.Delete(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Start server
	log.Println("Server starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
