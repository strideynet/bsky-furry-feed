ALTER TABLE candidate_posts DROP COLUMN tags;
ALTER TABLE candidate_posts DROP COLUMN is_nsfw;
ALTER TABLE candidate_actors DROP COLUMN is_nsfw;
ALTER TABLE candidate_actors DROP COLUMN is_hidden;
DELETE FROM candidate_posts WHERE raw IS NULL;
ALTER TABLE candidate_posts ALTER COLUMN raw SET NOT NULL;