package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli-auth/internal/auth"
	"github.com/cli-auth/internal/db"
	"github.com/cli-auth/internal/lockout"
	"github.com/cli-auth/internal/session"
	"github.com/cli-auth/internal/tui"
	"github.com/cli-auth/internal/user"
)

func main() {
	// Load configuration from environment
	dbCfg := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "cliauth"),
		Password: getEnv("DB_PASSWORD", "cliauth_secret"),
		DBName:   getEnv("DB_NAME", "cliauth"),
	}

	sessionTimeout := getEnvDuration("SESSION_TIMEOUT", 30*time.Minute)
	sessionFilePath := getEnv("SESSION_FILE_PATH", "/data/.session")

	// Connect to database
	fmt.Println("Connecting to database...")
	database, err := db.Connect(dbCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("Connected successfully!")

	// Initialize stores and services
	userStore := user.NewStore(database)
	sessionStore := session.NewStore(database, sessionTimeout, sessionFilePath)
	lockoutSvc := lockout.NewService(userStore)
	authSvc := auth.NewService(userStore, sessionStore, lockoutSvc)

	// Create and run the TUI
	model := tui.NewModel(authSvc, sessionStore)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

//sdfasdfasdfa
