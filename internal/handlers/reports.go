package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"log"
	"net/http"
	"time"
)

// Reports shows the reports dashboard
func Reports(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pages.Reports(user).Render(r.Context(), w)
}

// AttendanceReport generates attendance report
func AttendanceReport(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// TODO: Calculate actual stats from attendance records
	stats := &types.AttendanceReportStats{
		TotalRecords:   0,
		TotalPresent:   0,
		TotalAbsent:    0,
		AttendanceRate: 0,
		TopClasses:     []types.ChartDataItem{},
		TopStudents:    []types.ChartDataItem{},
		DailyTrend:     []types.TrendDataItem{},
	}

	classes, _ := getActiveClasses()
	pages.AttendanceReport(user, stats, classes, startDate, endDate).Render(r.Context(), w)
}

// FinancialReport generates financial report
func FinancialReport(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	query := `
		SELECT id, type, amount, date, description, payment_method, approved_by
		FROM transactions
		WHERE date >= $1 AND date <= $2 AND deleted_at IS NULL
		ORDER BY date DESC
	`

	rows, err := database.DB.Query(query, startDate, endDate)
	if err != nil {
		log.Printf("Error querying transactions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var totalIncome, totalExpenditure float64

	for rows.Next() {
		var t types.Transaction
		err := rows.Scan(&t.ID, &t.Type, &t.Amount, &t.Date, &t.Description, &t.PaymentMethod, &t.ApprovedBy)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		if t.Type == "income" {
			totalIncome += t.Amount
		} else {
			totalExpenditure += t.Amount
		}
	}

	// TODO: Calculate actual category breakdowns and trends
	stats := &types.FinancialStats{
		TotalIncome:           totalIncome,
		TotalExpenditure:      totalExpenditure,
		NetBalance:            totalIncome - totalExpenditure,
		IncomeByCategory:      []types.ChartDataItem{},
		ExpenditureByCategory: []types.ChartDataItem{},
		MonthlyTrend:          []types.FinancialTrendItem{},
	}

	pages.FinancialReport(user, stats, startDate, endDate).Render(r.Context(), w)
}
