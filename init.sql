-- Пользователи Telegram
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Упражнения (стандартные + пользовательские)
CREATE TABLE IF NOT EXISTS exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_standard BOOLEAN DEFAULT TRUE,
    user_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Подходы (основная таблица)
CREATE TABLE IF NOT EXISTS workout_sets (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    weight DECIMAL(10,2) DEFAULT 0,
    reps INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);