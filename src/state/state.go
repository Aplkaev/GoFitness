package state

type UserState struct {
	WaitingForReps     	bool
	WaitingForWeight   	bool
	WaitingForStats    	bool
	CurrentExerciseID  	int
	CurrentExerciseName string
	TempReps            int
}