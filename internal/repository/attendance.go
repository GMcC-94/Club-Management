package repository

import (
	"club-management/internal/database"
	"club-management/internal/types"
	"fmt"
	"time"
)

// AttendanceReader defines read operations for attendance.
type AttendanceReader interface {
	ListHistory() ([]types.AttendanceRecord, error)
	GetEditRecords(classID int, date time.Time) ([]types.AttendanceEditRecord, error)
	GetStudentsForClass(classID string) ([]types.Student, error)
	GetExistingAttendance(classID, dateStr string) (map[int]bool, error)
	GetAllStudentIDsForClass(classID int) ([]int, error)
	GetAttendanceStudentIDs(classID int, dateStr string) ([]int, error)
}

// AttendanceWriter defines write operations for attendance.
type AttendanceWriter interface {
	UpsertAttendance(classID int, date time.Time, allStudentIDs []int, presentMap map[int]bool) error
	UpdateAttendance(classID int, dateStr string, studentIDs []int, presentMap map[int]bool) error
}

// PostgresAttendanceRepository implements AttendanceReader and AttendanceWriter.
type PostgresAttendanceRepository struct{}

func NewAttendanceRepository() *PostgresAttendanceRepository {
	return &PostgresAttendanceRepository{}
}

func (r *PostgresAttendanceRepository) ListHistory() ([]types.AttendanceRecord, error) {
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
		return nil, fmt.Errorf("querying attendance history: %w", err)
	}
	defer rows.Close()

	var records []types.AttendanceRecord
	for rows.Next() {
		var rec types.AttendanceRecord
		if err := rows.Scan(&rec.ID, &rec.StudentID, &rec.ClassID, &rec.Date, &rec.Status, &rec.OffSchedule,
			&rec.StudentFirstName, &rec.StudentLastName, &rec.ClassName); err != nil {
			return nil, fmt.Errorf("scanning attendance record: %w", err)
		}
		records = append(records, rec)
	}

	return records, rows.Err()
}

func (r *PostgresAttendanceRepository) GetEditRecords(classID int, date time.Time) ([]types.AttendanceEditRecord, error) {
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
		return nil, fmt.Errorf("querying attendance edit records: %w", err)
	}
	defer rows.Close()

	var records []types.AttendanceEditRecord
	for rows.Next() {
		var rec types.AttendanceEditRecord
		if err := rows.Scan(&rec.ID, &rec.StudentID, &rec.Status, &rec.OffSchedule,
			&rec.StudentFirstName, &rec.StudentLastName, &rec.BeltLevel); err != nil {
			return nil, fmt.Errorf("scanning attendance edit record: %w", err)
		}
		records = append(records, rec)
	}

	return records, rows.Err()
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
		return nil, fmt.Errorf("loading students for class %s: %w", classID, err)
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var s types.Student
		if err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.BeltLevel, &s.Status); err != nil {
			return nil, fmt.Errorf("scanning student: %w", err)
		}
		students = append(students, s)
	}

	return students, rows.Err()
}

func (r *PostgresAttendanceRepository) GetExistingAttendance(classID, dateStr string) (map[int]bool, error) {
	query := `
		SELECT student_id, status FROM attendance
		WHERE class_id = $1 AND date = $2 AND deleted_at IS NULL
	`

	rows, err := database.DB.Query(query, classID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("querying existing attendance: %w", err)
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

	return attendance, rows.Err()
}

func (r *PostgresAttendanceRepository) GetAllStudentIDsForClass(classID int) ([]int, error) {
	query := `
		SELECT s.id FROM students s
		INNER JOIN student_classes sc ON s.id = sc.student_id
		WHERE sc.class_id = $1 AND s.deleted_at IS NULL AND s.status = 'active'
	`

	rows, err := database.DB.Query(query, classID)
	if err != nil {
		return nil, fmt.Errorf("querying student IDs for class %d: %w", classID, err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning student ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (r *PostgresAttendanceRepository) GetAttendanceStudentIDs(classID int, dateStr string) ([]int, error) {
	query := `SELECT student_id FROM attendance WHERE class_id = $1 AND date = $2 AND deleted_at IS NULL`

	rows, err := database.DB.Query(query, classID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("querying attendance student IDs: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning student ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (r *PostgresAttendanceRepository) UpsertAttendance(classID int, date time.Time, allStudentIDs []int, presentMap map[int]bool) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, studentID := range allStudentIDs {
		status := "absent"
		if presentMap[studentID] {
			status = "present"
		}
		_, err := tx.Exec(`
			INSERT INTO attendance (student_id, class_id, date, status)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (student_id, class_id, date) DO UPDATE SET status = $4, updated_at = NOW()
		`, studentID, classID, date, status)
		if err != nil {
			return fmt.Errorf("upserting attendance for student %d: %w", studentID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing attendance: %w", err)
	}

	return nil
}

func (r *PostgresAttendanceRepository) UpdateAttendance(classID int, dateStr string, studentIDs []int, presentMap map[int]bool) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, studentID := range studentIDs {
		status := "absent"
		if presentMap[studentID] {
			status = "present"
		}
		_, err := tx.Exec(
			`UPDATE attendance SET status = $1, updated_at = NOW()
			 WHERE student_id = $2 AND class_id = $3 AND date = $4`,
			status, studentID, classID, dateStr,
		)
		if err != nil {
			return fmt.Errorf("updating attendance for student %d: %w", studentID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing attendance update: %w", err)
	}

	return nil
}
