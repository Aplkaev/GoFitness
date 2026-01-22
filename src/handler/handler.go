package bot

import (
	"gofitness/src/database"
	"gofitness/src/helper"
	"gofitness/src/service/exercise"
	"gofitness/src/service/history"
	"gofitness/src/state"
	"log"
	"strings"

	"gopkg.in/telebot.v3"
)

// Состояние пользователя для ввода подхода
var userStates = make(map[int64]*state.UserState)

func SetupHandlers(b *telebot.Bot, db *database.Postgres) {
	// Команда /start
	// Инициализируем сервисы
	exerciseService := exercise.NewExerciseService(db)
	historyService := history.NewHistoryService(db)
	log.Printf("Start handler")
	// сохраняем пользователя и отдаем команды
	b.Handle("/start", func(c telebot.Context) error {
		user := c.Sender()
		return c.Send(historyService.HandlerStart(user.ID, helper.GetUserName(user)))
	})

	// Команда /add - начать добавление подхода
	b.Handle("/add", func(c telebot.Context) error {
		var menu, err = exerciseService.ShowExerciseSelection(c)

		if err == nil {
			c.Send(err)
		}

		return c.Send("Выбери упражнение:", menu)
	})

	// Команда /exercises - список упражнений
	b.Handle("/exercises", func(c telebot.Context) error {
		var message, err = exerciseService.GetExercises()
		if err != nil {
			return c.Send(err)
		}
		return c.Send(message)
	})

	// Команда /history - история тренировок
	b.Handle("/history", func(c telebot.Context) error {
		user := c.Sender()
		username := helper.GetUserName(user)
		var message, _ = historyService.GetHistory(user.ID, username, 10)
		return c.Send(message)
	})

	// Команда /stats - статистика
	b.Handle("/stats", func(c telebot.Context) error {
		user := c.Sender()
		username := helper.GetUserName(user)
		var message, _ = historyService.GetUserWorkoutHistory(user.ID, username, 100)
		return c.Send(message)
	})

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		username := helper.GetUserName(c.Sender())
		text := strings.TrimSpace(c.Text())

		states, exists := userStates[userID]
		if !exists {
			states = &state.UserState{}
			userStates[userID] = states
		}

		// Передаём управление сервису
		replyText, err := historyService.SaveHistory(userID, text, username, states)
		if err != nil {
			log.Printf("Ошибка сохранения истории: %v", err)
			return c.Send("Произошла ошибка. Попробуй позже.")
		}

		// Отправляем ответ пользователю
		return c.Send(replyText)
	})

	log.Printf("End handler")
}
