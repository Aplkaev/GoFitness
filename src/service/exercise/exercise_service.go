package exercise

import (
	"fmt"
	"gofitness/src/database"
	"strings"

	"gopkg.in/telebot.v3"
)

type ExerciseService struct { 
    db *database.Postgres
}

func NewExerciseService(db *database.Postgres) *ExerciseService {
    return &ExerciseService{
        db: db,
    }
}

func (s *ExerciseService) GetExercises() (string, error) {
    exercises, err := s.db.GetExercises()

    if err != nil { 
        fmt.Println(err)
        return "Ошибка при получение упражнений", err
    }

    var message strings.Builder
    message.WriteString("🏋️ Все упражнения:\n\n")

    for _, ex := range exercises {
        message.WriteString(fmt.Sprintf("• %s", ex.Name))
        if ex.Description != "" {
            message.WriteString(fmt.Sprintf(" - %s", ex.Description))
        }

        message.WriteString("\n")
    }
    return message.String(), nil
}

func (s *ExerciseService) ShowExerciseSelection(c telebot.Context) (*telebot.ReplyMarkup, error) {
    exercises, err := s.db.GetExercises()
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка упражнений")
	}

	menu := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for i := 0; i < len(exercises); i += 2 {
		var row telebot.Row
		btn1 := menu.Text(exercises[i].Name)
		row = append(row, btn1)
		if i+1 < len(exercises) {
			btn2 := menu.Text(exercises[i+1].Name)
			row = append(row, btn2)
		}
		rows = append(rows, row)
	}

	menu.Reply(rows...)
	// c.Set("exercises", exercises)
	// return c.Send("Выбери упражнение:", menu)
    return menu, nil
}
