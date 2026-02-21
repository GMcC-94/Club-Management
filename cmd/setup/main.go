package main

import (
	"bufio"
	"club-management/internal/auth"
	"club-management/internal/database"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Connect to database
	if err := database.Connect(); err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		fmt.Println("\nMake sure PostgreSQL is running and environment variables are set:")
		fmt.Println("  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME")
		os.Exit(1)
	}
	defer database.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║  Club Management System - Initial Setup              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get username
	fmt.Print("Enter admin username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	
	if username == "" {
		fmt.Println("❌ Username cannot be empty")
		os.Exit(1)
	}

	// Get email (optional)
	fmt.Print("Enter admin email (optional, press Enter to skip): ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Get password
	fmt.Println("\nPassword requirements:")
	fmt.Println("  • At least 8 characters")
	fmt.Println("  • At least one uppercase letter")
	fmt.Println("  • At least one lowercase letter")
	fmt.Println("  • At least one number")
	fmt.Print("\nEnter admin password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	
	if password == "" {
		fmt.Println("❌ Password cannot be empty")
		os.Exit(1)
	}

	// Confirm password
	fmt.Print("Confirm password: ")
	confirmPassword, _ := reader.ReadString('\n')
	confirmPassword = strings.TrimSpace(confirmPassword)
	
	if password != confirmPassword {
		fmt.Println("❌ Passwords do not match")
		os.Exit(1)
	}

	// Create user
	fmt.Println("\n⏳ Creating admin user...")
	err := auth.CreateUser(username, password, email, "admin")
	if err != nil {
		fmt.Printf("❌ Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║  ✅ Setup Complete!                                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Printf("\nAdmin user '%s' created successfully!\n", username)
	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Start the server: cd cmd/server && go run main.go")
	fmt.Println("  2. Open your browser: http://localhost:8080")
	fmt.Printf("  3. Log in with username: %s\n", username)
	fmt.Println("\n⚠️  Important: You'll be required to change your password on first login.")
}
