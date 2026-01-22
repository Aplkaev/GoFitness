package model

import "time"

// Модели данных
type User struct {
	ID        int64
	ChatID    int64
	Username  string
	CreatedAt time.Time
}

type Exercise struct {
	ID          int
	Name        string
	Description string
	IsStandard  bool
	UserID      int64
	CreatedAt   time.Time
}

type WorkoutSet struct {
	ID           int
	UserID       int64
	ExerciseID   int
	Weight       float64
	Reps         int
	CreatedAt    time.Time
	ExerciseName string
}