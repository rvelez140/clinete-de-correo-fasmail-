-- +goose Up
-- +goose StatementBegin
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    logo_path VARCHAR(500) DEFAULT '',
    primary_color VARCHAR(7) NOT NULL DEFAULT '#2563eb',
    success_color VARCHAR(7) NOT NULL DEFAULT '#16a34a',
    danger_color VARCHAR(7) NOT NULL DEFAULT '#dc2626',
    warning_color VARCHAR(7) NOT NULL DEFAULT '#d97706',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_slug ON companies(slug);
CREATE INDEX idx_companies_is_active ON companies(is_active);

-- Insert default company for existing data
INSERT INTO companies (id, name, slug)
VALUES ('00000000-0000-0000-0000-000000000001', 'FasMail', 'default');
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS companies;
