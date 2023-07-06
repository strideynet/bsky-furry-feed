ALTER TABLE candidate_likes ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE candidate_posts ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE candidate_follows ADD COLUMN deleted_at TIMESTAMPTZ;
