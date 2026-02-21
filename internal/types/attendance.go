package types

import (
	"database/sql"
	"time"
)

// Attendance represents a student's attendance record for a specific class on a specific date
type Attendance struct {
	ID               int
	StudentID        int
	ClassID          int
	Date             time.Time
	Status           string // "present" or "absent"
	IsClassCancelled bool
	OffSchedule      bool // TRUE if student attended outside their normal schedule
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        sql.NullTime
}

// AttendanceRecord is for displaying attendance history with joined data
type AttendanceRecord struct {
	ID               int
	StudentID        int
	ClassID          int
	Date             time.Time
	Status           string
	OffSchedule      bool
	StudentFirstName string
	StudentLastName  string
	ClassName        string
}

// AttendanceEditRecord is for editing attendance records
type AttendanceEditRecord struct {
	ID               int
	StudentID        int
	Status           string
	OffSchedule      bool
	StudentFirstName string
	StudentLastName  string
	BeltLevel        string
}

// AttendanceStats represents attendance statistics for a student
type AttendanceStats struct {
	StudentID       int
	StudentName     string
	TotalExpected   int
	TotalAttended   int
	AttendanceRate  float64
}

// AttendanceReportStats for attendance reports
type AttendanceReportStats struct {
	TotalRecords   int
	TotalPresent   int
	TotalAbsent    int
	AttendanceRate float64
	TopClasses     []ChartDataItem
	TopStudents    []ChartDataItem
	DailyTrend     []TrendDataItem
}
