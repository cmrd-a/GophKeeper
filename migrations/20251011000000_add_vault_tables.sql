-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS text_data
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  UUID NOT NULL REFERENCES "user" (id),
    text     text NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS binary_data
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  UUID NOT NULL REFERENCES "user" (id),
    data     bytea NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS card_data
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  UUID NOT NULL REFERENCES "user" (id),
    number   bytea NOT NULL, -- encrypted
    cvv      bytea NOT NULL, -- encrypted
    holder   text NOT NULL,
    expires  date NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meta
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relation UUID NOT NULL,
    name     text NOT NULL,
    data     text NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS text_data_user_id_idx ON text_data (user_id);
CREATE INDEX IF NOT EXISTS binary_data_user_id_idx ON binary_data (user_id);
CREATE INDEX IF NOT EXISTS card_data_user_id_idx ON card_data (user_id);
CREATE INDEX IF NOT EXISTS meta_relation_idx ON meta (relation);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS meta;
DROP TABLE IF EXISTS card_data;
DROP TABLE IF EXISTS binary_data;
DROP TABLE IF EXISTS text_data;
-- +goose StatementEnd