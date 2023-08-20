CREATE TABLE post_scores (
    uri TEXT NOT NULL,
    alg TEXT NOT NULL,
    generation_seq BIGINT NOT NULL,
    score REAL NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (uri, alg, generation_seq)
);
CREATE INDEX post_scores_alg_generation_seq_score_uri_idx ON post_scores (
    alg, generation_seq, score DESC, uri DESC
);
CREATE INDEX post_scores_generation_at_idx ON post_scores (generated_at);
CREATE SEQUENCE post_scores_generation_seq;
