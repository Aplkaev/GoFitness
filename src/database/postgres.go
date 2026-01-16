package database

import (
	"database/sql"
	"fmt"

	// "fmt"
	// "log"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

// Модели данных
type User struct {
	ID        int64
	ChatID    int64
	Username  string
	FirstName string
	LastName  string
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

	return &Postgres{db: db}, nil
}

func (p *Postgres) Init() error {
	return p.createStandardExercises()
}

func (p *Postgres) GetUserByChatID(chatID int64) (*User) {
	query := `SELECT id, chat_id, username, first_name, last_name FROM users WHERE chat_id = $1`
	row := p.db.QueryRow(query, chatID)

	if row == nil {
		return nil
	}

	var user User
	if err := row.Scan(&user.ID, &user.ChatID, &user.Username, &user.FirstName, &user.LastName); err != nil {
		return nil
	}

	return &user
}

// Сохраняем или получаем пользователя
func (p *Postgres) SaveUser(chatID int64, username, firstName, lastName string) (*User, error) {
    query := `
        INSERT INTO users (chat_id, username, first_name, last_name) 
        VALUES ($1, $2, $3, $4) 
        ON CONFLICT (chat_id) 
        DO UPDATE SET 
            username = EXCLUDED.username, 
            first_name = EXCLUDED.first_name, 
            last_name = EXCLUDED.last_name,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, chat_id, username, first_name, last_name, created_at, updated_at
    `
    
    var user User
    err := p.db.QueryRow(
        query, 
        chatID, username, firstName, lastName,
    ).Scan(
        &user.ID,
        &user.ChatID, 
        &user.Username,
        &user.FirstName,
        &user.LastName,
        &user.CreatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}


// Создаем стандартные упражнения (user_id = 0 для стандартных)
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

    successCount := 0
    for _, exercise := range standardExercises {
        // Сначала проверяем существует ли уже упражнение
        var exists bool
        checkQuery := `SELECT EXISTS(SELECT 1 FROM exercises WHERE name = $1)`
        err := p.db.QueryRow(checkQuery, exercise.name).Scan(&exists)
        
        if err != nil {
            fmt.Printf("❌ Ошибка при проверке упражнения '%s': %v\n", exercise.name, err)
            continue
        }
        
        if exists {
            fmt.Printf("⚠️ Упражнение '%s' уже существует, пропускаем\n", exercise.name)
            continue
        }
        
        // Если не существует - добавляем
        query := `INSERT INTO exercises (name, description, is_standard, user_id) VALUES ($1, $2, TRUE, 0)`
        _, err = p.db.Exec(query, exercise.name, exercise.description)
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

// Получаем список упражнений
func (p *Postgres) GetExercises() ([]Exercise, error) {
	query := `SELECT id, name, description FROM exercises ORDER BY name`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var ex Exercise
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description); err != nil {
			return nil, err
		}
		exercises = append(exercises, ex)
	}

	return exercises, nil
}

// Сохраняем подход (вес может быть 0)
func (p *Postgres) SaveWorkoutSet(userID int64, exerciseID int, weight float64, reps int) error {
	query := `INSERT INTO workout_sets (user_id, exercise_id, weight, reps) VALUES ($1, $2, $3, $4)`
	_, err := p.db.Exec(query, userID, exerciseID, weight, reps)
	return err
}

// Получаем историю подходов пользователя
func (p *Postgres) GetUserWorkoutHistory(userID int64, limit int) ([]WorkoutSet, error) {
	query := `
		SELECT ws.id, ws.exercise_id, e.name, ws.weight, ws.reps, ws.created_at 
		FROM workout_sets ws
		JOIN exercises e ON ws.exercise_id = e.id
		WHERE ws.user_id = $1 
		ORDER BY ws.created_at DESC 
		LIMIT $2
	`
	
	rows, err := p.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []WorkoutSet
	for rows.Next() {
		var set WorkoutSet
		if err := rows.Scan(&set.ID, &set.ExerciseID, &set.ExerciseName, &set.Weight, &set.Reps, &set.CreatedAt); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}

	return sets, nil
}

// Получаем упражнение по ID
func (p *Postgres) GetExerciseByID(id int) (*Exercise, error) {
	query := `SELECT id, name, description FROM exercises WHERE id = $1`
	var exercise Exercise
	err := p.db.QueryRow(query, id).Scan(&exercise.ID, &exercise.Name, &exercise.Description)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

// Получаем упражнение по имени
func (p *Postgres) GetExerciseByName(name string) (*Exercise, error) {
	query := `SELECT id, name, description FROM exercises WHERE name ILIKE $1`
	var exercise Exercise
	err := p.db.QueryRow(query, name).Scan(&exercise.ID, &exercise.Name, &exercise.Description)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}