ALTER TABLE candidate_posts ADD COLUMN self_labels TEXT [] DEFAULT ARRAY[]::TEXT [] NOT NULL;
ALTER TABLE actor_profiles ADD COLUMN self_labels TEXT [] DEFAULT ARRAY[]::TEXT [] NOT NULL;
CREATE INDEX candidate_posts_self_labels_idx ON candidate_posts USING gin (self_labels);
CREATE INDEX actor_profiles_self_labels_idx ON actor_profiles USING gin (self_labels);
