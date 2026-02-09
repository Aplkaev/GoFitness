package user

import (
    "gofitness/src/model"
    "gofitness/src/repository"
)

type UserService struct {
    repos repository.UserRepository
}

func NewUserUserService(repos repository.UserRepository) *UserService {
    return &UserService{repos: repos}
}

func (s *UserService) GetOrCreateUser(chatID int64, username string) (*model.User, error) {
    user, err := s.repos.GetByChatID(chatID)

    if err != nil {
        return nil, err
    }

    if user != nil {
        return user, nil
    }

    return s.repos.Create(chatID, username)
}