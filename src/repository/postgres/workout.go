package postgres

import (
	"database/sql"
	"gofitness/src/model"
	"time"
)

type WorkoutPostgres struct {
	db *sql.DB
}

func NewWorkoutRepository(db *sql.DB) *WorkoutPostgres {
	return &WorkoutPostgres{db: db}
}

func (r *WorkoutPostgres) GetByID(id int) (*model.WorkoutSet, error) {
	query := `SELECT id, user_id, exercise_name, repetitions, weight, created_at FROM workout_sets WHERE id = $1`

	var set model.WorkoutSet
	err := r.db.QueryRow(query, id).Scan(
		&set.ID,
		&set.UserID,
		&set.ExerciseID,
		&set.Reps,
		&set.Weight,
		&set.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &set, nil
}

func (r *WorkoutPostgres) CreateWorkoutSet(set model.WorkoutSet) (*model.WorkoutSet, error) {
	query := `INSERT INTO workout_sets (user_id, exercise_id, weight, reps) VALUES ($1, $2, $3, $4)`

	err := r.db.QueryRow(query, set.UserID, set.ExerciseID, set.Weight, set.Reps).Scan(
		&set.ID,
	)

	return &set, err
}	

func (r *WorkoutPostgres) GetListWorkoutSetsByUserID(userID int64) ([]model.WorkoutSet, error) {
	query := `SELECT id, user_id, exercise_id, reps, weight, created_at 
	FROM workout_sets WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []model.WorkoutSet
	for rows.Next() {
		var set model.WorkoutSet
		if err := rows.Scan(
			&set.ID,
			&set.UserID,
			&set.ExerciseID,
			&set.Reps,
			&set.Weight,
			&set.CreatedAt,
		); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}

	return sets, nil
}


func (r *WorkoutPostgres) GetProgressByExercise(userID int64, exerciseID int, days int) ([]model.ProgressPoint, error) {
    query := `
        SELECT 
            DATE_TRUNC('day', ws.created_at) AS day,
            ws.weight * ws.reps          AS total_volume,
            ws.weight                    AS avg_weight,
            ws.reps                      AS avg_reps
        FROM workout_sets ws
        WHERE ws.user_id = $1
          AND ws.exercise_id = $2
        ORDER BY day ASC
    `

    rows, err := r.db.Query(query, userID, exerciseID)

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    var points []model.ProgressPoint
    for rows.Next() {
        var p model.ProgressPoint
        var day time.Time
        var volume, avgWeight, avgReps sql.NullFloat64
        // var count int

        err := rows.Scan(&day, &volume, &avgWeight, &avgReps)
        if err != nil {
            return nil, err
        }

        p.Date = day
        p.TotalVolume = volume.Float64
        p.AvgWeight = avgWeight.Float64
        p.AvgReps = avgReps.Float64
        // p.SetsCount = count

        points = append(points, p)
    }

    return points, nil
}



// Сохраняем подход (вес может быть 0)
func (r *WorkoutPostgres) SaveWorkoutSet(userID int64, exerciseID int, weight float64, reps int) error {
	query := `INSERT INTO workout_sets (user_id, exercise_id, weight, reps) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, userID, exerciseID, weight, reps)
	return err
}
