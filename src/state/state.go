package state

type UserState struct {
	WaitingForReps     	bool
	WaitingForWeight   	bool
	WaitingForStats    	bool
	WaitSelfWeight   	bool
	CurrentExerciseID  	int
	CurrentExerciseName string
	TempReps            int
}