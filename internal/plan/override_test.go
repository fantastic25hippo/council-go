package plan

import (
	"strings"
	"testing"
)

func TestParseOverride(t *testing.T) {
	tests := []struct {
		name       string
		task       string
		want       string
		wantErrMsg string
	}{
		{
			name: "no override",
			task: "some task\nhello world\n",
			want: "",
		},
		{
			name: "valid override jules",
			task: "some task\nWORKER-OVERRIDE: jules\nmore text",
			want: "jules",
		},
		{
			name: "valid override local",
			task: "some task\nworker-override: local\nmore text",
			want: "local",
		},
		{
			name: "valid override auto",
			task: "some task\nworker-override:  AUTO \nmore text",
			want: "auto",
		},
		{
			name: "invalid value ignored",
			task: "WORKER-OVERRIDE: invalid\nWORKER-OVERRIDE: jules",
			want: "jules",
		},
		{
			name: "duplicate identical values",
			task: "WORKER-OVERRIDE: jules\nsomething\nworker-override: JULES\n",
			want: "jules",
		},
		{
			name:       "conflicting values",
			task:       "WORKER-OVERRIDE: jules\nWORKER-OVERRIDE: local\n",
			wantErrMsg: "conflicting WORKER-OVERRIDE",
		},
		{
			name:       "conflicting values spacing",
			task:       "worker-override: auto\nWORKER-OVERRIDE: jules",
			wantErrMsg: "conflicting WORKER-OVERRIDE",
		},
		{
			name: "only invalid values",
			task: "WORKER-OVERRIDE: wrong1\nWORKER-OVERRIDE: wrong2",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOverride(tt.task)
			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			}
		})
	}
}
