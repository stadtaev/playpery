-- +goose Up
INSERT INTO admins (email, password_hash) VALUES (
    'admin@playperu.com',
    '$2a$10$trCdqP4npsbw0R1vQxVwXeT1HebzRmP01SXaNGPz1eSAZ7mpcL0Uu'
);

-- +goose Down
DELETE FROM admins WHERE email = 'admin@playperu.com';
