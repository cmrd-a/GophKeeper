-- +goose Up
-- +goose StatementBegin
INSERT INTO "user" (id, login, password) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'john', '$2a$10$ItR6Olzk67ciGmeachWW4OOxapV5Vbyx241ADWcW71nEK75iRKVU.'),
    ('550e8400-e29b-41d4-a716-446655440002', 'jane', '$2a$10$ItR6Olzk67ciGmeachWW4OOxapV5Vbyx241ADWcW71nEK75iRKVU.'),
    ('550e8400-e29b-41d4-a716-446655440003', 'bob', '$2a$10$ItR6Olzk67ciGmeachWW4OOxapV5Vbyx241ADWcW71nEK75iRKVU.')
ON CONFLICT (login) DO NOTHING;

INSERT INTO text_data (id, user_id, text, created_at, updated_at) VALUES
    ('750e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'My important project notes', NOW(), NOW()),
    ('750e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', 'Shopping list items', NOW(), NOW()),
    ('750e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440003', 'WiFi network passwords', NOW(), NOW());

INSERT INTO binary_data (id, user_id, data, created_at, updated_at) VALUES
    ('850e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '\x89504e470d0a1a0a', NOW(), NOW()),
    ('850e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', '\x504b03040a000000', NOW(), NOW());

INSERT INTO card_data (id, user_id, number, cvv, holder, expires, created_at, updated_at) VALUES
    ('950e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '\x656e637279707465645f34353336', '\x656e637279707465645f313233', 'JOHN DOE', '2027-12-31', NOW(), NOW()),
    ('950e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', '\x656e637279707465645f35353535', '\x656e637279707465645f343536', 'JANE SMITH', '2026-08-31', NOW(), NOW());

INSERT INTO meta (id, relation, name, data, created_at, updated_at) VALUES
    ('a50e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440001', 'category', 'Personal Notes', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440002', '750e8400-e29b-41d4-a716-446655440002', 'category', 'Shopping', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440003', '750e8400-e29b-41d4-a716-446655440003', 'category', 'Network', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440004', '850e8400-e29b-41d4-a716-446655440001', 'file_type', 'image/png', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440005', '850e8400-e29b-41d4-a716-446655440002', 'file_type', 'application/zip', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440006', '950e8400-e29b-41d4-a716-446655440001', 'bank', 'Chase Bank', NOW(), NOW()),
    ('a50e8400-e29b-41d4-a716-446655440007', '950e8400-e29b-41d4-a716-446655440002', 'bank', 'Bank of America', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM meta WHERE id IN (
    'a50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440003',
    'a50e8400-e29b-41d4-a716-446655440004', 'a50e8400-e29b-41d4-a716-446655440005', 'a50e8400-e29b-41d4-a716-446655440006',
    'a50e8400-e29b-41d4-a716-446655440007'
);

DELETE FROM card_data WHERE id IN (
    '950e8400-e29b-41d4-a716-446655440001', '950e8400-e29b-41d4-a716-446655440002'
);

DELETE FROM binary_data WHERE id IN (
    '850e8400-e29b-41d4-a716-446655440001', '850e8400-e29b-41d4-a716-446655440002'
);

DELETE FROM text_data WHERE id IN (
    '750e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440002', '750e8400-e29b-41d4-a716-446655440003'
);

DELETE FROM "user" WHERE id IN (
    '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003'
);
-- +goose StatementEnd
