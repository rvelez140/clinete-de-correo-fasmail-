-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    server_url VARCHAR(500) NOT NULL,
    platform VARCHAR(50) NOT NULL DEFAULT 'android' CHECK (platform IN ('android', 'windows', 'linux')),
    is_used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_tokens_token_hash ON app_tokens(token_hash);
CREATE INDEX idx_app_tokens_user_id ON app_tokens(user_id);
CREATE INDEX idx_app_tokens_expires_at ON app_tokens(expires_at);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS app_tokens;
