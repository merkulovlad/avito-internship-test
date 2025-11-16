-- +goose Up
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id
    ON pull_requests (author_id);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pull_requests_author_id;
-- +goose StatementEnd
