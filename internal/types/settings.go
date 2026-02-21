package types

import (
	"time"
)

// Setting represents an app configuration setting
type Setting struct {
	ID        int
	Key       string
	Value     string
	UpdatedAt time.Time
}

// Settings represents club settings
type Settings struct {
	ClubName   string
	Address    string
	Phone      string
	Email      string
	Website    string
	Currency   string
	DateFormat string
}
