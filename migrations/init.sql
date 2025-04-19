CREATE TABLE IF NOT EXISTS users (
    guid uuid PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(guid),
    token VARCHAR(255) NOT NULL,
    jti uuid NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tokens_jti ON tokens(jti);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON users(guid);