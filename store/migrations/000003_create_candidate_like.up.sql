CREATE TABLE candidate_likes (
     uri TEXT PRIMARY KEY,
     repository_did TEXT NOT NULL REFERENCES candidate_repositories (did),
     subject_uri TEXT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL,
     indexed_at TIMESTAMPTZ NOT NULL
);