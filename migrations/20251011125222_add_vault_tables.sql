-- +goose Up
-- +goose StatementBegin
-- TextData table
CREATE TABLE IF NOT EXISTS text_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user" (id),
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS text_data_user_id_idx ON text_data (user_id);

-- BinaryData table
CREATE TABLE IF NOT EXISTS binary_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user" (id),
    data BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS binary_data_user_id_idx ON binary_data (user_id);

-- CardData table with encrypted fields
CREATE TABLE IF NOT EXISTS card_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user" (id),
    number BYTEA NOT NULL, -- encrypted
    cvv BYTEA NOT NULL,    -- encrypted
    holder TEXT NOT NULL,
    expires TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS card_data_user_id_idx ON card_data (user_id);

-- Meta table for additional data on any vault item
CREATE TABLE IF NOT EXISTS meta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relation UUID NOT NULL, -- references any vault item
    name TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS meta_relation_idx ON meta (relation);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS meta;
DROP TABLE IF EXISTS card_data;
DROP TABLE IF EXISTS binary_data;
DROP TABLE IF EXISTS text_data;
-- +goose StatementEnd