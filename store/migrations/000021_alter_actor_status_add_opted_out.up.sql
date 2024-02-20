ALTER TABLE candidate_actors ALTER COLUMN status DROP DEFAULT;
ALTER TABLE candidate_actors ALTER COLUMN status TYPE VARCHAR(255);
DROP TYPE actor_status;
CREATE TYPE actor_status AS ENUM ('none', 'pending', 'approved', 'banned', 'opted_out', 'rejected');
ALTER TABLE candidate_actors ALTER COLUMN status TYPE actor_status USING (status::actor_status);
ALTER TABLE candidate_actors ALTER COLUMN status SET DEFAULT 'none'::actor_status;
