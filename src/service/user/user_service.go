package user

import (
	"context"
	"gofitness/src/database"
)

type UserService struct {
    db *database.Postgres
}

func NewUserService(db *database.Postgres) *UserService {
    return &UserService{db: db}
}

func (s *UserService) GetUserOrCreate(ctx context.Context, chatID int64, username, firstName, lastName string) (*database.User, error) {
	var user = s.db.GetUserByChatID(chatID)

    if user != nil {
    	return user, nil
    }

    var userSave, err = s.db.SaveUser(chatID, username, firstName, lastName)
    
    if err != nil {
        return nil, err
    }

    return userSave, nil
}

func (s *UserService) GetUserStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
    // Бизнес-логика для статистики пользователя
    return map[string]interface{}{
        "total_messages": 10,
        "last_activity":  "2024-01-01",
    }, nil
}