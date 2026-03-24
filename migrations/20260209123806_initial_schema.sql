-- +goose Up

-- Пользователи Telegram
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    chat_id     BIGINT UNIQUE NOT NULL,
    username    VARCHAR(255),
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Упражнения (стандартные + пользовательские)
CREATE TABLE IF NOT EXISTS exercises (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    is_standard BOOLEAN DEFAULT TRUE,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT unique_user_exercise UNIQUE (user_id, name)
);

-- Подходы (тренировочные сеты)
CREATE TABLE IF NOT EXISTS workout_sets (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    weight      DECIMAL(10,2) DEFAULT 0.00,
    reps        INTEGER NOT NULL CHECK (reps >= 0),
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Полезные индексы для ускорения типичных запросов

CREATE INDEX IF NOT EXISTS idx_users_chat_id        ON users(chat_id);
CREATE INDEX IF NOT EXISTS idx_exercises_user_id    ON exercises(user_id);
CREATE INDEX IF NOT EXISTS idx_exercises_name       ON exercises(name);
CREATE INDEX IF NOT EXISTS idx_workout_sets_user_id ON workout_sets(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_sets_exercise_id ON workout_sets(exercise_id);
CREATE INDEX IF NOT EXISTS idx_workout_sets_created_at  ON workout_sets(created_at);

-- +goose Down

DROP TABLE IF EXISTS workout_sets CASCADE;
DROP TABLE IF EXISTS exercises CASCADE;
DROP TABLE IF EXISTS users CASCADE;