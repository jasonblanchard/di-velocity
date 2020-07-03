CREATE TABLE IF NOT EXISTS daily_counts(
  id serial PRIMARY KEY,
  day TIMESTAMP,
  count integer,
  creator_id VARCHAR(64)
);

CREATE INDEX creator_idx ON daily_counts (creator_id);
