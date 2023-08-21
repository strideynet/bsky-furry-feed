DROP INDEX candidate_posts_self_labels_idx;
DROP INDEX actor_profiles_self_labels_idx;

ALTER TABLE candidate_posts DROP COLUMN self_labels;
ALTER TABLE actor_profiles DROP COLUMN self_labels;
