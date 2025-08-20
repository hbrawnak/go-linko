--- Create urls table
CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    short_code varchar(10) NOT NULL UNIQUE,
    original_URL TEXT NOT NULL,
    hit_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT NOW(),
)
