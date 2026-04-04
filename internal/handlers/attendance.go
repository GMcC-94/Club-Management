package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/logger"
	"club-management/internal/repository"
	"club-management/web/templates/pages"
	"net/http"
	"strconv"
	"time"
)

// AttendanceHandler handles all attendance-related HTTP requests.
type AttendanceHandler struct {
	reader      repository.AttendanceReader
	writer      repository.AttendanceWriter
	classReader repository.ClassReader
}

// NewAttendanceHandler creates an AttendanceHandler with the given dependencies.
func NewAttendanceHandler(reader repository.AttendanceReader, writer repository.AttendanceWriter, classReader repository.ClassReader) *AttendanceHandler {
	return &AttendanceHandler{reader: reader, writer: writer, classReader: classReader}
}

// Take shows the form for taking attendance.
func (h *AttendanceHandler) Take(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := h.classReader.GetActive()
	if err != nil {
		logger.Error("failed to load classes", "error", err)
	}

	today := time.Now().Format("2006-01-02")
	pages.TakeAttendance(user, classes, today).Render(r.Context(), w)
}

// LoadStudentsForClass returns students for a specific class as an HTML fragment (HTMX endpoint).
func (h *AttendanceHandler) LoadStudentsForClass(w http.ResponseWriter, r *http.Request) {
	classID := r.URL.Query().Get("class_id")
	dateStr := r.URL.Query().Get("date")

	if classID == "" || dateStr == "" {
		http.Error(w, "Missing class_id or date", http.StatusBadRequest)
		return
	}

	students, err := h.reader.GetStudentsForClass(classID)
	if err != nil {
		logger.Error("failed to query students", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	existingAttendance, err := h.reader.GetExistingAttendance(classID, dateStr)
	if err != nil {
		logger.Error("failed to load existing attendance", "error", err)
	}

	w.Header().Set("Content-Type", "text/html")
	for _, s := range students {
		checked := ""
		if existingAttendance != nil && existingAttendance[s.ID] {
			checked = "checked"
		}
		w.Write([]byte(`<label class="flex items-center gap-2 p-2 hover:bg-white/5 rounded"><input type="checkbox" name="present" value="` + strconv.Itoa(s.ID) + `" ` + checked + ` class="form-checkbox"><span>` + s.FirstName + ` ` + s.LastName + `</span></label>`))
	}
}

// Submit handles the form submission for recording attendance.
func (h *AttendanceHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classIDStr := r.FormValue("class_id")
	dateStr := r.FormValue("date")
	presentStudents := r.Form["present"]

	if classIDStr == "" || dateStr == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	presentMap := make(map[int]bool)
	for _, studentIDStr := range presentStudents {
		if studentID, err := strconv.Atoi(studentIDStr); err == nil {
			presentMap[studentID] = true
		}
	}

	allStudentIDs, err := h.reader.GetAllStudentIDsForClass(classID)
	if err != nil {
		logger.Error("failed to query students", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.writer.UpsertAttendance(classID, date, allStudentIDs, presentMap); err != nil {
		logger.Error("failed to submit attendance", "error", err, "class_id", classID, "date", dateStr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("attendance recorded", "class_id", classID, "date", dateStr, "user", user.Username)
	http.Redirect(w, r, "/attendance/take", http.StatusSeeOther)
}

// List shows historical attendance records.
func (h *AttendanceHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	records, err := h.reader.ListHistory()
	if err != nil {
		logger.Error("failed to query attendance history", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.AttendanceHistory(user, records).Render(r.Context(), w)
}

// Edit shows the form for editing attendance for a specific class/date.
func (h *AttendanceHandler) Edit(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classIDStr := r.URL.Query().Get("class_id")
	dateStr := r.URL.Query().Get("date")

	if classIDStr == "" || dateStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	class, err := h.classReader.GetByID(classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	records, err := h.reader.GetEditRecords(classID, date)
	if err != nil {
		logger.Error("failed to query attendance records", "error", err, "class_id", classID, "date", dateStr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.AttendanceEdit(user, class, date, records).Render(r.Context(), w)
}

// Update handles updating existing attendance for a class/date.
func (h *AttendanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classIDStr := r.FormValue("class_id")
	dateStr := r.FormValue("date")
	presentStudents := r.Form["present"]

	if classIDStr == "" || dateStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	presentMap := make(map[int]bool)
	for _, studentIDStr := range presentStudents {
		if studentID, err := strconv.Atoi(studentIDStr); err == nil {
			presentMap[studentID] = true
		}
	}

	allStudentIDs, err := h.reader.GetAttendanceStudentIDs(classID, dateStr)
	if err != nil {
		logger.Error("failed to query attendance", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.writer.UpdateAttendance(classID, dateStr, allStudentIDs, presentMap); err != nil {
		logger.Error("failed to update attendance", "error", err, "class_id", classID, "date", dateStr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("attendance updated", "class_id", classID, "date", dateStr, "user", user.Username)
	http.Redirect(w, r, "/attendance", http.StatusSeeOther)
}
