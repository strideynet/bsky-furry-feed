CREATE TABLE candidate_posts (
    uri TEXT PRIMARY KEY,
    repository_did TEXT NOT NULL REFERENCES candidate_repositories (did),
    created_at TIMESTAMPTZ NOT NULL,
    indexed_at TIMESTAMPTZ NOT NULL
);