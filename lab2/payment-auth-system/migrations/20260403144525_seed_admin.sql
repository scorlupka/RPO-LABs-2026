-- +goose Up
INSERT INTO users (login, password_hash, is_admin, created_at)
VALUES (
    'admin',
    '$2a$10$PIfbICSwjoBW86KXy/LV7.osadqyCwjIBZWBRrX6olGHgoCmQLiTW',
    1,
    CURRENT_TIMESTAMP
);

-- +goose Down
DELETE FROM users WHERE login = 'admin';
