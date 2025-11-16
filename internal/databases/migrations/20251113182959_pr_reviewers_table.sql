-- +goose Up
-- +goose StatementBegin
CREATE TABLE pr_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pull_request_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pr_reviewers CASCADE;
-- +goose StatementEnd
