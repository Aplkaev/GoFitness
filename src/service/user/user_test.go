package user

import (
	"errors"
	"gofitness/src/model"
	"testing"
)

type mockRepo struct {
    getUser *model.User
    getErr error

    createUser *model.User
    createErr error

    createCalled bool
}

func (m *mockRepo) GetByChatID(chatID int64) (*model.User, error) {
    return m.getUser, m.getErr
}

func (m *mockRepo) Create(chatID int64, username string) (*model.User, error) {
    m.createCalled = true
    return m.createUser, m.createErr
}

func TestGetOrCreateUser_ReturnExistingUser(t *testing.T) {

    repo := &mockRepo{
        getUser: &model.User{
            ChatID: 123,
            Username: "alex",
        },
    }

    service := NewUserUserService(repo)

    user, err := service.GetOrCreateUser(123, "alex")

    if err != nil {
        t.Fatal(err)
    }

    if user.ChatID != 123 {
        t.Fatal("wrong user")
    }

    if repo.createCalled {
        t.Fatal("Create should not be called")
    }
}

func TestGetOrCreateUser_CreateUser(t *testing.T) {

    repo := &mockRepo{

        getUser: nil,

        createUser: &model.User{
            ChatID: 10,
            Username: "john",
        },
    }

    service := NewUserUserService(repo)

    user, err := service.GetOrCreateUser(10, "john")

    if err != nil {
        t.Fatal(err)
    }

    if user == nil {
        t.Fatal("user is nil")
    }

    if !repo.createCalled {
        t.Fatal("Create wasn't called")
    }
}

func TestGetOrCreateUser_GetError(t *testing.T) {

    repo := &mockRepo{
        getErr: errors.New("db error"),
    }

    service := NewUserUserService(repo)

    _, err := service.GetOrCreateUser(1, "john")

    if err == nil {
        t.Fatal("expected error")
    }
}