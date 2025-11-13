package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

// EnsureDatabaseExists ensures that the database specified in the DSN exists.
// If the database does not exist, it will be created automatically.
//
// This function is idempotent and safe to call multiple times.
// It's designed for use in development, CI/CD, and automated deployments.
//
// Process:
//  1. Parse DSN to extract database name and connection parameters
//  2. Connect to PostgreSQL using 'postgres' maintenance database
//  3. Check if target database exists
//  4. Create database if it doesn't exist
//  5. Close maintenance connection
//
// Example:
//
//	dsn := "postgres://user:pass@localhost:5432/myapp?sslmode=disable"
//	if err := EnsureDatabaseExists(ctx, dsn, logger); err != nil {
//	    return fmt.Errorf("ensure database: %w", err)
//	}
//
// Note: Requires PostgreSQL user to have CREATEDB privilege or be a superuser.
func EnsureDatabaseExists(ctx context.Context, dsn string, logger *slog.Logger) error {
	// Parse DSN to extract database name and connection info
	dbName, maintenanceDSN, err := parseDSNForMaintenance(dsn)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}

	if logger != nil {
		logger.Debug("checking database existence",
			slog.String("database", dbName))
	}

	// Connect to PostgreSQL maintenance database
	conn, err := pgx.Connect(ctx, maintenanceDSN)
	if err != nil {
		return fmt.Errorf("connect to maintenance database: %w", err)
	}
	defer conn.Close(ctx)

	// Check if database exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	if err := conn.QueryRow(ctx, query, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}

	if exists {
		if logger != nil {
			logger.Debug("database already exists",
				slog.String("database", dbName))
		}
		return nil
	}

	// Create database
	// Note: Database names must be quoted if they contain special characters
	// We use pgx's QuoteIdentifier for safe SQL generation
	createSQL := fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{dbName}.Sanitize())

	if logger != nil {
		logger.Info("creating database",
			slog.String("database", dbName))
	}

	if _, err := conn.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}

	if logger != nil {
		logger.Info("database created successfully",
			slog.String("database", dbName))
	}

	return nil
}

// parseDSNForMaintenance extracts the database name from a DSN and returns
// a modified DSN that connects to the 'postgres' maintenance database instead.
//
// Input:  postgres://user:pass@localhost:5432/myapp?sslmode=disable
// Output: "myapp", "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
//
// This allows us to connect to PostgreSQL to create the target database.
func parseDSNForMaintenance(dsn string) (dbName string, maintenanceDSN string, err error) {
	// Parse the DSN as a URL
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("invalid DSN format: %w", err)
	}

	// Extract database name from path
	// Path format: /database_name
	dbName = strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", "", fmt.Errorf("no database name in DSN path: %s", dsn)
	}

	// Special case: if already connecting to 'postgres' or 'template1', don't modify
	if dbName == "postgres" || dbName == "template1" {
		return dbName, dsn, nil
	}

	// Replace database name with 'postgres' (the default maintenance database)
	u.Path = "/postgres"

	maintenanceDSN = u.String()

	return dbName, maintenanceDSN, nil
}

// EnsureMultipleDatabases ensures that multiple databases exist.
// This is useful when your application uses multiple databases
// (e.g., main application database + whatsmeow store database).
//
// Example:
//
//	dsns := map[string]string{
//	    "application": "postgres://user:pass@localhost:5432/myapp?sslmode=disable",
//	    "whatsmeow":   "postgres://user:pass@localhost:5432/myapp_store?sslmode=disable",
//	}
//	if err := EnsureMultipleDatabases(ctx, dsns, logger); err != nil {
//	    return fmt.Errorf("ensure databases: %w", err)
//	}
func EnsureMultipleDatabases(ctx context.Context, dsns map[string]string, logger *slog.Logger) error {
	for name, dsn := range dsns {
		if logger != nil {
			logger.Debug("ensuring database exists",
				slog.String("name", name))
		}

		if err := EnsureDatabaseExists(ctx, dsn, logger); err != nil {
			return fmt.Errorf("ensure %s database: %w", name, err)
		}
	}

	if logger != nil {
		logger.Info("all databases verified/created",
			slog.Int("count", len(dsns)))
	}

	return nil
}
