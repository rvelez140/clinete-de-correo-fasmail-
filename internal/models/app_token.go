package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ServerURL string     `json:"server_url"`
	Platform  string     `json:"platform"`
	IsUsed    bool       `json:"is_used"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type AppTokenRepository struct {
	pool *pgxpool.Pool
}

func NewAppTokenRepository(pool *pgxpool.Pool) *AppTokenRepository {
	return &AppTokenRepository{pool: pool}
}

func (r *AppTokenRepository) Create(ctx context.Context, token *AppToken) error {
	query := `
		INSERT INTO app_tokens (user_id, token_hash, server_url, platform, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		token.UserID,
		token.TokenHash,
		token.ServerURL,
		token.Platform,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *AppTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*AppToken, error) {
	query := `
		SELECT id, user_id, token_hash, server_url, platform, is_used, expires_at, used_at, created_at
		FROM app_tokens WHERE token_hash = $1`

	t := &AppToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ServerURL, &t.Platform,
		&t.IsUsed, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("token no encontrado")
		}
		return nil, fmt.Errorf("get app token: %w", err)
	}

	return t, nil
}

func (r *AppTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE app_tokens SET is_used = true, used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *AppTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM app_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
