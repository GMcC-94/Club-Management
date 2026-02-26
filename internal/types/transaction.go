package types

import (
	"database/sql"
	"time"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID              int
	Type            string // "income" or "expenditure"
	Amount          float64
	Date            time.Time
	Description     string
	CategoryID      sql.NullInt32
	CategoryName    string
	PaymentMethod   sql.NullString
	ApprovedBy      sql.NullString
	ReceiptFilename sql.NullString
	CreatedBy       sql.NullInt32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
}

// Category represents a transaction category
type Category struct {
	ID        int
	Name      string
	CreatedAt time.Time
	DeletedAt sql.NullTime
}
