-- Create the donations table
CREATE TABLE donations (
    id SERIAL PRIMARY KEY,
    payment_id VARCHAR(255) NOT NULL,
    streamer_id INT NOT NULL,
    amount INT NOT NULL,
    message TEXT,
    name VARCHAR(255),
    status VARCHAR(255) NOT NULL,
    FOREIGN KEY (streamer_id) REFERENCES streamers(id)
);