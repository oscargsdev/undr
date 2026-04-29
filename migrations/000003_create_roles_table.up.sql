CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    code text UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users_roles (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    role_id bigint NOT NULL REFERENCES roles ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Add the two roles to the table.
INSERT INTO roles (code)
VALUES 
    ('user'),
    ('admin');
