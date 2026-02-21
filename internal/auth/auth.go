package auth

import (
	"club-management/internal/database"
	"club-management/internal/types"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	Store *sessions.CookieStore

	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
)

func init() {
	sessionKey := os.Getenv("SESSION_SECRET")
	if sessionKey == "" {
		sessionKey = "change-this-to-a-secure-random-key-32-bytes-long"
	}
	Store = sessions.NewCookieStore([]byte(sessionKey))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if password meets requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

// AuthenticateUser validates credentials and returns user if valid
func AuthenticateUser(username, password string) (*types.User, error) {
	var user types.User

	query := `
		SELECT id, username, email, password_hash, role, must_change_password, created_at, updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	err := database.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// CreateUser creates a new user in the database
func CreateUser(username, password, email, role string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	query := `
		INSERT INTO users (username, email, password_hash, role, must_change_password)
		VALUES ($1, $2, $3, $4, $5)
	`

	var emailVal interface{} = sql.NullString{}
	if email != "" {
		emailVal = email
	}

	_, err = database.DB.Exec(query, username, emailVal, hash, role, true)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*types.User, error) {
	var user types.User

	query := `
		SELECT id, username, email, password_hash, role, must_change_password, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := database.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	return &user, nil
}

// UpdatePassword updates a user's password
func UpdatePassword(userID int, newPassword string, clearMustChange bool) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	query := `
		UPDATE users 
		SET password_hash = $1, must_change_password = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	_, err = database.DB.Exec(query, hash, !clearMustChange, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	return nil
}

// SetSession creates a session for the authenticated user
func SetSession(w http.ResponseWriter, r *http.Request, user *types.User) error {
	session, err := Store.Get(r, "club-session")
	if err != nil {
		return err
	}

	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	session.Values["must_change_password"] = user.MustChangePassword

	return session.Save(r, w)
}

// GetSession retrieves the current session
func GetSession(r *http.Request) (*sessions.Session, error) {
	return Store.Get(r, "club-session")
}

// ClearSession destroys the current session
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := Store.Get(r, "club-session")
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	return session.Save(r, w)
}

// GetCurrentUser returns the current logged-in user from session
func GetCurrentUser(r *http.Request) (*types.User, error) {
	session, err := GetSession(r)
	if err != nil {
		return nil, err
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		return nil, errors.New("no user in session")
	}

	return GetUserByID(userID)
}

// IsAuthenticated checks if user is logged in
func IsAuthenticated(r *http.Request) bool {
	session, err := GetSession(r)
	if err != nil {
		return false
	}

	_, ok := session.Values["user_id"]
	return ok
}

// RequireAuth middleware ensures user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// RequireAdmin middleware ensures user is an admin
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r)
		if err != nil || session.Values["role"] != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// RequireTreasurer ensures user is admin or treasurer
func RequireTreasurer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		role, ok := session.Values["role"].(string)
		if !ok || (role != "admin" && role != "treasurer") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
