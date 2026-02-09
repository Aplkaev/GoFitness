package workout

import (
	"fmt"
	"gofitness/src/model"
	"gofitness/src/repository"
	"gofitness/src/service"
	"gofitness/src/service/chart"
	"gofitness/src/state"

	"log"
	"strconv"
	"strings"
)

type WorkoutService struct { 
	worksetRepository repository.WorkoutRepository
	exerciseRepository repository.ExerciseRepository
	user service.UserService
}

func NewWorkoutWorkoutService(
	worksetRepository repository.WorkoutRepository, 
	exerciseRepository repository.ExerciseRepository,
	user service.UserService,
) *WorkoutService {
    return &WorkoutService{
        worksetRepository: worksetRepository,
        exerciseRepository: exerciseRepository,
		user: user,
    }
}

func (s *WorkoutService) GetHistory(chatID int64, username string, countList int) (string, error) { 
	user, err := s.user.GetOrCreateUser(chatID, username)

	// sets, err := s.db.GetUserWorkoutHistory(user.ID, countList)
	sets, err := s.worksetRepository.GetListWorkoutSetsByUserID(user.ID)
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

func (s *WorkoutService) GetUserWorkoutHistory(chatID int64, username string, exercisId int, countList int) ([]model.ProgressPoint, error) { 
	user, err := s.user.GetOrCreateUser(chatID, username)

	// points, err := historyWorkoutService.GetProgressPoints(userID, exerciseID, 90) // твоя функция из БД
	// if err != nil || len(points) < 2 {
	// 	return c.Send("Недостаточно данных для графика (нужно минимум 2 тренировки)")
	// }


	// sets, err := s.db.GetUserWorkoutHistory(user.ID, countList)
	sets, err := s.worksetRepository.GetListWorkoutSetsByUserID(user.ID)

	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении статистики")
	}

	if len(sets) == 0 {
		return nil,  fmt.Errorf("Пока нет данных для статистики")
	}
	
	points, err := s.worksetRepository.GetProgressByExercise(user.ID, exercisId, countList)

	if err != nil {
		return nil,  fmt.Errorf(err.Error())
	}

	return points, err
}

func (s *WorkoutService) HandlerStart(chatID int64, username string) (string) {
	var _, err = s.user.GetOrCreateUser(chatID, username)
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

// убрать логику из сохранения сообщения
func (s *WorkoutService) SaveHistory(
    chatID int64,
    message string,
	username string,
    state *state.UserState,
) (model.MessageAnswer, error) { 
	user, err := s.user.GetOrCreateUser(chatID, username)
    if err != nil {
        return model.MessageAnswer{}, fmt.Errorf("ошибка получения/создания пользователя: %w", err)
    }

    // 2. Обрабатываем в зависимости от состояния
    if state == nil {
        return model.MessageAnswer{}, fmt.Errorf("Состояние не найдено. Нажми /add чтобы начать.")
    }

	if state.WaitingForStats {
		exercises, err_exe := s.exerciseRepository.GetExercises()
		if err_exe != nil {
			return model.MessageAnswer{}, fmt.Errorf("Ошибка при получении упражнений.")
		}
		// todo вынести в отедьное место
		var found bool
		var exercisId int = 0

		for _, ex := range exercises {
			if ex.Name == message {
				exercisId = ex.ID
				found = true
				break
			}
		}

		if !found {
			return model.MessageAnswer{ReplyText: "Неизвестное упражнение. Выбери из списка с помощью /add."}, nil
		}
		
		var points, err = s.GetUserWorkoutHistory(chatID, username, exercisId, 1000)
		
		if err != nil {
			log.Printf("Ошибка данных: %v", err)
			return model.MessageAnswer{}, fmt.Errorf("Ошибка данных: %w", err)
		}

		var buf, err_buf = chart.GenerateProgressChart(points, "");

		if err_buf != nil {
			log.Printf("Ошибка генерации графика: %v", err)
			return model.MessageAnswer{}, fmt.Errorf("Ошибка генерации графика: %w", err_buf)
		}
		
		return model.MessageAnswer{Buffer: buf}, nil
	}

    if state.WaitingForReps {
        reps, err := strconv.Atoi(message)
        if err != nil || reps <= 0 {
            return model.MessageAnswer{ReplyText: "Пожалуйста, введи положительное число повторений."}, nil
        }

        state.TempReps = reps
        state.WaitingForReps = false
        state.WaitingForWeight = true

        return model.MessageAnswer{
            ReplyText: fmt.Sprintf(
                "Отлично! Теперь введи вес (кг, 0 — без веса). Повторений: %d",
                reps,
            ),
        }, nil
    }

    if state.WaitingForWeight {
        weight, err := strconv.ParseFloat(message, 64)
        if err != nil || weight < 0 {
            return model.MessageAnswer{ReplyText: "Введи корректный вес (>= 0)."}, nil
        }

        // Сохраняем подход в базу
        err = s.worksetRepository.SaveWorkoutSet(user.ID, state.CurrentExerciseID, weight, state.TempReps)
        if err != nil {
            return model.MessageAnswer{}, fmt.Errorf("ошибка сохранения подхода: %w", err)
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

        return model.MessageAnswer{ReplyText: msg + "\n\nЧто дальше?"}, nil
    }

	exercises, err := s.exerciseRepository.GetExercises()
	if err != nil {
		return model.MessageAnswer{}, fmt.Errorf("Ошибка при получении упражнений.")
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
		return model.MessageAnswer{ReplyText: "Неизвестное упражнение. Выбери из списка с помощью /add."}, nil
	}

	return model.MessageAnswer{ReplyText: fmt.Sprintf("Выбрано: %s. Теперь введи количество повторений (например, 10).", message)}, nil
}
