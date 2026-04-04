package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/repository"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ClassHandler handles all class-related HTTP requests.
type ClassHandler struct {
	reader repository.ClassReader
	writer repository.ClassWriter
}

// NewClassHandler creates a ClassHandler with the given dependencies.
func NewClassHandler(reader repository.ClassReader, writer repository.ClassWriter) *ClassHandler {
	return &ClassHandler{reader: reader, writer: writer}
}

// List shows all classes.
func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := h.reader.List()
	if err != nil {
		logger.Error("failed to list classes", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.Classes(user, classes).Render(r.Context(), w)
}

// New shows the form for creating a new class.
func (h *ClassHandler) New(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pages.ClassNew(user).Render(r.Context(), w)
}

// Edit shows the form for editing a class.
func (h *ClassHandler) Edit(w http.ResponseWriter, r *http.Request) {
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

	class, err := h.reader.GetByID(classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	pages.ClassEdit(user, class).Render(r.Context(), w)
}

// Create handles the form submission for creating a class.
func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	if err := h.writer.Create(name, dayOfWeek, startTime, endTime, isActive); err != nil {
		logger.Error("failed to create class", "error", err, "name", name, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class created", "name", name, "day_of_week", dayOfWeek, "user", user.Username)
	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// Update handles the form submission for updating a class.
func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	if err := h.writer.Update(classID, name, dayOfWeek, startTime, endTime, isActive); err != nil {
		logger.Error("failed to update class", "error", err, "class_id", classID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class updated", "class_id", classID, "name", name, "user", user.Username)
	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// Delete handles soft-deleting a class.
func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.writer.SoftDelete(classID); err != nil {
		logger.Error("failed to delete class", "error", err, "class_id", classID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("class deleted", "class_id", classID, "user", user.Username)
	http.Redirect(w, r, "/classes", http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// Legacy helper functions — still used by other handlers (students, attendance,
// reports) that haven't been migrated yet. Remove once those domains use
// ClassReader via their own handler structs.
// ---------------------------------------------------------------------------

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
			log.Printf("Error scanning class: %v", err)
			continue
		}
		classes = append(classes, c)
	}

	return classes, nil
}
