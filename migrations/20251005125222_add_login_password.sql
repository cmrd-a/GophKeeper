-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS login_password
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login    text NOT NULL,
    password bytea NOT NULL,
    user_id  UUID NOT NULL REFERENCES "user" (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS login_password_user_id_uindex ON login_password (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS login_password;
-- +goose StatementEnd
