package workout

import (
	"errors"
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

func NewWorkoutService(
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

func (s *WorkoutService) GetHistory(user *model.User, countList int) (string, error) { 
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

func (s *WorkoutService) GetUserWorkoutHistory(user *model.User, exercisId int, countList int) ([]model.ProgressPoint, error) { 
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

func (s *WorkoutService) HandleTextInput(
	user *model.User,
    text string,
    states *state.UserState,
) (model.MessageAnswer, error) {
    if states == nil {
        return model.MessageAnswer{}, errors.New("состояние не найдено")
    }

    switch {
    case states.WaitingForStats:
        return s.handleStatsInput(user, text, states)

    case states.WaitingForReps:
        return s.handleRepsInput(user, text, states)

    case states.WaitingForWeight:
        return s.handleWeightInput(user, text, states)

    default:
        return s.handleExerciseSelection(user, text, states)
    }
}

func (s *WorkoutService) handleExerciseSelection(
    user *model.User,
    text string,
    states *state.UserState,
) (model.MessageAnswer, error) {
    ex, err := s.findExerciseByName(text)
    if err != nil {
        return model.MessageAnswer{ReplyText: "Не нашёл такое упражнение"}, nil
    }
    if ex == nil {
        return model.MessageAnswer{ReplyText: "Выбери упражнение из списка"}, nil
    }

    states.CurrentExerciseID = ex.ID
    states.CurrentExerciseName = ex.Name
    states.WaitingForReps = true

    return model.MessageAnswer{
        ReplyText: fmt.Sprintf("Выбрано: %s\nВведи количество повторений:", ex.Name),
    }, nil
}

// 2. Ввод повторений
func (s *WorkoutService) handleRepsInput(
    user *model.User,
    text string,
    states *state.UserState,
) (model.MessageAnswer, error) {
    reps, err := strconv.Atoi(strings.TrimSpace(text))
    if err != nil || reps <= 0 {
        return model.MessageAnswer{ReplyText: "Нужно положительное число повторений"}, nil
    }

    states.TempReps = reps
    states.WaitingForReps = false
    states.WaitingForWeight = true

    return model.MessageAnswer{
        ReplyText: fmt.Sprintf("Повторений: %d\nТеперь вес (кг, 0 = без веса):", reps),
    }, nil
}

// 3. Ввод веса → сохранение
func (s *WorkoutService) handleWeightInput(
    user *model.User,
    text string,
    states *state.UserState,
) (model.MessageAnswer, error) {
    weight, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
    if err != nil || weight < 0 {
        return model.MessageAnswer{ReplyText: "Вес должен быть ≥ 0"}, nil
    }

    err = s.worksetRepository.SaveWorkoutSet(
        user.ID,
        states.CurrentExerciseID,
        weight,
        states.TempReps,
    )
    if err != nil {
        return model.MessageAnswer{}, fmt.Errorf("save set: %w", err)
    }

    msg := fmt.Sprintf(
        "Сохранено: %s — %d × %.1f кг",
        states.CurrentExerciseName, states.TempReps, weight,
    )

    // Сброс состояния
    s.resetAddSetState(states)

    return model.MessageAnswer{ReplyText: msg + "\n\nЧто дальше?"}, nil
}

func (s *WorkoutService) handleStatsInput(
    user *model.User,
    text string,
    states *state.UserState,
) (model.MessageAnswer, error) {
    ex, err := s.findExerciseByName(text)
    if err != nil {
        return model.MessageAnswer{}, err
    }
    if ex == nil {
        return model.MessageAnswer{ReplyText: "Такого упражнения нет"}, nil
    }

    points, err := s.GetUserWorkoutHistory(user, ex.ID, 1000)
    if err != nil {
        return model.MessageAnswer{}, err
    }

    buf, err := chart.GenerateProgressChart(points, ex.Name)
    if err != nil {
        return model.MessageAnswer{}, err
    }

    states.WaitingForStats = false

    return model.MessageAnswer{
        Buffer:    buf,
        ReplyText: ex.Name,
    }, nil
}

func (s *WorkoutService) findExerciseByName(name string) (*model.Exercise, error) {
    exercises, err := s.exerciseRepository.GetExercises()
    if err != nil {
        return nil, err
    }
    for _, e := range exercises {
        if strings.EqualFold(e.Name, name) { // можно улучшить до fuzzy search позже
            return &e, nil
        }
    }
    return nil, nil
}

func (s *WorkoutService) resetAddSetState(states *state.UserState) {
    states.WaitingForReps = false
    states.WaitingForWeight = false
    states.TempReps = 0
    states.CurrentExerciseID = 0
    states.CurrentExerciseName = ""
}
