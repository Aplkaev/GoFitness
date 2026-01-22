package state

type UserState struct {
	WaitingForReps     bool
	WaitingForWeight   bool
	CurrentExerciseID  int
	CurrentExerciseName string
	TempReps            int
}