CREATE TABLE post_hotness (
    uri TEXT NOT NULL,
    alg TEXT NOT NULL,
    generation_seq BIGINT NOT NULL,
    score REAL NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (uri, alg, generation_seq)
);
CREATE INDEX post_hotness_alg_generation_seq_score_uri_idx ON post_hotness (
    alg, generation_seq, score DESC, uri DESC
);
CREATE INDEX post_hotness_generation_at_idx ON post_hotness (generated_at);
CREATE SEQUENCE post_hotness_generation_seq;
