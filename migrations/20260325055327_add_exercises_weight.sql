-- +goose Up
-- +goose StatementBegin
INSERT INTO exercises (name, description, is_standard, user_id, created_at)
VALUES
    ('Вес',          'Вес человек',                    true, NULL, CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
