package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync/atomic"

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

// DB provides low-level access to the Things 3 SQLite database.
type DB struct {
	sqlDB      *sql.DB
	filepath   string
	printSQL   bool
	queryCount atomic.Int64
}

// Open creates a new Things 3 database connection.
func Open(opts ...Option) (*DB, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	// Discover database path
	fp, err := discoverDatabasePath(options.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Open database connection
	sqlDB, err := openDatabase(fp)
	if err != nil {
		return nil, err
	}

	// Validate database version
	if err := validateDatabaseVersion(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &DB{
		sqlDB:    sqlDB,
		filepath: fp,
		printSQL: options.PrintSQL,
	}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d.sqlDB != nil {
		return d.sqlDB.Close()
	}
	return nil
}

// Filepath returns the path to the Things database file.
func (d *DB) Filepath() string {
	return d.filepath
}

// ExecuteQuery executes a SQL query and returns the results.
func (d *DB) ExecuteQuery(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d.printSQL {
		n := d.queryCount.Add(1)
		fmt.Printf("/* Query %d */\n", n)
		if len(args) > 0 {
			fmt.Printf("/* Parameters: %v */\n", args)
		}
		fmt.Println()
		fmt.Println(query)
		fmt.Println()
	}

	return d.sqlDB.QueryContext(ctx, query, args...)
}

// ExecuteQueryRow executes a SQL query that returns a single row.
func (d *DB) ExecuteQueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if d.printSQL {
		n := d.queryCount.Add(1)
		fmt.Printf("/* Query %d */\n", n)
		if len(args) > 0 {
			fmt.Printf("/* Parameters: %v */\n", args)
		}
		fmt.Println()
		fmt.Println(query)
		fmt.Println()
	}

	return d.sqlDB.QueryRowContext(ctx, query, args...)
}

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
	if envPath := os.Getenv(EnvDatabasePath); envPath != "" {
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
	sqlDB, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test the connection
	if err := sqlDB.PingContext(context.Background()); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return sqlDB, nil
}

// getDatabaseVersion retrieves the Things database version.
func getDatabaseVersion(sqlDB *sql.DB) (int, error) {
	var plistValue string
	query := fmt.Sprintf("SELECT value FROM %s WHERE key = 'databaseVersion'", tableMeta)
	if err := sqlDB.QueryRowContext(context.Background(), query).Scan(&plistValue); err != nil {
		return 0, fmt.Errorf("get database version: %w", err)
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

	return 0, fmt.Errorf("parse database version from %q", plistValue)
}

// validateDatabaseVersion checks if the database version is supported.
func validateDatabaseVersion(sqlDB *sql.DB) error {
	version, err := getDatabaseVersion(sqlDB)
	if err != nil {
		return err
	}

	if version <= minDatabaseVersion {
		return fmt.Errorf("%w: got version %d", ErrDatabaseVersionTooOld, version)
	}

	return nil
}
