CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL CHECK (scope IN ('activation', 'refresh'))
);

CREATE INDEX IF NOT EXISTS tokens_user_id_scope_idx ON tokens (user_id, scope);
