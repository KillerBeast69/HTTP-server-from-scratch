-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP not null,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP not null,
    email text NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;
