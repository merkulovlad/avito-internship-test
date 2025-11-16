-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(team_name),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
