CREATE TABLE shortened_urls (
    url_id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    shortcode VARCHAR(20) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER NOT NULL DEFAULT 0
);

-- Index for faster lookups by shortcode
CREATE INDEX idx_shortcode ON shortened_urls(shortcode);

-- Function to update the timestamp
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update the updated_at timestamp
CREATE TRIGGER update_shortened_urls_timestamp
BEFORE UPDATE ON shortened_urls
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();