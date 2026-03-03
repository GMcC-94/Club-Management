package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/logger"
	"club-management/web/templates/pages"
	"net/http"
)

// LoginPage shows the login form
func LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login("").Render(r.Context(), w)
}

// Login handles the login form submission
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := auth.AuthenticateUser(username, password)
	if err != nil {
		logger.Warn("failed login attempt", "username", username, "ip", r.RemoteAddr)
		pages.Login("Invalid username or password").Render(r.Context(), w)
		return
	}

	err = auth.SetSession(w, r, user)
	if err != nil {
		logger.Error("failed to create session", "error", err, "username", username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("successful login", "username", username, "role", user.Role, "ip", r.RemoteAddr)

	if user.MustChangePassword {
		http.Redirect(w, r, "/change-password", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetCurrentUser(r)

	err := auth.ClearSession(w, r)
	if err != nil {
		logger.Error("failed to clear session", "error", err)
	}

	if user != nil {
		logger.Info("user logout", "username", user.Username, "ip", r.RemoteAddr)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ChangePasswordPage shows the change password form
func ChangePasswordPage(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Create change password template
	_ = user
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ChangePassword handles the change password form submission
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Implement password change functionality
	_ = user
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
