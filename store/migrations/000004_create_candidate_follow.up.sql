CREATE TABLE candidate_follows (
    uri TEXT PRIMARY KEY,
    actor_did TEXT NOT NULL REFERENCES candidate_actors (did),
    subject_did TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    indexed_at TIMESTAMPTZ NOT NULL
);
