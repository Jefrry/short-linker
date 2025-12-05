CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(25) NOT NULL,
    email      VARCHAR(30) UNIQUE NOT NULL,
    password   TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
