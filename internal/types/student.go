package types

import (
	"database/sql"
	"time"
)

// Student represents a club member
type Student struct {
	ID              int
	FirstName       string
	LastName        string
	Age             sql.NullInt32
	WeightRange     sql.NullString
	BeltLevel       string
	FightExperience sql.NullString
	Status          string // "active" or "inactive"
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
	ClassIDs        []int // For form binding
}

// StudentWithClasses represents a student along with their scheduled classes
type StudentWithClasses struct {
	Student
	ClassIDs []int
}
