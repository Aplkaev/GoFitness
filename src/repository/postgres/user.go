package postgres

import (
	"database/sql"
	"fmt"
	"gofitness/src/model"
)
type UserPostgres struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserPostgres {
    return &UserPostgres{db: db}
}

func (r *UserPostgres) GetByChatID(chatID int64) (*model.User, error) { 
	query := `SELECT id, chat_id, username
              FROM users WHERE chat_id = $1`
    
    var user model.User
    err := r.db.QueryRow(query, chatID).Scan(
        &user.ID, 
		&user.ChatID, 
		&user.Username, 
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса пользователя: %w", err)
    }
    
    return &user, nil
 }


// Сохраняем или получаем пользователя
func (r *UserPostgres) Create(chatID int64, username string) (*model.User, error) {
    query := `
        INSERT INTO users (chat_id, username) 
        VALUES ($1, $2) 
        ON CONFLICT (chat_id) 
        DO UPDATE SET 
            username = EXCLUDED.username
        RETURNING id, chat_id, username, created_at
    `
    
    var user model.User
    err := r.db.QueryRow(
        query, 
        chatID, 
		username,
    ).Scan(
        &user.ID,
        &user.ChatID, 
        &user.Username,
        &user.CreatedAt,
    )

    if err != nil {
        return nil, err
    }
    
    return &user, nil
}