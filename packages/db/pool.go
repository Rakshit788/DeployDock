package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

//go:embed migrations/*.sql
var migrationFiles embed.FS

func InitDB(databaseURL string) {
	var err error

	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Failed to create pool:", err)
	}

	err = Pool.Ping(context.Background())
	if err != nil {
		log.Fatal("DB not reachable:", err)
	}

	if err := RunMigrations(context.Background()); err != nil {
		log.Fatal("Database migration failed:", err)
	}

	log.Println("DB connected")
}

// RunMigrations applies each embedded SQL migration once. The advisory lock
// lets every service call this safely during concurrent startup.
func RunMigrations(ctx context.Context) error {
	if Pool == nil {
		return fmt.Errorf("database pool not initialized")
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	tx, err := Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(746381)"); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("create migration metadata table: %w", err)
	}

	for _, filename := range files {
		version := strings.TrimSuffix(filename, path.Ext(filename))
		var applied bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if applied {
			continue
		}

		sql, err := migrationFiles.ReadFile(path.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}
		upSQL := strings.SplitN(string(sql), "-- +goose Down", 2)[0]
		if _, err := tx.Exec(ctx, upSQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", filename, err)
		}
		log.Printf("Applied database migration %s", filename)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}
