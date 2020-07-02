CREATE TABLE IF NOT EXISTS velocities(
  id serial PRIMARY KEY,
  day TIMESTAMP,
  score integer,
  creator_id VARCHAR(64)
);

CREATE INDEX creator_idx ON velocities (creator_id);
