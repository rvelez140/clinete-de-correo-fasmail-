package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemConfigRepository struct {
	pool *pgxpool.Pool
}

func NewSystemConfigRepository(pool *pgxpool.Pool) *SystemConfigRepository {
	return &SystemConfigRepository{pool: pool}
}

func (r *SystemConfigRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx,
		`SELECT value FROM system_config WHERE key = $1`, key,
	).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("config key not found: %s", key)
		}
		return "", fmt.Errorf("get config: %w", err)
	}
	return value, nil
}

func (r *SystemConfigRepository) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO system_config (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}
	return nil
}

func (r *SystemConfigRepository) IsInstalled(ctx context.Context) (bool, error) {
	value, err := r.Get(ctx, "installed")
	if err != nil {
		return false, nil
	}
	return value == "true", nil
}

func (r *SystemConfigRepository) MarkInstalled(ctx context.Context) error {
	if err := r.Set(ctx, "installed", "true"); err != nil {
		return err
	}
	return r.Set(ctx, "installed_at", time.Now().Format(time.RFC3339))
}

func (r *SystemConfigRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value FROM system_config ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("list config: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		result[k] = v
	}

	return result, nil
}
