CREATE TABLE actor_profiles (
    cid BYTEA NOT NULL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    display_name TEXT,
    description TEXT
);

CREATE INDEX candidate_actors_created_at_idx ON actor_profiles (created_at);
ALTER TABLE candidate_actors ADD COLUMN current_profile_cid BYTEA REFERENCES actor_profiles (cid) ON UPDATE CASCADE ON DELETE CASCADE;
CREATE INDEX candidate_actors_current_profile_cid_idx on candidate_actors(current_profile_cid);
