package bot

import (
	"fmt"
	"gofitness/src/helper"
	"gofitness/src/repository"
	"gofitness/src/service"
	"gofitness/src/state"
	"log"
	"strings"

	"gopkg.in/telebot.v3"
)

// Состояние пользователя для ввода подхода
var userStates = make(map[int64]*state.UserState)

func SetupHandlers(b *telebot.Bot, db *repository.Postgres, services *service.Services) {

	log.Printf("Start handler")
	// сохраняем пользователя и отдаем команды
	b.Handle("/start", func(c telebot.Context) error {
		user := c.Sender()
		username := helper.GetUserName(user)
		return c.Send(services.Workout.HandlerStart(user.ID, username))
	})

	// Команда /add - начать добавление подхода
	b.Handle("/add", func(c telebot.Context) error {
		var menu, err = services.Exercise.ShowExerciseSelection()

		if err == nil {
			c.Send(err)
		}

		return c.Send("Выбери упражнение:", menu)
	})

	// Команда /exercises - список упражнений
	b.Handle("/exercises", func(c telebot.Context) error {
		var message, err = services.Exercise.GetExercises()
		if err != nil {
			return c.Send(err)
		}
		return c.Send(message)
	})

	// Команда /history - история тренировок
	b.Handle("/history", func(c telebot.Context) error {
		user := c.Sender()
		username := helper.GetUserName(user)
		userModel, err := services.User.GetOrCreateUser(user.ID, username)
		if err != nil { 
			c.Send("Ошибка сохранения пользователя")
		}
		var message, _ = services.Workout.GetHistory(userModel, 10)
		return c.Send(message)
	})

	// Команда /stats - статистика
	b.Handle("/stats", func(c telebot.Context) error {
		userID := c.Sender().ID

		states, exists := userStates[userID]
		if !exists {
			states = &state.UserState{}
			userStates[userID] = states
		}

		states.WaitingForStats = true

		var menu, err = services.Exercise.ShowExerciseSelection()

		if err == nil {
			c.Send(err)
		}

		return c.Send("Выбери упражнение:", menu)
	})

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		user := c.Sender()
		username := helper.GetUserName(user)
		userModel, err := services.User.GetOrCreateUser(user.ID, username)
		if err != nil { 
			c.Send("Ошибка сохранения пользователя")
		}
		
		text := strings.TrimSpace(c.Text())

		states, exists := userStates[userModel.ChatID]
		if !exists {
			states = &state.UserState{}
			userStates[userModel.ChatID] = states
		}

		reply, err := services.Workout.HandleTextInput(userModel, text, states)
		if err != nil {
			log.Printf("Ошибка сохранения истории: %v", err)
			return c.Send("Произошла ошибка. Попробуй позже.")
		}


		if reply.Buffer != nil {
			photo := &telebot.Photo{
				File:    telebot.FromReader(reply.Buffer),
				Caption: fmt.Sprintf("Прогресс по %s за 90 дней\nСиняя — вес, оранжевая — повторения", reply.ReplyText),
			}
			return c.Send(photo)
		}

		return c.Send(reply.ReplyText)
	})

	log.Printf("End handler")
}
