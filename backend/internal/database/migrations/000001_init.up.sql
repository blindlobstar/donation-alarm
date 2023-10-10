CREATE TABLE IF NOT EXISTS streamers (
    id SERIAL PRIMARY KEY,
    twitch_id TEXT,
    twitch_name TEXT,
    secret_code TEXT
);