package repository

import "gofitness/src/model"


type ExerciseRepository interface {
	GetByID(id int) (*model.Exercise, error)
    GetByName(name string) (*model.Exercise, error)
	GetExercises() ([]model.Exercise, error)
    
    CreateUserExercise(userID int64, name, description string, is_standard bool) (*model.Exercise, error)
}

type UserRepository interface {
	GetByChatID(chatID int64) (*model.User, error)
	Create(chatID int64, username string) (*model.User, error)
}

type WorkoutRepository interface {
	GetByID(id int) (*model.WorkoutSet, error)
	CreateWorkoutSet(set model.WorkoutSet) (*model.WorkoutSet, error)
	GetListWorkoutSetsByUserID(userID int64) ([]model.WorkoutSet, error)
	GetProgressByExercise(userID int64, exerciseID int, countList int) ([]model.ProgressPoint, error)
	SaveWorkoutSet(userID int64, exerciseID int, weight float64, reps int) error
}
