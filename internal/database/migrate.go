package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fasmail/panel/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func GetMigrationStatus(ctx context.Context, pool *pgxpool.Pool) ([]MigrationInfo, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("set dialect: %w", err)
	}

	current, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		current = 0
	}

	allMigrations, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("collect migrations: %w", err)
	}

	var result []MigrationInfo
	for _, m := range allMigrations {
		info := MigrationInfo{
			Version: m.Version,
			Source:  m.Source,
			Applied: m.Version <= current,
		}
		result = append(result, info)
	}

	return result, nil
}

func RollbackLast(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	return goose.DownContext(ctx, db, ".")
}

func CheckTablesExist(ctx context.Context, pool *pgxpool.Pool) bool {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'system_config'
		)
	`).Scan(&exists)

	return err == nil && exists
}

func GetDBVersion(db *sql.DB) (int64, error) {
	goose.SetBaseFS(migrations.EmbeddedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return 0, err
	}
	return goose.GetDBVersion(db)
}

type MigrationInfo struct {
	Version int64
	Source  string
	Applied bool
}
