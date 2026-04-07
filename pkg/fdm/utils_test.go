package fdm

import "testing"

func TestLevelWeight(t *testing.T) {
	tests := []struct {
		level  DegradationLevel
		weight int
	}{
		{DegradationNone, 0},
		{DegradationMinor, 1},
		{DegradationMajor, 2},
		{DegradationCritical, 3},
		{"unknown", 0},
	}

	for _, tc := range tests {
		w := LevelWeight(tc.level)
		if w != tc.weight {
			t.Errorf("LevelWeight(%s) = %d, expected %d", tc.level, w, tc.weight)
		}
	}
}
