package mysql

import (
	"strings"
	"testing"
)

func TestQueriesLength(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"queryUpsertVM", queryUpsertVM},
		{"queryGetVM", queryGetVM},
		{"queryListVMsByCluster", queryListVMsByCluster},
		{"queryListProtectedVMs", queryListProtectedVMs},
		{"queryCreateTask", queryCreateTask},
		{"queryUpdateTaskStatus", queryUpdateTaskStatus},
		{"queryListTasksByPlan", queryListTasksByPlan},
		{"queryCreatePlan", queryCreatePlan},
		{"queryGetPlan", queryGetPlan},
		{"queryListStaleBootingClaims", queryListStaleBootingClaims},
		{"queryReleaseStaleClaim", queryReleaseStaleClaim},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(strings.TrimSpace(tt.query)) == 0 {
				t.Errorf("Query %s is empty", tt.name)
			}
		})
	}
}
