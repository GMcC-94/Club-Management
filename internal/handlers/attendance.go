package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"log"
	"net/http"
	"strconv"
	"time"
)

// TakeAttendance shows the form for taking attendance
func TakeAttendance(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := getActiveClasses()
	if err != nil {
		log.Printf("Error loading classes: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	pages.TakeAttendance(user, classes, today).Render(r.Context(), w)
}

// LoadStudentsForClass returns students for a specific class
func LoadStudentsForClass(w http.ResponseWriter, r *http.Request) {
	classID := r.URL.Query().Get("class_id")
	dateStr := r.URL.Query().Get("date")

	if classID == "" || dateStr == "" {
		http.Error(w, "Missing class_id or date", http.StatusBadRequest)
		return
	}

	query := `
		SELECT s.id, s.first_name, s.last_name, s.belt_level, s.status
		FROM students s
		INNER JOIN student_classes sc ON s.id = sc.student_id
		WHERE sc.class_id = $1 AND s.deleted_at IS NULL AND s.status = 'active'
		ORDER BY s.first_name, s.last_name
	`

	rows, err := database.DB.Query(query, classID)
	if err != nil {
		log.Printf("Error querying students: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var s types.Student
		err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.BeltLevel, &s.Status)
		if err != nil {
			log.Printf("Error scanning student: %v", err)
			continue
		}
		students = append(students, s)
	}

	existingAttendance, err := getExistingAttendance(classID, dateStr)
	if err != nil {
		log.Printf("Error loading existing attendance: %v", err)
	}

	// Return simple HTML fragment with student checkboxes
	w.Header().Set("Content-Type", "text/html")
	for _, s := range students {
		checked := ""
		if existingAttendance != nil && existingAttendance[s.ID] {
			checked = "checked"
		}
		w.Write([]byte(`<label class="flex items-center gap-2 p-2 hover:bg-white/5 rounded"><input type="checkbox" name="present" value="` + strconv.Itoa(s.ID) + `" ` + checked + ` class="form-checkbox"><span>` + s.FirstName + ` ` + s.LastName + ` (` + s.BeltLevel + `)</span></label>`))
	}
}

// SubmitAttendance handles the form submission for recording attendance
func SubmitAttendance(w http.ResponseWriter, r *http.Request) {
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

	tx, err := database.DB.Begin()
	if err != nil {
		logger.Error("failed to begin transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM attendance_records WHERE class_id = $1 AND date = $2", classID, date)
	if err != nil {
		logger.Error("failed to delete existing attendance", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT id FROM students s
		INNER JOIN student_classes sc ON s.id = sc.student_id
		WHERE sc.class_id = $1 AND s.deleted_at IS NULL AND s.status = 'active'
	`

	rows, err := tx.Query(query, classID)
	if err != nil {
		logger.Error("failed to query students", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allStudentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			allStudentIDs = append(allStudentIDs, id)
		}
	}
	rows.Close()

	presentMap := make(map[int]bool)
	for _, studentIDStr := range presentStudents {
		if studentID, err := strconv.Atoi(studentIDStr); err == nil {
			presentMap[studentID] = true
		}
	}

	presentCount := len(presentStudents)
	absentCount := len(allStudentIDs) - presentCount

	var attendanceID int
	attendanceQuery := `
		INSERT INTO attendance (class_id, date, present_count, absent_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (class_id, date) DO UPDATE
		SET present_count = $3, absent_count = $4
		RETURNING id
	`

	err = tx.QueryRow(attendanceQuery, classID, date, presentCount, absentCount).Scan(&attendanceID)
	if err != nil {
		logger.Error("failed to insert attendance", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, studentID := range allStudentIDs {
		present := presentMap[studentID]
		_, err := tx.Exec(
			"INSERT INTO attendance_records (attendance_id, student_id, present, class_id, date) VALUES ($1, $2, $3, $4, $5)",
			attendanceID, studentID, present, classID, date,
		)
		if err != nil {
			logger.Error("failed to insert attendance record", "error", err, "student_id", studentID)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("failed to commit transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("attendance recorded", "class_id", classID, "date", dateStr, "user", user.Username)

	http.Redirect(w, r, "/attendance/take", http.StatusSeeOther)
}

// AttendanceHistory shows historical attendance records
func AttendanceHistory(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT a.id, a.student_id, a.class_id, a.date, a.status, a.off_schedule,
		       s.first_name, s.last_name, c.name as class_name
		FROM attendance a
		LEFT JOIN students s ON a.student_id = s.id
		LEFT JOIN classes c ON a.class_id = c.id
		ORDER BY a.date DESC, a.class_id
		LIMIT 500
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying attendance: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []types.AttendanceRecord
	for rows.Next() {
		var r types.AttendanceRecord
		err := rows.Scan(&r.ID, &r.StudentID, &r.ClassID, &r.Date, &r.Status, &r.OffSchedule, &r.StudentFirstName, &r.StudentLastName, &r.ClassName)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		records = append(records, r)
	}

	pages.AttendanceHistory(user, records).Render(r.Context(), w)
}

// AttendanceEdit shows the form for editing attendance for a specific class/date
func AttendanceEdit(w http.ResponseWriter, r *http.Request) {
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

	class, err := getClassByID(classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	query := `
		SELECT a.id, a.student_id, a.status, a.off_schedule,
		       s.first_name, s.last_name, s.belt_level
		FROM attendance a
		INNER JOIN students s ON a.student_id = s.id
		WHERE a.class_id = $1 AND a.date = $2
		ORDER BY s.first_name, s.last_name
	`

	rows, err := database.DB.Query(query, classID, date)
	if err != nil {
		log.Printf("Error querying attendance: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []types.AttendanceEditRecord
	for rows.Next() {
		var r types.AttendanceEditRecord
		err := rows.Scan(&r.ID, &r.StudentID, &r.Status, &r.OffSchedule, &r.StudentFirstName, &r.StudentLastName, &r.BeltLevel)
		if err != nil {
			log.Printf("Error scanning record: %v", err)
			continue
		}
		records = append(records, r)
	}

	pages.AttendanceEdit(user, class, date, records).Render(r.Context(), w)
}

// AttendanceUpdate handles updating existing attendance
func AttendanceUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	attendanceID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid attendance ID", http.StatusBadRequest)
		return
	}

	presentStudents := r.Form["present"]

	var classID int
	var date time.Time
	err = database.DB.QueryRow("SELECT class_id, date FROM attendance WHERE id = $1", attendanceID).Scan(&classID, &date)
	if err != nil {
		http.Error(w, "Attendance not found", http.StatusNotFound)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		logger.Error("failed to begin transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var allStudentIDs []int
	rows, err := tx.Query("SELECT student_id FROM attendance_records WHERE attendance_id = $1", attendanceID)
	if err != nil {
		logger.Error("failed to query attendance records", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			allStudentIDs = append(allStudentIDs, id)
		}
	}
	rows.Close()

	presentMap := make(map[int]bool)
	for _, studentIDStr := range presentStudents {
		if studentID, err := strconv.Atoi(studentIDStr); err == nil {
			presentMap[studentID] = true
		}
	}

	for _, studentID := range allStudentIDs {
		present := presentMap[studentID]
		_, err := tx.Exec("UPDATE attendance_records SET present = $1 WHERE attendance_id = $2 AND student_id = $3", present, attendanceID, studentID)
		if err != nil {
			logger.Error("failed to update attendance record", "error", err, "student_id", studentID)
		}
	}

	presentCount := len(presentStudents)
	absentCount := len(allStudentIDs) - presentCount

	_, err = tx.Exec("UPDATE attendance SET present_count = $1, absent_count = $2 WHERE id = $3", presentCount, absentCount, attendanceID)
	if err != nil {
		logger.Error("failed to update attendance", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("failed to commit transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("attendance updated", "class_id", classID, "date", date.Format("2006-01-02"), "user", user.Username)

	http.Redirect(w, r, "/attendance", http.StatusSeeOther)
}

// Helper functions
func getExistingAttendance(classID, dateStr string) (map[int]bool, error) {
	query := `
		SELECT ar.student_id, ar.present
		FROM attendance_records ar
		INNER JOIN attendance a ON ar.attendance_id = a.id
		WHERE a.class_id = $1 AND a.date = $2
	`

	rows, err := database.DB.Query(query, classID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attendance := make(map[int]bool)
	for rows.Next() {
		var studentID int
		var present bool
		if err := rows.Scan(&studentID, &present); err == nil {
			attendance[studentID] = present
		}
	}

	return attendance, nil
}

func mustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// AttendanceList is an alias for AttendanceHistory
func AttendanceList(w http.ResponseWriter, r *http.Request) {
	AttendanceHistory(w, r)
}

// GetClassStudents returns students for a specific class (alias for LoadStudentsForClass)
func GetClassStudents(w http.ResponseWriter, r *http.Request) {
	LoadStudentsForClass(w, r)
}

// SearchStudents searches for students by name
func SearchStudents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	searchQuery := `
		SELECT id, first_name, last_name, belt_level, status
		FROM students
		WHERE deleted_at IS NULL 
		AND status = 'active'
		AND (LOWER(first_name) LIKE LOWER($1) OR LOWER(last_name) LIKE LOWER($1))
		ORDER BY first_name, last_name
		LIMIT 20
	`

	rows, err := database.DB.Query(searchQuery, "%"+query+"%")
	if err != nil {
		log.Printf("Error searching students: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var s types.Student
		err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.BeltLevel, &s.Status)
		if err != nil {
			log.Printf("Error scanning student: %v", err)
			continue
		}
		students = append(students, s)
	}

	// Return simple HTML fragment with student results
	w.Header().Set("Content-Type", "text/html")
	for _, s := range students {
		w.Write([]byte(`<div class="p-2 hover:bg-white/5 rounded cursor-pointer">` + s.FirstName + ` ` + s.LastName + ` (` + s.BeltLevel + `)</div>`))
	}
}
