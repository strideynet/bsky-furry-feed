CREATE TABLE jetstream_cursor (
    cursor BIGINT NOT NULL
);
CREATE UNIQUE INDEX jetstream_cursor_single_row_idx ON jetstream_cursor ((0));
