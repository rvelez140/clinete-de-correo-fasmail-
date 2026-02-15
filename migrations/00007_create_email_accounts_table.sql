-- +goose Up
-- +goose StatementBegin
CREATE TABLE email_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_address VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    imap_host VARCHAR(255) NOT NULL,
    imap_port INTEGER NOT NULL DEFAULT 993,
    imap_use_tls BOOLEAN NOT NULL DEFAULT true,
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INTEGER NOT NULL DEFAULT 587,
    smtp_use_tls BOOLEAN NOT NULL DEFAULT true,
    username VARCHAR(255) NOT NULL,
    password_encrypted TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_accounts_user_id ON email_accounts(user_id);
CREATE UNIQUE INDEX idx_email_accounts_user_email ON email_accounts(user_id, email_address);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS email_accounts;
