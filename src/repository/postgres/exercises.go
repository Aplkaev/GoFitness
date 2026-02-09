package postgres

import (
	"database/sql"
	"gofitness/src/model"
)

type ExercisePostgres struct {
	db *sql.DB
}

func NewExerciseRepository(db *sql.DB) *ExercisePostgres {
	return &ExercisePostgres{db: db}
}

func (r *ExercisePostgres) GetByID(id int) (*model.Exercise, error) {
	query := `SELECT id, name, description, is_standard, user_id FROM exercises WHERE id = $1`
	var exercise model.Exercise
	err := r.db.QueryRow(query, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.IsStandard,
		&exercise.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *ExercisePostgres) GetByName(name string) (*model.Exercise, error) {
	query := `SELECT id, name, description, is_standard, user_id FROM exercises WHERE name = $1`
	var exercise model.Exercise
	err := r.db.QueryRow(query, name).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.IsStandard,
		&exercise.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}
    
func (r *ExercisePostgres) CreateUserExercise(userID int64, name, description string, isStandard bool) (*model.Exercise, error) {
	query := `INSERT INTO exercises (name, description, is_standard, user_id) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, name, description, isStandard, userID)

	if( err != nil) {
		return nil, err
	}
	
	exercise := model.Exercise{
		Name:        name,
		Description: description,
		IsStandard:  isStandard,
		UserID:      userID,
	}

	return &exercise, nil
}

func (r *ExercisePostgres) GetExercises() ([]model.Exercise, error) {
	query := `SELECT id, name, description FROM exercises ORDER BY name`
	rows, err := r.db.Query(query)
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