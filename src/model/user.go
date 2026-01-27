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

type ProgressPoint struct {
    Date       time.Time `json:"date"`
    TotalVolume float64   `json:"total_volume"` // вес × повторения × подходы (или просто вес × повторения)
    AvgWeight  float64   `json:"avg_weight"`
    AvgReps    float64   `json:"avg_reps"`
    SetsCount  int       `json:"sets_count"`
}