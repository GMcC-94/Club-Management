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

// StudentList shows all students
func StudentList(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT id, first_name, last_name, age, weight_range, belt_level, 
		       fight_experience, status
		FROM students
		WHERE deleted_at IS NULL
		ORDER BY first_name, last_name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying students: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var s types.Student
		err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Age, &s.WeightRange, &s.BeltLevel, &s.FightExperience, &s.Status)
		if err != nil {
			log.Printf("Error scanning student: %v", err)
			continue
		}
		students = append(students, s)
	}

	pages.Students(user, students).Render(r.Context(), w)
}

// StudentNew shows the form for creating a new student
func StudentNew(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := getActiveClasses()
	if err != nil {
		log.Printf("Error loading classes: %v", err)
	}

	pages.StudentNew(user, classes).Render(r.Context(), w)
}

// StudentEdit shows the form for editing a student
func StudentEdit(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	studentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	student, err := getStudentByID(studentID)
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	classes, err := getActiveClasses()
	if err != nil {
		log.Printf("Error loading classes: %v", err)
	}

	classIDs, err := getStudentClasses(studentID)
	if err != nil {
		log.Printf("Error loading student classes: %v", err)
	} else {
		student.ClassIDs = classIDs
	}

	pages.StudentEdit(user, student, classes).Render(r.Context(), w)
}

// StudentCreate handles the form submission for creating a student
func StudentCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	ageStr := r.FormValue("age")
	weightRange := r.FormValue("weight_range")
	beltLevel := r.FormValue("belt_level")
	fightExperience := r.FormValue("fight_experience")
	status := r.FormValue("status")
	classIDs := r.Form["classes"]

	if firstName == "" || lastName == "" || beltLevel == "" || status == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	var age sql.NullInt32
	if ageStr != "" {
		if ageVal, err := strconv.Atoi(ageStr); err == nil {
			age = sql.NullInt32{Int32: int32(ageVal), Valid: true}
		}
	}

	var studentID int
	query := `
		INSERT INTO students (first_name, last_name, age, weight_range, belt_level, fight_experience, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err = database.DB.QueryRow(query, firstName, lastName, age, nullString(weightRange), beltLevel, nullString(fightExperience), status, time.Now()).Scan(&studentID)
	if err != nil {
		logger.Error("failed to create student", "error", err, "user", user.Username, "name", firstName+" "+lastName)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("student created", "student_id", studentID, "name", firstName+" "+lastName, "belt_level", beltLevel, "user", user.Username, "user_id", user.ID)

	for _, classIDStr := range classIDs {
		if classID, err := strconv.Atoi(classIDStr); err == nil {
			_, err := database.DB.Exec("INSERT INTO student_classes (student_id, class_id) VALUES ($1, $2)", studentID, classID)
			if err != nil {
				logger.Error("failed to assign class", "error", err, "student_id", studentID, "class_id", classID)
			}
		}
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// StudentUpdate handles the form submission for updating a student
func StudentUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	studentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	ageStr := r.FormValue("age")
	weightRange := r.FormValue("weight_range")
	beltLevel := r.FormValue("belt_level")
	fightExperience := r.FormValue("fight_experience")
	status := r.FormValue("status")
	classIDs := r.Form["classes"]

	if firstName == "" || lastName == "" || beltLevel == "" || status == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	var age sql.NullInt32
	if ageStr != "" {
		if ageVal, err := strconv.Atoi(ageStr); err == nil {
			age = sql.NullInt32{Int32: int32(ageVal), Valid: true}
		}
	}

	query := `
		UPDATE students
		SET first_name = $1, last_name = $2, age = $3, weight_range = $4, 
		    belt_level = $5, fight_experience = $6, status = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	_, err = database.DB.Exec(query, firstName, lastName, age, nullString(weightRange), beltLevel, nullString(fightExperience), status, time.Now(), studentID)
	if err != nil {
		logger.Error("failed to update student", "error", err, "student_id", studentID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("student updated", "student_id", studentID, "name", firstName+" "+lastName, "belt_level", beltLevel, "user", user.Username)

	_, err = database.DB.Exec("DELETE FROM student_classes WHERE student_id = $1", studentID)
	if err != nil {
		logger.Error("failed to remove class assignments", "error", err, "student_id", studentID)
	}

	for _, classIDStr := range classIDs {
		if classID, err := strconv.Atoi(classIDStr); err == nil {
			_, err := database.DB.Exec("INSERT INTO student_classes (student_id, class_id) VALUES ($1, $2)", studentID, classID)
			if err != nil {
				logger.Error("failed to assign class", "error", err, "student_id", studentID, "class_id", classID)
			}
		}
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// StudentDelete handles soft-deleting a student
func StudentDelete(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	studentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE students SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, time.Now(), studentID)
	if err != nil {
		logger.Error("failed to delete student", "error", err, "student_id", studentID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("student deleted", "student_id", studentID, "user", user.Username, "user_id", user.ID)

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// Helper functions
func getStudentByID(id int) (*types.Student, error) {
	query := `
		SELECT id, first_name, last_name, age, weight_range, belt_level, 
		       fight_experience, status
		FROM students
		WHERE id = $1 AND deleted_at IS NULL
	`

	var s types.Student
	err := database.DB.QueryRow(query, id).Scan(&s.ID, &s.FirstName, &s.LastName, &s.Age, &s.WeightRange, &s.BeltLevel, &s.FightExperience, &s.Status)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func getStudentClasses(studentID int) ([]int, error) {
	query := `SELECT class_id FROM student_classes WHERE student_id = $1`
	rows, err := database.DB.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classIDs []int
	for rows.Next() {
		var classID int
		if err := rows.Scan(&classID); err == nil {
			classIDs = append(classIDs, classID)
		}
	}

	return classIDs, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
