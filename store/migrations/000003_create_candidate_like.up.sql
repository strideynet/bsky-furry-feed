CREATE TABLE candidate_likes (
     uri TEXT PRIMARY KEY,
     actor_did TEXT NOT NULL REFERENCES candidate_actors (did),
     subject_uri TEXT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL,
     indexed_at TIMESTAMPTZ NOT NULL
);