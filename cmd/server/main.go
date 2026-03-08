package main

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/handlers"
	"club-management/internal/logger"
	"club-management/internal/middleware"
	"club-management/internal/repository"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
)

func main() {
	if err := database.Connect(); err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// CSRF protection
	csrfKey := []byte(os.Getenv("SESSION_SECRET"))
	if len(csrfKey) < 32 {
		logger.Error("SESSION_SECRET must be at least 32 characters")
		os.Exit(1)
	}

	csrfMiddleware := csrf.Protect(
		csrfKey,
		csrf.Secure(os.Getenv("APP_ENV") == "production"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn("CSRF token invalid", "path", r.URL.Path, "remote", r.RemoteAddr)
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		})),
	)

	// Initialise repositories
	attendanceRepo := repository.NewAttendanceRepository()

	// Initialise handlers
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo)

	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	mux.Handle("/", middleware.LoginRateLimit(http.HandlerFunc(handlers.Index)))
	mux.HandleFunc("GET /login", handlers.LoginPage)
	mux.Handle("/login", middleware.LoginRateLimit(http.HandlerFunc(handlers.Login)))
	mux.HandleFunc("/logout", handlers.Logout)

	// Protected routes - all users
	mux.HandleFunc("/dashboard", auth.RequireAuth(handlers.Dashboard))
	mux.HandleFunc("/attendance", auth.RequireAuth(handlers.AttendanceList))
	mux.HandleFunc("/attendance/take", auth.RequireAuth(handlers.TakeAttendance))
	mux.HandleFunc("/attendance/submit", auth.RequireAuth(handlers.SubmitAttendance))
	mux.HandleFunc("/attendance/students", auth.RequireAuth(attendanceHandler.LoadStudentsForClass))
	mux.HandleFunc("/attendance/students/search", auth.RequireAuth(handlers.SearchStudents))
	mux.HandleFunc("/attendance/edit", auth.RequireAuth(handlers.AttendanceEdit))
	mux.HandleFunc("/attendance/update", auth.RequireAuth(handlers.AttendanceUpdate))

	// Protected routes - admin only
	mux.HandleFunc("/students", auth.RequireAuth(auth.RequireAdmin(handlers.StudentList)))
	mux.HandleFunc("/students/new", auth.RequireAuth(auth.RequireAdmin(handlers.StudentNew)))
	mux.HandleFunc("/students/create", auth.RequireAuth(auth.RequireAdmin(handlers.StudentCreate)))
	mux.HandleFunc("/students/edit", auth.RequireAuth(auth.RequireAdmin(handlers.StudentEdit)))
	mux.HandleFunc("/students/update", auth.RequireAuth(auth.RequireAdmin(handlers.StudentUpdate)))
	mux.HandleFunc("/students/delete", auth.RequireAuth(auth.RequireAdmin(handlers.StudentDelete)))

	mux.HandleFunc("/classes", auth.RequireAuth(auth.RequireAdmin(handlers.ClassList)))
	mux.HandleFunc("/classes/new", auth.RequireAuth(auth.RequireAdmin(handlers.ClassNew)))
	mux.HandleFunc("/classes/create", auth.RequireAuth(auth.RequireAdmin(handlers.ClassCreate)))
	mux.HandleFunc("/classes/edit", auth.RequireAuth(auth.RequireAdmin(handlers.ClassEdit)))
	mux.HandleFunc("/classes/update", auth.RequireAuth(auth.RequireAdmin(handlers.ClassUpdate)))
	mux.HandleFunc("/classes/delete", auth.RequireAuth(auth.RequireAdmin(handlers.ClassDelete)))

	mux.HandleFunc("/transactions", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionList)))
	mux.HandleFunc("/transactions/new", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionNew)))
	mux.HandleFunc("/transactions/create", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionCreate)))
	mux.HandleFunc("/transactions/edit", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionEdit)))
	mux.HandleFunc("/transactions/update", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionUpdate)))
	mux.HandleFunc("/transactions/delete", auth.RequireAuth(auth.RequireTreasurer(handlers.TransactionDelete)))

	mux.HandleFunc("/categories/manage", auth.RequireAuth(auth.RequireTreasurer(handlers.CategoryList)))
	mux.HandleFunc("/categories/create", auth.RequireAuth(auth.RequireTreasurer(handlers.CategoryCreate)))
	mux.HandleFunc("/categories/delete", auth.RequireAuth(auth.RequireTreasurer(handlers.CategoryDelete)))

	mux.HandleFunc("/reports", auth.RequireAuth(auth.RequireAdmin(handlers.Reports)))
	mux.HandleFunc("/reports/attendance", auth.RequireAuth(auth.RequireAdmin(handlers.AttendanceReport)))
	mux.HandleFunc("/reports/financial", auth.RequireAuth(auth.RequireTreasurer(handlers.FinancialReport)))

	mux.HandleFunc("/users", auth.RequireAuth(auth.RequireAdmin(handlers.UserList)))
	mux.HandleFunc("/users/new", auth.RequireAuth(auth.RequireAdmin(handlers.UserNew)))
	mux.HandleFunc("/users/create", auth.RequireAuth(auth.RequireAdmin(handlers.UserCreate)))
	mux.HandleFunc("/users/edit", auth.RequireAuth(auth.RequireAdmin(handlers.UserEdit)))
	mux.HandleFunc("/users/update", auth.RequireAuth(auth.RequireAdmin(handlers.UserUpdate)))
	mux.HandleFunc("/users/delete", auth.RequireAuth(auth.RequireAdmin(handlers.UserDelete)))

	mux.HandleFunc("/settings", auth.RequireAuth(auth.RequireAdmin(handlers.SettingsPage)))
	mux.HandleFunc("/settings/update", auth.RequireAuth(auth.RequireAdmin(handlers.SettingsUpdate)))

	// Global middleware chain: CSRF → inject token into ctx q → request logging
	handler := csrfMiddleware(middleware.InjectCSRFToken(middleware.RequestLogger(mux)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	logger.Info("server starting", "addr", addr, "env", os.Getenv("APP_ENV"))
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
