-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_reviewer_limit()
RETURNS TRIGGER AS $$
DECLARE
    reviewer_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO reviewer_count
    FROM pull_request_reviewers
    WHERE pull_request_id = NEW.pull_request_id;

    IF reviewer_count >= 2 THEN
        RAISE EXCEPTION 'Pull request % already has 2 reviewers', NEW.pull_request_id
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_reviewer_limit
BEFORE INSERT ON pull_request_reviewers
FOR EACH ROW
EXECUTE FUNCTION check_reviewer_limit();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trg_check_reviewer_limit ON pull_request_reviewers;
DROP FUNCTION check_reviewer_limit();
-- +goose StatementEnd
