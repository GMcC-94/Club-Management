package types

import (
	"database/sql"
	"time"
)

// Class represents a scheduled training class
type Class struct {
	ID        int
	Name      string
	DayOfWeek int // 0 = Sunday, 1 = Monday, ..., 6 = Saturday
	StartTime time.Time
	EndTime   sql.NullTime
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}
