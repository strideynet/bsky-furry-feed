ALTER TABLE candidate_posts ADD COLUMN tags TEXT [] DEFAULT ARRAY[]::TEXT [] NOT NULL;
ALTER TABLE candidate_posts ADD COLUMN is_nsfw BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE candidate_actors ADD COLUMN is_nsfw BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE candidate_actors ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE candidate_posts ALTER COLUMN raw DROP NOT NULL;
