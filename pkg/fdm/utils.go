package fdm

func LevelWeight(l DegradationLevel) int {
	switch l {
	case DegradationNone:
		return 0
	case DegradationMinor:
		return 1
	case DegradationMajor:
		return 2
	case DegradationCritical:
		return 3
	default:
		return 0
	}
}

