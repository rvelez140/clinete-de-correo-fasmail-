-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN company_id UUID REFERENCES companies(id) ON DELETE SET NULL;

-- Assign all existing users to the default company
UPDATE users SET company_id = '00000000-0000-0000-0000-000000000001';

-- Promote existing admin users to super_admin
UPDATE users SET role = 'super_admin' WHERE role = 'admin';

CREATE INDEX idx_users_company_id ON users(company_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE users SET role = 'admin' WHERE role = 'super_admin';
DROP INDEX IF EXISTS idx_users_company_id;
ALTER TABLE users DROP COLUMN IF EXISTS company_id;
-- +goose StatementEnd
