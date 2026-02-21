-- Settings table for club configuration
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO settings (key, value) VALUES
    ('club_name', 'My Kickboxing Club'),
    ('address', ''),
    ('phone', ''),
    ('email', ''),
    ('website', ''),
    ('currency', 'GBP'),
    ('date_format', '02/01/2006')
ON CONFLICT (key) DO NOTHING;
