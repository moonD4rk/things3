package things3

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Default database paths for Things 3.
// Things 3.15.16+ uses a new path pattern with ThingsData-* directory.
const (
	defaultPathPattern31616 = "~/Library/Group Containers/" +
		"JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/" +
		"Things Database.thingsdatabase/main.sqlite"
	defaultPath31516 = "~/Library/Group Containers/" +
		"JLMPQHK86H.com.culturedcode.ThingsMac/" +
		"Things Database.thingsdatabase/main.sqlite"
)

// discoverDatabasePath finds the Things database path.
// Priority: custom path > environment variable > auto-discovery.
func discoverDatabasePath(customPath string) (string, error) {
	// 1. Use custom path if provided
	if customPath != "" {
		expanded := expandPath(customPath)
		if _, err := os.Stat(expanded); err != nil {
			return "", fmt.Errorf("%w: %s", ErrDatabaseNotFound, expanded)
		}
		return expanded, nil
	}

	// 2. Check environment variable
	if envPath := os.Getenv(envDatabasePath); envPath != "" {
		expanded := expandPath(envPath)
		if _, err := os.Stat(expanded); err != nil { //nolint:gosec // user-provided database path via env var is intentional
			return "", fmt.Errorf("%w: %s", ErrDatabaseNotFound, expanded)
		}
		return expanded, nil
	}

	// 3. Try Things 3.15.16+ path pattern first
	pattern := expandPath(defaultPathPattern31616)
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		return matches[0], nil
	}

	// 4. Fall back to older path
	oldPath := expandPath(defaultPath31516)
	if _, err := os.Stat(oldPath); err == nil {
		return oldPath, nil
	}

	return "", ErrDatabaseNotFound
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if path != "" && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// openDatabase opens a read-only SQLite connection to the Things database.
func openDatabase(path string) (*sql.DB, error) {
	// Open in read-only mode with URI
	uri := fmt.Sprintf("file:%s?mode=ro", path)
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// getDatabaseVersion retrieves the Things database version.
func getDatabaseVersion(db *sql.DB) (int, error) {
	var plistValue string
	query := fmt.Sprintf("SELECT value FROM %s WHERE key = 'databaseVersion'", tableMeta)
	if err := db.QueryRowContext(context.Background(), query).Scan(&plistValue); err != nil {
		return 0, fmt.Errorf("failed to get database version: %w", err)
	}

	// Try parsing as simple integer first
	var version int
	if n, err := fmt.Sscanf(plistValue, "%d", &version); err == nil && n == 1 {
		return version, nil
	}

	// Parse as plist XML format: <integer>N</integer>
	re := regexp.MustCompile(`<integer>(\d+)</integer>`)
	matches := re.FindStringSubmatch(plistValue)
	if len(matches) == 2 {
		if _, err := fmt.Sscanf(matches[1], "%d", &version); err == nil {
			return version, nil
		}
	}

	return 0, fmt.Errorf("failed to parse database version from %q", plistValue)
}

// validateDatabaseVersion checks if the database version is supported.
func validateDatabaseVersion(db *sql.DB) error {
	version, err := getDatabaseVersion(db)
	if err != nil {
		return err
	}

	if version <= minDatabaseVersion {
		return fmt.Errorf("%w: got version %d", ErrDatabaseVersionTooOld, version)
	}

	return nil
}
