-- +goose Up
-- +goose StatementBegin
CREATE TABLE pull_requests (
    pull_request_id     TEXT PRIMARY KEY,
    author_id            TEXT NOT NULL REFERENCES users(user_id),
    title               TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_merged           BOOLEAN NOT NULL DEFAULT FALSE,
    merged_at         TIMESTAMP NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pull_requests;
-- +goose StatementEnd
