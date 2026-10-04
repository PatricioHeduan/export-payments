package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

// Sheet names configured directly in code (not environment variables)
const (
	IncomeSheetName  = "Ingresos"
	ExpenseSheetName = "Gastos"
)

// Config holds all service configuration loaded from environment variables and code constants.
type Config struct {
	// MercadoPago configuration
	MPAccessToken string
	MPUserID      int64
	MPBaseURL     string
	MPFetchLimit  int

	// Google Sheets configuration
	GoogleCredentialsJSON string
	GoogleCredentialsPath string
	SpreadsheetID         string
	IncomeSheetName       string
	ExpenseSheetName      string

	// Scheduler configuration
	CronSchedule string
}

// Load reads all required and optional configurations directly from environment variables.
// It also checks for an optional local .env file without requiring external libraries.
func Load() (*Config, error) {
	_ = loadDotEnvIfExists(".env")

	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	cfg := &Config{
		MPAccessToken:         accessToken,
		MPUserID:              extractUserIDFromToken(accessToken),
		MPBaseURL:             getEnv("MP_BASE_URL", "https://api.mercadopago.com"),
		MPFetchLimit:          getEnvAsInt("MP_FETCH_LIMIT", 50),
		GoogleCredentialsJSON: os.Getenv("GOOGLE_CREDENTIALS_JSON"),
		SpreadsheetID:         os.Getenv("GOOGLE_SPREADSHEET_ID"),
		IncomeSheetName:       IncomeSheetName,
		ExpenseSheetName:      ExpenseSheetName,
		CronSchedule:          getEnv("CRON_SCHEDULE", "0 */1 * * *"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// extractUserIDFromToken extracts the account owner's ID from the last segment of the MercadoPago token.
// Tokens follow the format: [PREFIX]-[CLIENT_ID]-[TIMESTAMP]-[HASH]-[USER_ID]
func extractUserIDFromToken(token string) int64 {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, "-")
	if len(parts) < 2 {
		return 0
	}
	lastPart := parts[len(parts)-1]
	userID, err := strconv.ParseInt(lastPart, 10, 64)
	if err != nil {
		return 0
	}
	return userID
}

// loadDotEnvIfExists parses a local .env file using only the standard library if present.
func loadDotEnvIfExists(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return nil // File not found, ignore silently (standard in CI/CD)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// Validate checks that all required configuration values are provided.
func (c *Config) Validate() error {
	if c.MPAccessToken == "" {
		return errors.New("missing required environment variable: MP_ACCESS_TOKEN")
	}

	if c.SpreadsheetID == "" {
		return errors.New("missing required environment variable: SPREADSHEET_ID")
	}

	if c.GoogleCredentialsJSON == "" && c.GoogleCredentialsPath == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return errors.New("missing Google credentials: provide GOOGLE_CREDENTIALS_JSON or GOOGLE_CREDENTIALS_PATH (or GOOGLE_APPLICATION_CREDENTIALS)")
	}

	if c.CronSchedule == "" {
		return errors.New("missing required environment variable: CRON_SCHEDULE")
	}

	return nil
}

// getEnv retrieves an environment variable or falls back to a default value.
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or falls back to a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}
