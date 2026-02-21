package handlers

import (
	"club-management/internal/auth"
	"club-management/web/templates/pages"
	"net/http"
)

// Index redirects to dashboard if logged in, otherwise to login
func Index(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Dashboard shows the main dashboard
func Dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pages.Dashboard(user).Render(r.Context(), w)
}
