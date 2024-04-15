package task

var stateTransitionMap = map[State][]State{
	Pending:   {Scheduled, Failed},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func ValidateStateTransition(src State, dst State) bool {
	return contains(stateTransitionMap[src], dst)
}
