-- +goose Up
CREATE TABLE system_config (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_config (key, value) VALUES ('installed', 'false');
INSERT INTO system_config (key, value) VALUES ('installed_at', '');
INSERT INTO system_config (key, value) VALUES ('app_version', '1.0.0');

-- +goose Down
DROP TABLE IF EXISTS system_config;
