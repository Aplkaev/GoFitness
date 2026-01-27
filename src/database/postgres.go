package database

import (
	"database/sql"
	"fmt"
	"gofitness/src/model"

	"log"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
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

func (p *Postgres) GetUserByChatID(chatID int64) (*model.User, error) {
    query := `SELECT id, chat_id, username
              FROM users WHERE chat_id = $1`
    
    var user model.User
    err := p.db.QueryRow(query, chatID).Scan(
        &user.ID, 
		&user.ChatID, 
		&user.Username, 
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса пользователя: %w", err)
    }
    
    return &user, nil
}

func (p *Postgres) GetOrCreateUser(chatID int64, username string) (*model.User, error) {
    user, err := p.GetUserByChatID(chatID)
    if err != nil {
        return nil, err
    }
    if user != nil {
        return user, nil
    }

    return p.SaveUser(chatID, username)
}

// Сохраняем или получаем пользователя
func (p *Postgres) SaveUser(chatID int64, username string) (*model.User, error) {
    query := `
        INSERT INTO users (chat_id, username) 
        VALUES ($1, $2) 
        ON CONFLICT (chat_id) 
        DO UPDATE SET 
            username = EXCLUDED.username
        RETURNING id, chat_id, username, created_at
    `
    
    var user model.User
    err := p.db.QueryRow(
        query, 
        chatID, 
		username,
    ).Scan(
        &user.ID,
        &user.ChatID, 
        &user.Username,
        &user.CreatedAt,
    )
	log.Printf("save user %d %d %s", user.ID, user.ChatID, user.Username)

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
func (p *Postgres) GetExercises() ([]model.Exercise, error) {
	query := `SELECT id, name, description FROM exercises ORDER BY name`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var ex model.Exercise
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
func (p *Postgres) GetUserWorkoutHistory(userID int64, limit int) ([]model.WorkoutSet, error) {
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

	var sets []model.WorkoutSet
	for rows.Next() {
		var set model.WorkoutSet
		if err := rows.Scan(&set.ID, &set.ExerciseID, &set.ExerciseName, &set.Weight, &set.Reps, &set.CreatedAt); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}

	return sets, nil
}

// В Postgres репозитории
func (p *Postgres) GetProgressByExercise(userID int64, exerciseID int, days int) ([]model.ProgressPoint, error) {
    query := `
        SELECT 
            DATE_TRUNC('day', ws.created_at) AS day,
            SUM(ws.weight * ws.reps)          AS total_volume,
            AVG(ws.weight)                    AS avg_weight,
            AVG(ws.reps)                      AS avg_reps,
            COUNT(*)                          AS sets_count
        FROM workout_sets ws
        WHERE ws.user_id = $1
          AND ws.exercise_id = $2
          AND ws.created_at >= NOW() - $3 * INTERVAL '1 day'
        GROUP BY day
        ORDER BY day ASC
    `

    rows, err := p.db.Query(query, userID, exerciseID, days)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var points []model.ProgressPoint
    for rows.Next() {
        var p model.ProgressPoint
        var day time.Time
        var volume, avgWeight, avgReps sql.NullFloat64
        var count int

        err := rows.Scan(&day, &volume, &avgWeight, &avgReps, &count)
        if err != nil {
            return nil, err
        }

        p.Date = day
        p.TotalVolume = volume.Float64
        p.AvgWeight = avgWeight.Float64
        p.AvgReps = avgReps.Float64
        p.SetsCount = count

        points = append(points, p)
    }

    return points, nil
}

// Получаем упражнение по ID
func (p *Postgres) GetExerciseByID(id int) (*model.Exercise, error) {
	query := `SELECT id, name, description FROM exercises WHERE id = $1`
	var exercise model.Exercise
	err := p.db.QueryRow(query, id).Scan(&exercise.ID, &exercise.Name, &exercise.Description)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

// Получаем упражнение по имени
func (p *Postgres) GetExerciseByName(name string) (*model.Exercise, error) {
	query := `SELECT id, name, description FROM exercises WHERE name ILIKE $1`
	var exercise model.Exercise
	err := p.db.QueryRow(query, name).Scan(&exercise.ID, &exercise.Name, &exercise.Description)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}