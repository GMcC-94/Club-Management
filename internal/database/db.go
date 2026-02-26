package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect initializes the database connection
func Connect() error {
	host := getEnv("DB_HOST", os.Getenv("DB_HOST"))
	port := getEnv("DB_PORT", os.Getenv("DB_PORT"))
	user := getEnv("DB_USER", os.Getenv("DB_USER"))
	password := getEnv("DB_PASSWORD", os.Getenv("DB_PASSWORD"))
	dbname := getEnv("DB_NAME", os.Getenv("DB_NAME"))

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
