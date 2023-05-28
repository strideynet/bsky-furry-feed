CREATE TYPE actor_status AS ENUM ('none', 'pending', 'approved', 'banned');
ALTER TABLE candidate_actors ADD COLUMN status actor_status DEFAULT 'none' NOT NULL;
