CREATE TABLE visits (
  id BIGSERIAL PRIMARY KEY,
  link_id VARCHAR(255) NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  visited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip INET,
  ua TEXT,
  ref TEXT
);

CREATE INDEX idx_link_visits_link_id ON visits(link_id);
CREATE INDEX idx_link_visits_visited_at ON visits(visited_at);
