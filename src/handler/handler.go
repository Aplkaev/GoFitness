package bot

import (
	"fmt"
	"gofitness/src/database"
	"gofitness/src/service/exercise"
	"gofitness/src/service/history"
	"log"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"
)

var exerciseSvc *exercise.ExerciseService
var historySvc *history.HistoryService

// Состояние пользователя для ввода подхода
type UserState struct {
	WaitingForReps     bool
	WaitingForWeight   bool
	CurrentExerciseID  int
	CurrentExerciseName string

}

var userStates = make(map[int64]*UserState)
var exerciseBtn = telebot.Btn{Unique: "exercis"}

func SetupHandlers(b *telebot.Bot, db *database.Postgres) {
	// Команда /start
	// Инициализируем сервисы
	exerciseService := exercise.NewExerciseService(db)
	historyService := history.NewHistoryService(db)
	log.Printf("Start handler")		
	b.Handle("/start", func(c telebot.Context) error {
		user := c.Sender()
		var _, err = db.SaveUser(user.ID, user.Username, user.FirstName, user.LastName)
		// Сохраняем пользователя в БД
		if err != nil {
			log.Printf("Failed to save user: %v", err)
		}

		return c.Send(`🏋️‍♂️ Привет! Я твой фитнес-помощник!

Доступные команды:
/add - Добавить подход
/history - История тренировок  
/exercises - Список упражнений
/stats - Статистика тренировок

Нажми /add чтобы начать тренировку!
Набираем по всякому! ходж твинс! test`)
	})

	// Команда /add - начать добавление подхода
	b.Handle("/add", func(c telebot.Context) error {
		return showExerciseSelection(c, db)
	})

	// Команда /exercises - список упражнений
	b.Handle("/exercises", func(c telebot.Context) error {
		fmt.Println("trst exercises")
		user := c.Sender()
		fmt.Println("trst exercises", user.ID)
		var message, err = exerciseService.GetExercises(user.ID)
		if err != nil { 
			return c.Send(err)
		}
		return c.Send(message)
	})

	// Команда /history - история тренировок
	b.Handle("/history", func(c telebot.Context) error {
		user := c.Sender()
		var message, _ = historyService.GetHistory(user.ID, 10)
		return c.Send(message)
	})

	// Команда /stats - статистика
	b.Handle("/stats", func(c telebot.Context) error {
		user := c.Sender()
		var message, _ = historyService.GetUserWorkoutHistory(user.ID, 100)
		return c.Send(message)
	})

	// Обработчик кнопок с упражнениями
	// b.Handle(&btnSelectExercise, func(c telebot.Context) error {
	// 	return showExerciseSelection(c, db)
	// })

	// Обработчик текстовых сообщений (выбор упражнения или ввод reps/веса)
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		text := strings.TrimSpace(c.Text())
		state, exists := userStates[userID]
		if !exists {
			state = &UserState{}
			userStates[userID] = state
		}

		// Если ожидаем повторения
		if state.WaitingForReps {
			reps, err := strconv.Atoi(text)
			if err != nil || reps <= 0 {
				return c.Send("Пожалуйста, введи положительное число повторений (например, 9).")
			}
			// Сохраняем reps временно (можно добавить поле в state, если нужно)
			state.WaitingForReps = false
			state.WaitingForWeight = true
			// Сохраняем reps в дополнительном поле, если нужно (для простоты используем временную переменную или расширьте state)
			// Здесь я использую fmt для примера, но лучше добавить Reps int в UserState
			c.Set("temp_reps", reps) // Используем контекст для временного хранения
			return c.Send(fmt.Sprintf("Отлично! Теперь введи вес (в кг, например, 113). Если без веса — введи 0."))
		}

		// Если ожидаем вес
		if state.WaitingForWeight {
			weight, err := strconv.ParseFloat(text, 64)
			if err != nil || weight < 0 {
				return c.Send("Пожалуйста, введи вес (число >= 0, например, 80).")
			}
			reps := c.Get("temp_reps").(int) // Получаем из контекста
			// Сохраняем в историю (предполагаю, что у historySvc есть метод SaveSet)
			// err = historySvc.SaveSet(userID, state.CurrentExerciseID, reps, weight) // Добавьте такой метод, если нет
			// if err != nil {
			// 	log.Printf("Failed to save set: %v", err)
			// 	return c.Send("Ошибка сохранения подхода. Попробуй снова.")
			// }
			// Сбрасываем состояние
			state.WaitingForWeight = false
			state.CurrentExerciseID = 0
			state.CurrentExerciseName = ""
			delete(userStates, userID) // Очищаем, если не нужно

			// Предлагаем добавить ещё
			menu := &telebot.ReplyMarkup{}
			menu.Reply(
				menu.Row(menu.Text("Добавить ещё подход")),
				menu.Row(menu.Text("Завершить тренировку")),
			)
			return c.Send(fmt.Sprintf("Подход сохранён: %s - %d повторений, %.1f кг.", state.CurrentExerciseName, reps, weight), menu)
		}

		// Если ничего не ожидаем — проверяем, является ли текст упражнением
		exercises, err := db.GetExercises()
		if err != nil {
			return c.Send("Ошибка при получении упражнений.")
		}
		var found bool
		for _, ex := range exercises {
			if ex.Name == text {
				state.CurrentExerciseID = ex.ID // Предполагаю, что в модели Exercise есть ID
				state.CurrentExerciseName = ex.Name
				state.WaitingForReps = true
				found = true
				break
			}
		}
		if !found {
			return c.Send("Неизвестное упражнение. Выбери из списка с помощью /add.")
		}

		// Убираем клавиатуру после выбора
		removeMenu := &telebot.ReplyMarkup{RemoveKeyboard: true}
		c.Send("Клавиатура убрана.", removeMenu)

		return c.Send(fmt.Sprintf("Выбрано: %s. Теперь введи количество повторений (например, 10).", text))
	})
	log.Printf("End handler")		
}

func showExerciseSelection(c telebot.Context, db *database.Postgres) error {
	exercises, err := db.GetExercises()
	if err != nil {
		return c.Send("Ошибка при получении списка упражнений")
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
	c.Set("exercises", exercises)
	return c.Send("Выбери упражнение:", menu)
}