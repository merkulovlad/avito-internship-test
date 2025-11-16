-- +goose Up
CREATE INDEX IF NOT EXISTS idx_users_team_name_is_active
    ON users (team_name, is_active);

CREATE INDEX IF NOT EXISTS idx_users_team_name
    ON users (team_name);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_team_name_is_active;
DROP INDEX IF EXISTS idx_users_team_name;
-- +goose StatementEnd
