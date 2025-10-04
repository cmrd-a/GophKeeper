-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "user"
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login    text NOT NULL,
    password bytea NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS user_login_uindex ON "user" (login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd
