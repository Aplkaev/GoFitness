package history

import (
	"fmt"
	"log"
	"gofitness/src/database"
	"gofitness/src/state"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
)

type HistoryService struct { 
	db *database.Postgres
}

// NewHistoryService - конструктор для HistoryService
func NewHistoryService(db *database.Postgres) *HistoryService {
    return &HistoryService{
        db: db,
    }
}

var (
	btnSelectExercise = telebot.Btn{Text: "🏋️ Выбрать упражнение"}
	btnSkipWeight     = telebot.Btn{Text: "➡️ Без веса"}
)

func (s *HistoryService) GetHistory(chatID int64, username string, countList int) (string, error) { 
	user, err := s.db.GetOrCreateUser(chatID, username)

	sets, err := s.db.GetUserWorkoutHistory(user.ID, countList)
	if err != nil {
		return "Ошибка при получении истории тренировок", nil
	}

	if len(sets) == 0 {
		return "У тебя пока нет записанных подходов. Используй /add чтобы добавить первый подход!", nil
	}

	var message strings.Builder
	message.WriteString("📊 Последние подходы:\n\n")
	
	for _, set := range sets {
		timeStr := set.CreatedAt.Format("02.01 15:04")
		if set.Weight > 0 {
			message.WriteString(fmt.Sprintf("• %s: %.1f кг × %d\n  %s\n", 
				set.ExerciseName, set.Weight, set.Reps, timeStr))
		} else {
			message.WriteString(fmt.Sprintf("• %s: %d раз\n  %s\n", 
				set.ExerciseName, set.Reps, timeStr))
		}
	}
	return message.String(), nil
}	

func (s *HistoryService) GetUserWorkoutHistory(chatID int64, username string, countList int) (string, error) { 
	user, err := s.db.GetOrCreateUser(chatID, username)

	sets, err := s.db.GetUserWorkoutHistory(user.ID, countList)
	if err != nil {
		return "Ошибка при получении статистики", nil
	}

	if len(sets) == 0 {
		return "Пока нет данных для статистики", nil
	}

	exerciseCount := make(map[string]int)
	totalSets := len(sets)
	var totalReps int

	for _, set := range sets {
		exerciseCount[set.ExerciseName]++
		totalReps += set.Reps
	}

	var message strings.Builder
	message.WriteString("📈 Статистика тренировок:\n\n")
	message.WriteString(fmt.Sprintf("Всего подходов: %d\n", totalSets))
	message.WriteString(fmt.Sprintf("Всего повторений: %d\n\n", totalReps))
	message.WriteString("Частота упражнений:\n")

	for exercise, count := range exerciseCount {
		message.WriteString(fmt.Sprintf("• %s: %d раз\n", exercise, count))
	}

	return message.String(), nil
}

func (s *HistoryService) HandlerStart(chatID int64, username string) (string) {
	var _, err = s.db.SaveUser(chatID, username)
	// Сохраняем пользователя в БД
	if err != nil {
		log.Printf("Failed to save user: %v", err)
	}
	return `🏋️‍♂️ Привет! Я твой фитнес-помощник!

Доступные команды:
/add - Добавить подход
/history - История тренировок  
/exercises - Список упражнений
/stats - Статистика тренировок

Нажми /add чтобы начать тренировку!
Набираем по всякому! ходж твинс!`;
}

// HistoryService
func (s *HistoryService) SaveHistory(
    chatID int64,
    message string,
	username string,
    state *state.UserState,
) (string, error) { 
	user, err := s.db.GetOrCreateUser(chatID, username)
    if err != nil {
        return "", fmt.Errorf("ошибка получения/создания пользователя: %w", err)
    }

    // 2. Обрабатываем в зависимости от состояния
    if state == nil {
        return "Состояние не найдено. Нажми /add чтобы начать.", nil
    }

    if state.WaitingForReps {
        reps, err := strconv.Atoi(message)
        if err != nil || reps <= 0 {
            return "Пожалуйста, введи положительное число повторений.", nil
        }

        state.TempReps = reps
        state.WaitingForReps = false
        state.WaitingForWeight = true

        return fmt.Sprintf(
            "Отлично! Теперь введи вес (кг, 0 — без веса). Повторений: %d",
            reps,
        ), nil
    }

    if state.WaitingForWeight {
        weight, err := strconv.ParseFloat(message, 64)
        if err != nil || weight < 0 {
            return "Введи корректный вес (>= 0).", nil
        }

        // Сохраняем подход в базу
        err = s.db.SaveWorkoutSet(user.ID, state.CurrentExerciseID, weight, state.TempReps)
        if err != nil {
            return "", fmt.Errorf("ошибка сохранения подхода: %w", err)
        }


        msg := fmt.Sprintf(
            "Подход сохранён: %s — %d повторений, %.1f кг.",
            state.CurrentExerciseName, state.TempReps, weight,
        )

        // Сбрасываем состояние
        state.WaitingForWeight = false
        state.TempReps = 0
        state.CurrentExerciseID = 0
        state.CurrentExerciseName = ""

        return msg + "\n\nЧто дальше?", nil
    }

	exercises, err := s.db.GetExercises()
	if err != nil {
		return "Ошибка при получении упражнений.", nil
	}

	var found bool
	for _, ex := range exercises {
		if ex.Name == message {
			state.CurrentExerciseID = ex.ID // Предполагаю, что в модели Exercise есть ID
			state.CurrentExerciseName = ex.Name
			state.WaitingForReps = true
			found = true
			break
		}
	}
	if !found {
		return "Неизвестное упражнение. Выбери из списка с помощью /add.", nil
	}

	return fmt.Sprintf("Выбрано: %s. Теперь введи количество повторений (например, 10).", message), nil
}

// Обработчик инлайн-кнопок упражнений
func SetupInlineHandlers(b *telebot.Bot, db *database.Postgres) {
	b.Handle(telebot.OnCallback, func(c telebot.Context) error {
		// user := c.S ender()
		data := c.Callback().Data

		if strings.HasPrefix(data, "exercise_") {
			exerciseIDStr := strings.TrimPrefix(data, "exercise_")
			exerciseID, err := strconv.Atoi(exerciseIDStr)
			if err != nil {
				return c.Respond(&telebot.CallbackResponse{Text: "Ошибка выбора упражнения"})
			}
			return handleExerciseSelection(c, db, exerciseID)
		}

		return nil
	})
}

// Обработчик выбора упражнения
func handleExerciseSelection(c telebot.Context, db *database.Postgres, exerciseID int) error {
	// user := c.Sender()
	exercise, err := db.GetExerciseByID(exerciseID)
	if err != nil {
		return c.Send("Упражнение не найдено")
	}

	// Сохраняем состояние пользователя
	// userStates[user.ID] = &UserState{
	// 	WaitingForReps:     true,
	// 	CurrentExerciseID:  exerciseID,
	// 	CurrentExerciseName: exercise.Name,
	// }

	return c.Send(fmt.Sprintf("Выбрано: %s\n\nТеперь введи количество повторений (только цифру):", exercise.Name))
}

// Обработчик ввода повторений
func handleRepsInput(message string, WaitingForReps bool, WaitingForWeight bool) string {
	reps, err := strconv.Atoi(message)
	if err != nil || reps <= 0 {
		return "Неверный формат повторений. Введи целое число больше 0:"
	}

	// Обновляем состояние
	WaitingForReps = false
	WaitingForWeight = true

	// Создаем меню для веса
	// menu := &telebot.ReplyMarkup{}
	// menu.Reply(menu.Row(btnSkipWeight))

	// return c.Send(fmt.Sprintf("Повторения: %d\n\nТеперь введи вес в кг (например: 67.5)\nИли нажми 'Без веса' если упражнение без отягощения:", reps), menu)
	return fmt.Sprintf("Повторения: %d\n\nТеперь введи вес в кг (например: 67.5)\nИли нажми 'Без веса' если упражнение без отягощения:",  reps)
}

// Обработчик ввода веса
func (s *HistoryService) handleWeightInput(chatID int64, CurrentExerciseID int, CurrentExerciseName, message string) string {
	// user := c.Sender()
	// state := userStates[user.ID]

	var weight float64
	// var err error

	// if message == btnSkipWeight.Text || message == "0" {
	// 	weight = 0
	// } else {
	// 	weight, err = strconv.ParseFloat(message, 64)
	// 	if err != nil {
	// 		return c.Send("Неверный формат веса. Введи число (например: 67.5):")
	// 	}
	// }

	// Сохраняем подход в БД
	if err := s.db.SaveWorkoutSet(chatID, CurrentExerciseID, weight, 0); err != nil {
		// Предположим, что reps уже были сохранены или нужно исправить логику
		// log.Printf("Failed to save workout set: %v", err)
		return "Ошибка при сохранении подхода"
	}

	// Формируем ответ
	response := fmt.Sprintf("✅ Подход сохранен!\n"+
		"Упражнение: %s\n", CurrentExerciseName)
	
	if weight > 0 {
		response += fmt.Sprintf("Вес: %.1f кг\n", weight)
	} else {
		response += "Без веса\n"
	}
	
	response += fmt.Sprintf("Время: %s", time.Now().Format("15:04"))

	// Очищаем состояние пользователя
	// delete(userStates, user.ID)

	return response
}

func handleWorkoutMessage(c telebot.Context, db *database.Postgres, message string) error {
	user := c.Sender()
	parts := strings.Fields(message)

	// Простой формат: "Упражнение повторения" или "Упражнение вес повторения"
	if len(parts) == 2 {
		// Формат: "Упражнение повторения"
		exerciseName := parts[0]
		repsStr := parts[1]

		reps, err := strconv.Atoi(repsStr)
		if err != nil || reps <= 0 {
			return c.Send("Неверный формат повторений")
		}

		exercise, err := db.GetExerciseByName(exerciseName)
		if err != nil {
			return c.Send(fmt.Sprintf("Упражнение '%s' не найдено", exerciseName))
		}

		if err := db.SaveWorkoutSet(user.ID, exercise.ID, 0, reps); err != nil {
			return c.Send("Ошибка при сохранении подхода")
		}

		return c.Send(fmt.Sprintf("✅ %s: %d раз", exercise.Name, reps))
	}

	return c.Send("Используй формат: Упражнение Повторения\nИли нажми /add для выбора из списка")
}