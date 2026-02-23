CREATE TABLE IF NOT EXISTS links (
    code VARCHAR(16) PRIMARY KEY,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_url ON links (url);