package repository

import (
	"club-management/internal/database"
	"club-management/internal/types"
)

// AttendanceRepository defines the data access methods for attendance
type AttendanceRepository interface {
	GetStudentsForClass(classID string) ([]types.Student, error)
	GetExistingAttendance(classID, dateStr string) (map[int]bool, error)
}

// PostgresAttendanceRepository is the PostgreSQL implementation
type PostgresAttendanceRepository struct{}

func NewAttendanceRepository() AttendanceRepository {
	return &PostgresAttendanceRepository{}
}

func (r *PostgresAttendanceRepository) GetStudentsForClass(classID string) ([]types.Student, error) {
	query := `
		SELECT s.id, s.first_name, s.last_name, s.belt_level, s.status
		FROM students s
		INNER JOIN student_classes sc ON s.id = sc.student_id
		WHERE sc.class_id = $1 AND s.deleted_at IS NULL AND s.status = 'active'
		ORDER BY s.first_name, s.last_name
	`

	rows, err := database.DB.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var s types.Student
		if err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.BeltLevel, &s.Status); err != nil {
			continue
		}
		students = append(students, s)
	}
	return students, nil
}

func (r *PostgresAttendanceRepository) GetExistingAttendance(classID, dateStr string) (map[int]bool, error) {
	rows, err := database.DB.Query(`
		SELECT student_id, status FROM attendance
		WHERE class_id = $1 AND date = $2 AND deleted_at IS NULL
	`, classID, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attendance := make(map[int]bool)
	for rows.Next() {
		var studentID int
		var status string
		if err := rows.Scan(&studentID, &status); err == nil {
			attendance[studentID] = status == "present"
		}
	}
	return attendance, nil
}
