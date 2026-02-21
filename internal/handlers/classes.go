package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ClassList shows all classes
func ClassList(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE deleted_at IS NULL
		ORDER BY day_of_week, start_time
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying classes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var classes []types.Class
	for rows.Next() {
		var c types.Class
		err := rows.Scan(&c.ID, &c.Name, &c.DayOfWeek, &c.StartTime, &c.EndTime, &c.IsActive)
		if err != nil {
			log.Printf("Error scanning class: %v", err)
			continue
		}
		classes = append(classes, c)
	}

	pages.Classes(user, classes).Render(r.Context(), w)
}

// ClassNew shows the form for creating a new class
func ClassNew(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pages.ClassNew(user).Render(r.Context(), w)
}

// ClassEdit shows the form for editing a class
func ClassEdit(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	class, err := getClassByID(classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	pages.ClassEdit(user, class).Render(r.Context(), w)
}

// ClassCreate handles the form submission for creating a class
func ClassCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	dayOfWeekStr := r.FormValue("day_of_week")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	isActiveStr := r.FormValue("is_active")

	if name == "" || dayOfWeekStr == "" || startTimeStr == "" || isActiveStr == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	dayOfWeek, err := strconv.Atoi(dayOfWeekStr)
	if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
		http.Error(w, "Invalid day of week", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	var endTime sql.NullTime
	if endTimeStr != "" {
		t, err := time.Parse("15:04", endTimeStr)
		if err == nil {
			endTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	isActive := isActiveStr == "true"

	query := `
		INSERT INTO classes (name, day_of_week, start_time, end_time, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = database.DB.Exec(query, name, dayOfWeek, startTime, endTime, isActive, time.Now())
	if err != nil {
		logger.Error("failed to create class", "error", err, "name", name, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class created", "name", name, "day_of_week", dayOfWeek, "user", user.Username)

	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// ClassUpdate handles the form submission for updating a class
func ClassUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	dayOfWeekStr := r.FormValue("day_of_week")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	isActiveStr := r.FormValue("is_active")

	if name == "" || dayOfWeekStr == "" || startTimeStr == "" || isActiveStr == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	dayOfWeek, err := strconv.Atoi(dayOfWeekStr)
	if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
		http.Error(w, "Invalid day of week", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	var endTime sql.NullTime
	if endTimeStr != "" {
		t, err := time.Parse("15:04", endTimeStr)
		if err == nil {
			endTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	isActive := isActiveStr == "true"

	query := `
		UPDATE classes
		SET name = $1, day_of_week = $2, start_time = $3, end_time = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	_, err = database.DB.Exec(query, name, dayOfWeek, startTime, endTime, isActive, time.Now(), classID)
	if err != nil {
		logger.Error("failed to update class", "error", err, "class_id", classID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class updated", "class_id", classID, "name", name, "user", user.Username)

	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// ClassDelete handles soft-deleting a class
func ClassDelete(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE classes SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, time.Now(), classID)
	if err != nil {
		logger.Error("failed to delete class", "error", err, "class_id", classID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class deleted", "class_id", classID, "user", user.Username)

	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// Helper functions
func getClassByID(id int) (*types.Class, error) {
	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE id = $1 AND deleted_at IS NULL
	`

	var class types.Class
	err := database.DB.QueryRow(query, id).Scan(&class.ID, &class.Name, &class.DayOfWeek, &class.StartTime, &class.EndTime, &class.IsActive)
	if err != nil {
		return nil, err
	}

	return &class, nil
}

func getActiveClasses() ([]types.Class, error) {
	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY day_of_week, start_time
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []types.Class
	for rows.Next() {
		var c types.Class
		err := rows.Scan(&c.ID, &c.Name, &c.DayOfWeek, &c.StartTime, &c.EndTime, &c.IsActive)
		if err != nil {
			continue
		}
		classes = append(classes, c)
	}

	return classes, nil
}
