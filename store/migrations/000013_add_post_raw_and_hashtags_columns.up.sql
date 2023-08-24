ALTER TABLE candidate_posts ADD COLUMN raw JSONB;
ALTER TABLE candidate_posts ADD COLUMN hashtags TEXT [] DEFAULT ARRAY[]::TEXT [] NOT NULL;
CREATE INDEX candidate_posts_hashtags_idx ON candidate_posts USING gin (hashtags);
ALTER TABLE candidate_posts ADD COLUMN has_media BOOLEAN;
