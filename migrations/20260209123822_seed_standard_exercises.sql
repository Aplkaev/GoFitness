-- +goose Up

INSERT INTO exercises (name, description, is_standard, user_id, created_at)
VALUES
    ('Приседания',          'Приседания со штангой',                    true, NULL, CURRENT_TIMESTAMP),
    ('Жим лежа',            'Жим штанги лежа',                          true, NULL, CURRENT_TIMESTAMP),
    ('Становая тяга',       'Классическая становая тяга',              true, NULL, CURRENT_TIMESTAMP),
    ('Подтягивания',        'Подтягивания широким хватом',             true, NULL, CURRENT_TIMESTAMP),
    ('Отжимания',           'Отжимания от пола',                       true, NULL, CURRENT_TIMESTAMP),
    ('Жим стоя',            'Армейский жим',                           true, NULL, CURRENT_TIMESTAMP),
    ('Тяга штанги',         'Тяга штанги в наклоне',                   true, NULL, CURRENT_TIMESTAMP),
    ('Бицепс',              'Подъем штанги на бицепс',                 true, NULL, CURRENT_TIMESTAMP),
    ('Трицепс',             'Жим лежа узким хватом',                   true, NULL, CURRENT_TIMESTAMP),
    ('Планка',              'Упражнение на пресс',                     true, NULL, CURRENT_TIMESTAMP)
ON CONFLICT (name) WHERE user_id IS NULL DO NOTHING;

-- +goose Down

DELETE FROM exercises
WHERE is_standard = true
  AND user_id IS NULL;