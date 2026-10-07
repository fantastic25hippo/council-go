package verify

import (
	"reflect"
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		timeoutSecs int
		want        []string
		wantErr     string
	}{
		{
			name:        "normal command",
			cmd:         "python3 -m unittest discover",
			timeoutSecs: 60,
			want:        []string{"timeout", "60", "bash", "-lc", "python3 -m unittest discover"},
		},
		{
			name:        "double timeout command",
			cmd:         "timeout 60 ls",
			timeoutSecs: 180,
			want:        []string{"timeout", "180", "bash", "-lc", "timeout 60 ls"},
		},
		{
			name:        "timeoutSecs 0",
			cmd:         "ls",
			timeoutSecs: 0,
			wantErr:     "timeout",
		},
		{
			name:        "timeoutSecs -1",
			cmd:         "ls",
			timeoutSecs: -1,
			wantErr:     "timeout",
		},
		{
			name:        "empty cmd",
			cmd:         "",
			timeoutSecs: 60,
			wantErr:     "VERIFY",
		},
		{
			name:        "whitespace cmd",
			cmd:         "   ",
			timeoutSecs: 60,
			wantErr:     "VERIFY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Wrap(tt.cmd, tt.timeoutSecs)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Wrap() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Wrap() error = %v, expected containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Wrap() unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Wrap() = %v, want %v", got, tt.want)
			}
		})
	}
}
