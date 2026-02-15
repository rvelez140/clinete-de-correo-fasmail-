package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailAccount struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	EmailAddress      string    `json:"email_address"`
	DisplayName       string    `json:"display_name"`
	IMAPHost          string    `json:"imap_host"`
	IMAPPort          int       `json:"imap_port"`
	IMAPUseTLS        bool      `json:"imap_use_tls"`
	SMTPHost          string    `json:"smtp_host"`
	SMTPPort          int       `json:"smtp_port"`
	SMTPUseTLS        bool      `json:"smtp_use_tls"`
	Username          string    `json:"username"`
	PasswordEncrypted string    `json:"-"`
	IsDefault         bool      `json:"is_default"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type EmailAccountRepository struct {
	pool *pgxpool.Pool
}

func NewEmailAccountRepository(pool *pgxpool.Pool) *EmailAccountRepository {
	return &EmailAccountRepository{pool: pool}
}

func (r *EmailAccountRepository) Create(ctx context.Context, acct *EmailAccount) error {
	query := `
		INSERT INTO email_accounts (user_id, email_address, display_name,
			imap_host, imap_port, imap_use_tls, smtp_host, smtp_port, smtp_use_tls,
			username, password_encrypted, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		acct.UserID, acct.EmailAddress, acct.DisplayName,
		acct.IMAPHost, acct.IMAPPort, acct.IMAPUseTLS,
		acct.SMTPHost, acct.SMTPPort, acct.SMTPUseTLS,
		acct.Username, acct.PasswordEncrypted, acct.IsDefault,
	).Scan(&acct.ID, &acct.CreatedAt, &acct.UpdatedAt)
}

func (r *EmailAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*EmailAccount, error) {
	query := `
		SELECT id, user_id, email_address, display_name,
		       imap_host, imap_port, imap_use_tls,
		       smtp_host, smtp_port, smtp_use_tls,
		       username, password_encrypted, is_default,
		       created_at, updated_at
		FROM email_accounts WHERE id = $1`

	acct := &EmailAccount{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&acct.ID, &acct.UserID, &acct.EmailAddress, &acct.DisplayName,
		&acct.IMAPHost, &acct.IMAPPort, &acct.IMAPUseTLS,
		&acct.SMTPHost, &acct.SMTPPort, &acct.SMTPUseTLS,
		&acct.Username, &acct.PasswordEncrypted, &acct.IsDefault,
		&acct.CreatedAt, &acct.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("cuenta de correo no encontrada")
		}
		return nil, fmt.Errorf("get email account: %w", err)
	}

	return acct, nil
}

func (r *EmailAccountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]EmailAccount, error) {
	query := `
		SELECT id, user_id, email_address, display_name,
		       imap_host, imap_port, imap_use_tls,
		       smtp_host, smtp_port, smtp_use_tls,
		       username, password_encrypted, is_default,
		       created_at, updated_at
		FROM email_accounts WHERE user_id = $1
		ORDER BY is_default DESC, created_at ASC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list email accounts: %w", err)
	}
	defer rows.Close()

	var accounts []EmailAccount
	for rows.Next() {
		var a EmailAccount
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.EmailAddress, &a.DisplayName,
			&a.IMAPHost, &a.IMAPPort, &a.IMAPUseTLS,
			&a.SMTPHost, &a.SMTPPort, &a.SMTPUseTLS,
			&a.Username, &a.PasswordEncrypted, &a.IsDefault,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan email account: %w", err)
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

func (r *EmailAccountRepository) Update(ctx context.Context, acct *EmailAccount) error {
	query := `
		UPDATE email_accounts SET
			email_address = $1, display_name = $2,
			imap_host = $3, imap_port = $4, imap_use_tls = $5,
			smtp_host = $6, smtp_port = $7, smtp_use_tls = $8,
			username = $9, password_encrypted = $10, is_default = $11,
			updated_at = NOW()
		WHERE id = $12 AND user_id = $13`

	tag, err := r.pool.Exec(ctx, query,
		acct.EmailAddress, acct.DisplayName,
		acct.IMAPHost, acct.IMAPPort, acct.IMAPUseTLS,
		acct.SMTPHost, acct.SMTPPort, acct.SMTPUseTLS,
		acct.Username, acct.PasswordEncrypted, acct.IsDefault,
		acct.ID, acct.UserID,
	)
	if err != nil {
		return fmt.Errorf("update email account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cuenta de correo no encontrada")
	}

	return nil
}

func (r *EmailAccountRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM email_accounts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete email account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cuenta de correo no encontrada")
	}
	return nil
}
