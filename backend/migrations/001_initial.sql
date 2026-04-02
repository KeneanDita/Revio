-- ─── Extensions ─────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ─── Users ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255),
    name            VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url      TEXT,
    role            VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    github_token    TEXT,
    github_login    VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_github_login ON users (github_login);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- ─── Repositories ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS repositories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    github_id       BIGINT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    full_name       VARCHAR(512) NOT NULL,
    owner           VARCHAR(255) NOT NULL,
    description     TEXT,
    private         BOOLEAN NOT NULL DEFAULT FALSE,
    html_url        TEXT,
    last_sync_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (user_id, full_name)
);

CREATE INDEX IF NOT EXISTS idx_repositories_user_id ON repositories (user_id);
CREATE INDEX IF NOT EXISTS idx_repositories_github_id ON repositories (github_id);
CREATE INDEX IF NOT EXISTS idx_repositories_deleted_at ON repositories (deleted_at);

-- ─── Pull Requests ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pull_requests (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id             UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    github_id           BIGINT NOT NULL,
    number              INTEGER NOT NULL,
    title               TEXT NOT NULL,
    body                TEXT,
    author              VARCHAR(255) NOT NULL,
    author_avatar_url   TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'merged')),
    base_branch         VARCHAR(255),
    head_branch         VARCHAR(255),
    additions           INTEGER NOT NULL DEFAULT 0,
    deletions           INTEGER NOT NULL DEFAULT 0,
    changed_files       INTEGER NOT NULL DEFAULT 0,
    commit_count        INTEGER NOT NULL DEFAULT 0,
    comment_count       INTEGER NOT NULL DEFAULT 0,
    review_count        INTEGER NOT NULL DEFAULT 0,
    html_url            TEXT,
    merged_at           TIMESTAMPTZ,
    closed_at           TIMESTAMPTZ,
    first_review_at     TIMESTAMPTZ,
    github_created_at   TIMESTAMPTZ NOT NULL,
    github_updated_at   TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prs_repo_id ON pull_requests (repo_id);
CREATE INDEX IF NOT EXISTS idx_prs_status ON pull_requests (status);
CREATE INDEX IF NOT EXISTS idx_prs_author ON pull_requests (author);
CREATE INDEX IF NOT EXISTS idx_prs_github_id ON pull_requests (github_id);
CREATE INDEX IF NOT EXISTS idx_prs_github_created_at ON pull_requests (github_created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prs_merged_at ON pull_requests (merged_at);

-- Full-text search on PR titles
CREATE INDEX IF NOT EXISTS idx_prs_title_trgm ON pull_requests USING gin (title gin_trgm_ops);

-- ─── Reviews ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pr_id           UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    github_id       BIGINT NOT NULL UNIQUE,
    reviewer        VARCHAR(255) NOT NULL,
    reviewer_avatar TEXT,
    state           VARCHAR(30) NOT NULL CHECK (state IN ('approved', 'changes_requested', 'commented', 'dismissed', 'APPROVED', 'CHANGES_REQUESTED', 'COMMENTED', 'DISMISSED')),
    body            TEXT,
    html_url        TEXT,
    submitted_at    TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reviews_pr_id ON reviews (pr_id);
CREATE INDEX IF NOT EXISTS idx_reviews_reviewer ON reviews (reviewer);
CREATE INDEX IF NOT EXISTS idx_reviews_state ON reviews (state);
CREATE INDEX IF NOT EXISTS idx_reviews_submitted_at ON reviews (submitted_at DESC);

-- ─── Notifications ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    title       VARCHAR(500) NOT NULL,
    body        TEXT,
    link        TEXT,
    read        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications (read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at DESC);

-- ─── Sync Jobs ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sync_jobs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id     UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_repo_id ON sync_jobs (repo_id);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status ON sync_jobs (status);

-- ─── Triggers: updated_at ────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['users', 'repositories', 'pull_requests', 'reviews', 'notifications', 'sync_jobs']
    LOOP
        EXECUTE format(
            'CREATE OR REPLACE TRIGGER set_updated_at
             BEFORE UPDATE ON %I
             FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at()',
            t
        );
    END LOOP;
END $$;

-- ─── Analytics Views ─────────────────────────────────────────────────────────
CREATE OR REPLACE VIEW pr_metrics AS
SELECT
    pr.id,
    pr.repo_id,
    pr.author,
    pr.status,
    pr.github_created_at,
    pr.merged_at,
    pr.first_review_at,
    EXTRACT(EPOCH FROM (pr.merged_at - pr.github_created_at)) / 3600 AS merge_time_hours,
    EXTRACT(EPOCH FROM (pr.first_review_at - pr.github_created_at)) / 3600 AS time_to_first_review_hours,
    r.full_name AS repo_full_name,
    r.user_id
FROM pull_requests pr
JOIN repositories r ON r.id = pr.repo_id;
