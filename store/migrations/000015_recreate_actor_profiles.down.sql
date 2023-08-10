ALTER TABLE candidate_actors DROP COLUMN current_profile_commit_cid;
DROP TABLE actor_profiles;

CREATE TABLE actor_profiles (
    id CHAR(20) NOT NULL PRIMARY KEY,
    actor_did TEXT NOT NULL REFERENCES candidate_actors (did) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    indexed_at TIMESTAMPTZ NOT NULL,
    display_name TEXT,
    description TEXT
);

CREATE INDEX actor_profiles_created_at_idx ON actor_profiles (created_at);
ALTER TABLE candidate_actors ADD COLUMN current_profile_id CHAR(20) REFERENCES actor_profiles (id) ON UPDATE CASCADE ON DELETE SET NULL;
