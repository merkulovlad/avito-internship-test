-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    team_name TEXT PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE teams;
-- +goose StatementEnd
