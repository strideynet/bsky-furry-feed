CREATE TABLE audit_events (
   id CHAR(20) PRIMARY KEY,
   actor_did TEXT NOT NULL REFERENCES candidate_actors (did),
   subject_did TEXT NOT NULL,
   subject_record_uri TEXT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL,

   payload JSON NOT NULL
);