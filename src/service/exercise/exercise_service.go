package exercise

import (
	"fmt"
	"gofitness/src/database"
	"strings"
)

type ExerciseService struct { 
    db *database.Postgres
}

func NewExerciseService(db *database.Postgres) *ExerciseService {
    return &ExerciseService{
        db: db,
    }
}

func (s *ExerciseService) GetExercises(chatID int64) (string, error) {

    exercises, err := s.db.GetExercises()
    fmt.Println("asd")

    if err != nil { 
        fmt.Println(err)
        return "Ошибка при получение упражнений", err
    }
    fmt.Println("asd")

    var message strings.Builder
    message.WriteString("🏋️ Все упражнения:\n\n")
    fmt.Println("asd")

    for _, ex := range exercises {
            fmt.Println("asd")

        message.WriteString(fmt.Sprintf("• %s", ex.Name))
        if ex.Description != "" {
            message.WriteString(fmt.Sprintf(" - %s", ex.Description))
        }
        fmt.Println("asd")

        message.WriteString("\n")
    }
    fmt.Println("asd")

    fmt.Println(message.String())
    fmt.Println("asd")

    return message.String(), nil
}