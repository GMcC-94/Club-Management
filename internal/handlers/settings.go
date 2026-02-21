package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"log"
	"net/http"
)

// SettingsPage shows the settings form
func SettingsPage(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	settingsMap := make(map[string]string)
	query := `SELECT key, value FROM settings`
	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error loading settings: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err == nil {
				settingsMap[key] = value
			}
		}
	}

	// Convert map to Settings struct
	settings := &types.Settings{
		ClubName:   settingsMap["club_name"],
		Address:    settingsMap["club_address"],
		Phone:      settingsMap["club_phone"],
		Email:      settingsMap["club_email"],
		Website:    settingsMap["club_website"],
		Currency:   settingsMap["currency"],
		DateFormat: settingsMap["date_format"],
	}

	pages.Settings(user, settings).Render(r.Context(), w)
}

// SettingsUpdate handles the settings form submission
func SettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	clubName := r.FormValue("club_name")
	clubEmail := r.FormValue("club_email")
	clubPhone := r.FormValue("club_phone")
	clubWebsite := r.FormValue("club_website")

	updateSetting := func(key, value string) error {
		query := `
			INSERT INTO settings (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = $2
		`
		_, err := database.DB.Exec(query, key, value)
		return err
	}

	if clubName != "" {
		if err := updateSetting("club_name", clubName); err != nil {
			logger.Error("failed to update club_name", "error", err, "user", user.Username)
		}
	}
	if clubEmail != "" {
		if err := updateSetting("club_email", clubEmail); err != nil {
			logger.Error("failed to update club_email", "error", err, "user", user.Username)
		}
	}
	if clubPhone != "" {
		if err := updateSetting("club_phone", clubPhone); err != nil {
			logger.Error("failed to update club_phone", "error", err, "user", user.Username)
		}
	}
	if clubWebsite != "" {
		if err := updateSetting("club_website", clubWebsite); err != nil {
			logger.Error("failed to update club_website", "error", err, "user", user.Username)
		}
	}

	logger.Info("settings updated", "user", user.Username)

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
