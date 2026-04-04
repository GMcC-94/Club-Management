package repository

import (
	"club-management/internal/database"
	"club-management/internal/types"
	"database/sql"
	"fmt"
	"time"
)

// ClassReader defines read operations for classes.
type ClassReader interface {
	List() ([]types.Class, error)
	GetByID(id int) (*types.Class, error)
	GetActive() ([]types.Class, error)
}

// ClassWriter defines write operations for classes.
type ClassWriter interface {
	Create(name string, dayOfWeek int, startTime time.Time, endTime sql.NullTime, isActive bool) error
	Update(id int, name string, dayOfWeek int, startTime time.Time, endTime sql.NullTime, isActive bool) error
	SoftDelete(id int) error
}

// PostgresClassRepository implements ClassReader and ClassWriter.
type PostgresClassRepository struct{}

func NewClassRepository() *PostgresClassRepository {
	return &PostgresClassRepository{}
}

func (r *PostgresClassRepository) List() ([]types.Class, error) {
	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE deleted_at IS NULL
		ORDER BY day_of_week, start_time
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying classes: %w", err)
	}
	defer rows.Close()

	var classes []types.Class
	for rows.Next() {
		var c types.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.DayOfWeek, &c.StartTime, &c.EndTime, &c.IsActive); err != nil {
			return nil, fmt.Errorf("scanning class: %w", err)
		}
		classes = append(classes, c)
	}

	return classes, rows.Err()
}

func (r *PostgresClassRepository) GetByID(id int) (*types.Class, error) {
	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE id = $1 AND deleted_at IS NULL
	`

	var c types.Class
	err := database.DB.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.DayOfWeek, &c.StartTime, &c.EndTime, &c.IsActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("class %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("querying class %d: %w", id, err)
	}

	return &c, nil
}

func (r *PostgresClassRepository) GetActive() ([]types.Class, error) {
	query := `
		SELECT id, name, day_of_week, start_time, end_time, is_active
		FROM classes
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY day_of_week, start_time
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying active classes: %w", err)
	}
	defer rows.Close()

	var classes []types.Class
	for rows.Next() {
		var c types.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.DayOfWeek, &c.StartTime, &c.EndTime, &c.IsActive); err != nil {
			return nil, fmt.Errorf("scanning class: %w", err)
		}
		classes = append(classes, c)
	}

	return classes, rows.Err()
}

func (r *PostgresClassRepository) Create(name string, dayOfWeek int, startTime time.Time, endTime sql.NullTime, isActive bool) error {
	query := `
		INSERT INTO classes (name, day_of_week, start_time, end_time, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := database.DB.Exec(query, name, dayOfWeek, startTime, endTime, isActive, time.Now())
	if err != nil {
		return fmt.Errorf("creating class: %w", err)
	}

	return nil
}

func (r *PostgresClassRepository) Update(id int, name string, dayOfWeek int, startTime time.Time, endTime sql.NullTime, isActive bool) error {
	query := `
		UPDATE classes
		SET name = $1, day_of_week = $2, start_time = $3, end_time = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	_, err := database.DB.Exec(query, name, dayOfWeek, startTime, endTime, isActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("updating class %d: %w", id, err)
	}

	return nil
}

func (r *PostgresClassRepository) SoftDelete(id int) error {
	query := `UPDATE classes SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	_, err := database.DB.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("deleting class %d: %w", id, err)
	}

	return nil
}
