CREATE TABLE IF NOT EXISTS streamers (
    id SERIAL PRIMARY KEY,
    twitch_id TEXT NOT NULL,
    twitch_name TEXT NOT NULL,
    secret_code TEXT NOT NULL 
);