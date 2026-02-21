package types

import (
	"database/sql"
	"time"
)

// User represents a system user (admin or coach)
type User struct {
	ID                 int
	Username           string
	Email              sql.NullString
	PasswordHash       string
	Role               string // "admin" or "coach"
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          sql.NullTime
}
