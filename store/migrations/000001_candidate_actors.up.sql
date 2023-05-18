CREATE TABLE candidate_actors (
    did TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    is_artist BOOL NOT NULL,
    comment TEXT NOT NULL
);
