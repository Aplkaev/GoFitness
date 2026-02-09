package repository

import (
	"database/sql"
	"fmt"
	"gofitness/src/repository/postgres"

	"time"
    _ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

type Repositories struct {
	User     UserRepository
	Exercise ExerciseRepository
	Workout  WorkoutRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:     postgres.NewUserRepository(db),
		Exercise: postgres.NewExerciseRepository(db),
		Workout:  postgres.NewWorkoutRepository(db),
	}
}


func NewPostgres(connString string) (*Postgres, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Настройки пула подключений
	db.SetMaxOpenConns(25)        // Максимум 25 одновременных подключений
	db.SetMaxIdleConns(25)        // 25 подключений в пуле ожидания
	db.SetConnMaxLifetime(5 * time.Minute) // Подключение живет 5 минут

	return &Postgres{Db: db}, nil
}

func (p *Postgres) Init() error {
	return p.createStandardExercises()
}

// // Создаем стандартные упражнения (user_id = 0 для стандартных)
func (p *Postgres) createStandardExercises() error {
    fmt.Println("🔄 Инициализация стандартных упражнений...")
    
    standardExercises := []struct {
        name        string
        description string
    }{
        {"Приседания", "Приседания со штангой"},
        {"Жим лежа", "Жим штанги лежа"},
        {"Становая тяга", "Классическая становая тяга"},
        {"Подтягивания", "Подтягивания широким хватом"},
        {"Отжимания", "Отжимания от пола"},
        {"Жим стоя", "Армейский жим"},
        {"Тяга штанги", "Тяга штанги в наклоне"},
        {"Бицепс", "Подъем штанги на бицепс"},
        {"Трицепс", "Жим лежа узким хватом"},
        {"Планка", "Упражнение на пресс"},
    }
    repos := NewRepositories(p.Db)
    successCount := 0
    for _, exercise := range standardExercises {
        result, err := repos.Exercise.GetByName(exercise.name)
        if err != nil && err != sql.ErrNoRows {
            fmt.Printf("❌ Ошибка при проверке упражнения '%s': %v\n", exercise.name, err)
            continue
        }
        if result != nil {
            fmt.Printf("⚠️ Упражнение '%s' уже существует, пропускаем\n", exercise.name)
            continue
        }
        
         _, err = repos.Exercise.CreateUserExercise(0, exercise.name, exercise.description, true)
        if err != nil {
            fmt.Printf("❌ Ошибка при добавлении упражнения '%s': %v\n", exercise.name, err)
            continue
        }

        fmt.Printf("✅ Добавлено упражнение: %s\n", exercise.name)
        successCount++
    }
    
    fmt.Printf("🎯 Инициализация завершена. Добавлено %d/%d упражнений\n", 
        successCount, len(standardExercises))
    return nil
}

// // Получаем список упражнений
// func (p *Postgres) GetExercises() ([]model.Exercise, error) {
// 	query := `SELECT id, name, description FROM exercises ORDER BY name`
// 	rows, err := p.db.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var exercises []model.Exercise
// 	for rows.Next() {
// 		var ex model.Exercise
// 		if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description); err != nil {
// 			return nil, err
// 		}
// 		exercises = append(exercises, ex)
// 	}

// 	return exercises, nil
// }



func (p *Postgres) Close() error {
	return p.Db.Close()
}