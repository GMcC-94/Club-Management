package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"log"
	"net/http"
	"strconv"
	"time"
)

// CategoryList shows all categories
func CategoryList(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT id, name, type
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY type, name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []types.Category
	for rows.Next() {
		var c types.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Type)
		if err != nil {
			log.Printf("Error scanning category: %v", err)
			continue
		}
		categories = append(categories, c)
	}

	pages.Categories(user, categories).Render(r.Context(), w)
}

// CategoryCreate handles creating a new category
func CategoryCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	catType := r.FormValue("type")

	if name == "" || catType == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO categories (name, type) VALUES ($1, $2)`
	_, err = database.DB.Exec(query, name, catType)
	if err != nil {
		logger.Error("failed to create category", "error", err, "name", name, "type", catType, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("category created", "name", name, "type", catType, "user", user.Username)

	http.Redirect(w, r, "/categories/manage", http.StatusSeeOther)
}

// CategoryDelete handles soft-deleting a category
func CategoryDelete(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categoryID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE categories SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, time.Now(), categoryID)
	if err != nil {
		logger.Error("failed to delete category", "error", err, "category_id", categoryID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("category deleted", "category_id", categoryID, "user", user.Username)

	http.Redirect(w, r, "/categories/manage", http.StatusSeeOther)
}

// Helper function
func getAllCategories() ([]types.Category, error) {
	query := `
		SELECT id, name, type
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY type, name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []types.Category
	for rows.Next() {
		var c types.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Type)
		if err != nil {
			continue
		}
		categories = append(categories, c)
	}

	return categories, nil
}
