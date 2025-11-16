-- +goose Up
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id
    ON pr_reviewers (user_id);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;
-- +goose StatementEnd
