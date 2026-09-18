package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	dsn string
	dir string
)

func init() {
	flag.StringVar(&dsn, "dsn", os.Getenv("DB_DSN"), "database DSN")
	flag.StringVar(&dir, "dir", envOrDefault("MIGRATIONS_DIR", "./migrations"), "migrations directory")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	flag.Parse()

	if dsn == "" {
		log.Fatal("dsn is required (flag -dsn or DB_DSN env var)")
	}

	if err := run(dsn, dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run(dsn, dir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	names, err := migrationFiles(dir)
	if err != nil {
		return err
	}

	for _, name := range names {
		applied, err := isApplied(ctx, db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, db, dir, name); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		log.Printf("applied migration %s", name)
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	return err
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

func isApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
	).Scan(&exists)
	return exists, err
}

var upSectionRe = regexp.MustCompile(`(?is)--\s*\+goose\s+Up\b(.*?)(?:--\s*\+goose\s+Down\b|\z)`)

func upStatements(data []byte) ([]string, error) {
	m := upSectionRe.FindSubmatch(data)
	if m == nil {
		return nil, fmt.Errorf("no '-- +goose Up' section found")
	}

	var statements []string
	for stmt := range strings.SplitSeq(string(m[1]), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements, nil
}

func applyMigration(ctx context.Context, db *sql.DB, dir, name string) error {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	statements, err := upStatements(data)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return err
	}
	return tx.Commit()
}
