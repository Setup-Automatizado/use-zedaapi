package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

func Apply(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

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
		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
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

	// Validate critical tables exist after migrations
	if err := validateSchema(ctx, conn, logger); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// parseGooseMigration extracts the Up section from a goose-formatted migration file.
// If no goose directives are found, returns the entire content for backward compatibility.
func parseGooseMigration(content []byte) (string, error) {
	text := string(content)

	upMarker := "-- +goose Up"
	downMarker := "-- +goose Down"

	// Find Up marker
	upIdx := strings.Index(text, upMarker)
	if upIdx == -1 {
		// No goose directives, return entire content (backward compatibility)
		return text, nil
	}

	// Skip to content after the Up marker line
	startIdx := upIdx + len(upMarker)
	if newlineIdx := strings.Index(text[startIdx:], "\n"); newlineIdx != -1 {
		startIdx += newlineIdx + 1
	}

	// Find Down marker to determine end of Up section
	downIdx := strings.Index(text[startIdx:], downMarker)
	if downIdx == -1 {
		// No Down section, use rest of file
		return text[startIdx:], nil
	}

	// Extract only the Up section (content between Up and Down markers)
	return text[startIdx : startIdx+downIdx], nil
}

// validateSchema checks that critical tables exist after migrations
func validateSchema(ctx context.Context, conn *pgxpool.Conn, logger *slog.Logger) error {
	// All tables created by 000001_init.sql migration
	criticalTables := []string{
		"instances",
		"webhook_configs",
		"webhook_outbox",
		"webhook_dlq",
		"instance_events_log",
		"event_outbox",
		"event_dlq",
		"media_metadata",
		"instance_event_sequence",
	}

	var missingTables []string
	for _, table := range criticalTables {
		var exists bool
		err := conn.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = $1
			)`, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("check table %s: %w", table, err)
		}

		if !exists {
			missingTables = append(missingTables, table)
		}
	}

	if len(missingTables) > 0 {
		if logger != nil {
			logger.Error("schema validation failed - tables missing",
				slog.Any("missing_tables", missingTables),
				slog.String("hint", "Database may be in inconsistent state. Consider dropping schema_migrations table to force re-run."))
		}
		return fmt.Errorf("critical tables missing: %v", missingTables)
	}

	if logger != nil {
		logger.Info("schema validation passed",
			slog.Int("tables_validated", len(criticalTables)))
	}

	return nil
}
