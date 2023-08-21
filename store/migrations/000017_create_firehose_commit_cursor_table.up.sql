CREATE TABLE firehose_commit_cursor (
    cursor BIGINT NOT NULL
);
CREATE UNIQUE INDEX firehose_commit_cursor_single_row_idx ON firehose_commit_cursor((0));
