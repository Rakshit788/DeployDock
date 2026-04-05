-- +goose Up

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    github_id BIGINT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    repo_url TEXT NOT NULL,
    framework TEXT,
    subdomain TEXT UNIQUE,
    build_command TEXT DEFAULT 'npm run build',
    output_directory TEXT DEFAULT 'dist',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    url TEXT,
    commit_sha TEXT,
    logs TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_subdomain ON projects(subdomain);
CREATE INDEX idx_deployments_project_id ON deployments(project_id);
CREATE INDEX idx_deployments_status ON deployments(status);

-- +goose Down

DROP INDEX IF EXISTS idx_deployments_status;
DROP INDEX IF EXISTS idx_deployments_project_id;
DROP INDEX IF EXISTS idx_projects_subdomain;
DROP INDEX IF EXISTS idx_projects_user_id;
DROP INDEX IF EXISTS idx_users_github_id;

DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
