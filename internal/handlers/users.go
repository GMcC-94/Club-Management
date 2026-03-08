package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

// UserList shows all users
func UserList(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT id, username, email, role, must_change_password
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY username
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []types.User
	for rows.Next() {
		var u types.User
		var email sql.NullString
		err := rows.Scan(&u.ID, &u.Username, &email, &u.Role, &u.MustChangePassword)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		u.Email = email
		users = append(users, u)
	}

	pages.UserList(user, users).Render(r.Context(), w)
}

func UserNew(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pages.UserNew(user, "", "", "").Render(r.Context(), w)
}

// UserEdit shows the form for editing a user
func UserEdit(w http.ResponseWriter, r *http.Request) {
	currentUser, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	editUser, err := getUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	pages.UserEdit(currentUser, editUser).Render(r.Context(), w)
}

// UserCreate handles the form submission for creating a user
func UserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if username == "" || email == "" || password == "" || role == "" {
		pages.UserNew(user, "All fields are required", username, email).Render(r.Context(), w)
		return
	}

	if role != "admin" && role != "coach" && role != "treasurer" {
		pages.UserNew(user, "Invalid role selected", username, email).Render(r.Context(), w)
		return
	}

	err = auth.CreateUser(username, password, email, role)
	if err != nil {
		logger.Error("failed to create user", "error", err, "username", username, "created_by", user.Username)
		pages.UserNew(user, err.Error(), username, email).Render(r.Context(), w)
		return
	}

	logger.Info("user created", "username", username, "role", role, "created_by", user.Username)
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// UserUpdate handles the form submission for updating a user
func UserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if username == "" || email == "" || role == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	if role != "admin" && role != "coach" && role != "treasurer" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	var emailVal interface{} = sql.NullString{}
	if email != "" {
		emailVal = email
	}

	query := `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, username, emailVal, role, userID)
	if err != nil {
		logger.Error("failed to update user", "error", err, "user_id", userID, "updated_by", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if password != "" {
		if err := auth.ValidatePassword(password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			logger.Error("failed to hash password", "error", err, "user_id", userID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec("UPDATE users SET password_hash = $1, must_change_password = true WHERE id = $2", hash, userID)
		if err != nil {
			logger.Error("failed to update password", "error", err, "user_id", userID)
		}
	}

	logger.Info("user updated", "user_id", userID, "updated_by", user.Username)

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// UserDelete handles soft-deleting a user
func UserDelete(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if userID == user.ID {
		http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
		return
	}

	query := `UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, userID)
	if err != nil {
		logger.Error("failed to delete user", "error", err, "user_id", userID, "deleted_by", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("user deleted", "user_id", userID, "deleted_by", user.Username)

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// Helper function
func getUserByID(id int) (*types.User, error) {
	query := `
		SELECT id, username, email, role, must_change_password
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var u types.User
	var email sql.NullString
	err := database.DB.QueryRow(query, id).Scan(&u.ID, &u.Username, &email, &u.Role, &u.MustChangePassword)
	if err != nil {
		return nil, err
	}

	u.Email = email

	return &u, nil
}
