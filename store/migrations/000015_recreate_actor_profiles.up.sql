ALTER TABLE candidate_actors DROP COLUMN current_profile_id;
DROP TABLE actor_profiles;

CREATE TABLE actor_profiles (
    actor_did TEXT NOT NULL REFERENCES candidate_actors (did) ON UPDATE CASCADE ON DELETE CASCADE,
    commit_cid TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    indexed_at TIMESTAMPTZ NOT NULL,
    display_name TEXT,
    description TEXT,
    PRIMARY KEY (actor_did, commit_cid)
);

CREATE INDEX actor_profiles_created_at_idx ON actor_profiles (created_at);
ALTER TABLE candidate_actors ADD COLUMN current_profile_commit_cid TEXT;
ALTER TABLE candidate_actors ADD FOREIGN KEY (did, current_profile_commit_cid) REFERENCES actor_profiles (actor_did, commit_cid) ON UPDATE CASCADE ON DELETE SET NULL;
