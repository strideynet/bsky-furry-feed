CREATE TABLE candidate_posts (
    uri TEXT PRIMARY KEY,
    actor_did TEXT NOT NULL REFERENCES candidate_actors (did),
    created_at TIMESTAMPTZ NOT NULL,
    indexed_at TIMESTAMPTZ NOT NULL
);
