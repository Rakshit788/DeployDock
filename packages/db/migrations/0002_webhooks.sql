-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS github_access_token TEXT;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS webhook_secret TEXT;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS github_hook_id BIGINT;

-- +goose Down

ALTER TABLE projects DROP COLUMN IF EXISTS github_hook_id;
ALTER TABLE projects DROP COLUMN IF EXISTS webhook_secret;
ALTER TABLE users DROP COLUMN IF EXISTS github_access_token;
