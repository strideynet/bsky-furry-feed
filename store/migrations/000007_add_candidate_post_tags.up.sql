ALTER TABLE candidate_posts ADD COLUMN tags TEXT [] DEFAULT ARRAY[]::TEXT [] NOT NULL;
