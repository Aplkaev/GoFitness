package service

import (
	"gofitness/src/model"
	"gofitness/src/state"

	"gopkg.in/telebot.v3"
)

type UserService interface {
	GetOrCreateUser(chatID int64, username string) (*model.User, error)
}

type WorkoutService interface {
	GetHistory(user *model.User, countList int) (string, error)
	GetUserWorkoutHistory(user *model.User, exercisId int, countList int) ([]model.ProgressPoint, error)
	HandlerStart(chatID int64, username string) (string)
	HandleTextInput(user *model.User, text string, states *state.UserState) (model.MessageAnswer, error)
}

type ExerciseService interface {
	GetExercises() (string, error)
	ShowExerciseSelection() (*telebot.ReplyMarkup, error)
}

type Services struct {
	User     UserService
	Exercise ExerciseService
	Workout  WorkoutService
}

func NewServices(
	userService UserService,
	exerciseService ExerciseService,
	workoutService WorkoutService,
) *Services {
	return &Services{
		User:     userService,
		Exercise: exerciseService,
		Workout:  workoutService,
	}
}