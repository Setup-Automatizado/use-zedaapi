package migrations

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

// Apply executes all pending SQL migrations in order
func Apply(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Ensure schema_migrations table exists
	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	// Read all migration files
	entries, err := files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	var applied, skipped int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")

		// Check if migration already applied
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}

		if exists {
			skipped++
			if logger != nil {
				logger.Debug("migration already applied", slog.String("version", version))
			}
			continue
		}

		// Read migration file
		contents, err := files.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if logger != nil {
			logger.Info("applying migration", slog.String("version", version))
		}

		// Parse goose directives to extract only the Up section
		sql, err := parseGooseMigration(contents)
		if err != nil {
			return fmt.Errorf("parse migration %s: %w", entry.Name(), err)
		}

		// Execute migration
		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}

		// Record migration as applied
		if _, err := conn.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}

		applied++
	}

	if logger != nil {
		logger.Info("migrations completed",
			slog.Int("applied", applied),
			slog.Int("skipped", skipped),
			slog.Int("total", applied+skipped))
	}

	// Validate schema after migrations
	if err := validateSchema(ctx, conn, logger); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// parseGooseMigration extracts the Up section from goose-style migrations
func parseGooseMigration(content []byte) (string, error) {
	text := string(content)

	upMarker := "-- +goose Up"
	downMarker := "-- +goose Down"

	upIdx := strings.Index(text, upMarker)
	if upIdx == -1 {
		// No goose markers, return entire content
		return text, nil
	}

	// Find start of Up section (after marker and newline)
	startIdx := upIdx + len(upMarker)
	if newlineIdx := strings.Index(text[startIdx:], "\n"); newlineIdx != -1 {
		startIdx += newlineIdx + 1
	}

	// Find end of Up section (before Down marker)
	downIdx := strings.Index(text[startIdx:], downMarker)
	if downIdx == -1 {
		// No Down marker, return everything after Up marker
		return text[startIdx:], nil
	}

	return text[startIdx : startIdx+downIdx], nil
}

// validateSchema ensures all required tables exist after migrations
func validateSchema(ctx context.Context, conn *pgxpool.Conn, logger *slog.Logger) error {
	// Required tables created by 000001_init.sql
	requiredTables := []string{
		"instances",
		"webhook_configs",
		"event_outbox",
		"event_dlq",
		"media_metadata",
		"instance_event_sequence",
		"message_queue",
		"message_dlq",
		"worker_sessions",
	}

	// Check all required tables exist
	var missingTables []string
	for _, table := range requiredTables {
		var exists bool
		err := conn.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("check table %s: %w", table, err)
		}

		if !exists {
			missingTables = append(missingTables, table)
		}
	}

	// Fail if any required tables are missing
	if len(missingTables) > 0 {
		if logger != nil {
			logger.Error("schema validation failed - required tables missing",
				slog.Any("missing_tables", missingTables),
				slog.String("hint", "Database may be inconsistent. Drop schema_migrations and re-run migrations."))
		}
		return fmt.Errorf("required tables missing: %v", missingTables)
	}

	if logger != nil {
		logger.Info("schema validation passed",
			slog.Int("validated_tables", len(requiredTables)))
	}

	return nil
}
