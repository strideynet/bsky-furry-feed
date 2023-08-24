ALTER TABLE candidate_actors ADD COLUMN is_nsfw BOOL DEFAULT false NOT NULL;
ALTER TABLE candidate_actors ADD COLUMN is_hidden BOOL DEFAULT false NOT NULL;

ALTER TABLE candidate_posts ADD COLUMN is_nsfw BOOL DEFAULT false NOT NULL;
ALTER TABLE candidate_posts ADD COLUMN is_hidden BOOL DEFAULT false NOT NULL;
