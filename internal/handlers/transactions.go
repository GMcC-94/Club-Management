package handlers

import (
	"club-management/internal/auth"
	"club-management/internal/database"
	"club-management/internal/logger"
	"club-management/internal/types"
	"club-management/web/templates/pages"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// TransactionList shows all transactions
func TransactionList(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
		SELECT t.id, t.type, t.amount, t.date, t.description, t.payment_method, 
		       t.approved_by, t.receipt_filename, c.name as category_name
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.deleted_at IS NULL
		ORDER BY t.date DESC, t.id DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying transactions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var t types.Transaction
		err := rows.Scan(&t.ID, &t.Type, &t.Amount, &t.Date, &t.Description, &t.PaymentMethod, &t.ApprovedBy, &t.ReceiptFilename, &t.CategoryName)
		if err != nil {
			log.Printf("Error scanning transaction: %v", err)
			continue
		}
		transactions = append(transactions, t)
	}

	categories, err := getAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
	}

	pages.Transactions(user, transactions, categories).Render(r.Context(), w)
}

// TransactionNew shows the form for creating a new transaction
func TransactionNew(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categories, err := getAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
	}

	pages.TransactionNew(user, categories).Render(r.Context(), w)
}

// TransactionEdit shows the form for editing a transaction
func TransactionEdit(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	transactionID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	transaction, err := getTransactionByID(transactionID)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	categories, err := getAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
	}

	pages.TransactionEdit(user, transaction, categories).Render(r.Context(), w)
}

// TransactionCreate handles the form submission for creating a transaction
func TransactionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	txnType := r.FormValue("type")
	amountStr := r.FormValue("amount")
	dateStr := r.FormValue("date")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	paymentMethod := r.FormValue("payment_method")
	approvedBy := r.FormValue("approved_by")

	if txnType == "" || amountStr == "" || dateStr == "" || description == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	var categoryID sql.NullInt32
	if categoryIDStr != "" {
		if catID, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = sql.NullInt32{Int32: int32(catID), Valid: true}
		}
	}

	var receiptFilename sql.NullString
	file, header, err := r.FormFile("receipt")
	if err == nil {
		defer file.Close()

		uploadsDir := "/root/uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			logger.Error("failed to create uploads directory", "error", err)
		} else {
			filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(header.Filename))
			filepath := filepath.Join(uploadsDir, filename)

			dst, err := os.Create(filepath)
			if err != nil {
				logger.Error("failed to create receipt file", "error", err)
			} else {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err != nil {
					logger.Error("failed to save receipt", "error", err)
				} else {
					receiptFilename = sql.NullString{String: filename, Valid: true}
				}
			}
		}
	}

	query := `
		INSERT INTO transactions (type, amount, date, description, category_id, payment_method, approved_by, receipt_filename, created_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = database.DB.Exec(query, txnType, amount, date, description, categoryID, nullString(paymentMethod), nullString(approvedBy), receiptFilename, user.ID, time.Now())
	if err != nil {
		logger.Error("failed to create transaction", "error", err, "type", txnType, "amount", amount, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("transaction created", "type", txnType, "amount", amount, "user", user.Username)

	http.Redirect(w, r, "/transactions", http.StatusSeeOther)
}

// TransactionUpdate handles the form submission for updating a transaction
func TransactionUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	transactionID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	txnType := r.FormValue("type")
	amountStr := r.FormValue("amount")
	dateStr := r.FormValue("date")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	paymentMethod := r.FormValue("payment_method")
	approvedBy := r.FormValue("approved_by")

	if txnType == "" || amountStr == "" || dateStr == "" || description == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	var categoryID sql.NullInt32
	if categoryIDStr != "" {
		if catID, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = sql.NullInt32{Int32: int32(catID), Valid: true}
		}
	}

	query := `
		UPDATE transactions
		SET type = $1, amount = $2, date = $3, description = $4, category_id = $5, payment_method = $6, approved_by = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	_, err = database.DB.Exec(query, txnType, amount, date, description, categoryID, nullString(paymentMethod), nullString(approvedBy), time.Now(), transactionID)
	if err != nil {
		logger.Error("failed to update transaction", "error", err, "transaction_id", transactionID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("transaction updated", "transaction_id", transactionID, "type", txnType, "amount", amount, "user", user.Username)

	http.Redirect(w, r, "/transactions", http.StatusSeeOther)
}

// TransactionDelete handles soft-deleting a transaction
func TransactionDelete(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	transactionID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE transactions SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.DB.Exec(query, time.Now(), transactionID)
	if err != nil {
		logger.Error("failed to delete transaction", "error", err, "transaction_id", transactionID, "user", user.Username)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("transaction deleted", "transaction_id", transactionID, "user", user.Username)

	http.Redirect(w, r, "/transactions", http.StatusSeeOther)
}

// Helper function
func getTransactionByID(id int) (*types.Transaction, error) {
	query := `
		SELECT id, type, amount, date, description, category_id, payment_method, approved_by, receipt_filename
		FROM transactions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var t types.Transaction
	err := database.DB.QueryRow(query, id).Scan(&t.ID, &t.Type, &t.Amount, &t.Date, &t.Description, &t.CategoryID, &t.PaymentMethod, &t.ApprovedBy, &t.ReceiptFilename)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
