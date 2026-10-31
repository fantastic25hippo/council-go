package verify

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCheckSyntax(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr string
	}{
		{
			name:    "empty cmd",
			cmd:     "",
			wantErr: "VERIFY",
		},
		{
			name:    "whitespace cmd",
			cmd:     "   ",
			wantErr: "VERIFY",
		},
		{
			name:    "valid cmd",
			cmd:     "echo hi",
			wantErr: "",
		},
		{
			name:    "broken quoting",
			cmd:     "echo '",
			wantErr: "VERIFY",
		},
		{
			name:    "command starting with hyphen",
			cmd:     "-s",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSyntax(tt.cmd)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("CheckSyntax() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("CheckSyntax() error = %v, expected containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("CheckSyntax() unexpected error: %v", err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "council-verify-run")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		cmd         string
		timeoutSecs int
		wantOutput  string
		wantExit    int
		wantErr     string
	}{
		{
			name:        "successful echo hi",
			cmd:         "echo hi",
			timeoutSecs: 60,
			wantOutput:  "hi",
			wantExit:    0,
			wantErr:     "",
		},
		{
			name:        "false command",
			cmd:         "false",
			timeoutSecs: 60,
			wantOutput:  "",
			wantExit:    1,
			wantErr:     "",
		},
		{
			name:        "timeoutSecs 0",
			cmd:         "echo hi",
			timeoutSecs: 0,
			wantOutput:  "",
			wantExit:    0,
			wantErr:     "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exit, err := Run(tempDir, tt.cmd, tt.timeoutSecs)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Run() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Run() error = %v, expected containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Run() unexpected error: %v", err)
				return
			}
			if exit != tt.wantExit {
				t.Errorf("Run() exit = %v, want %v", exit, tt.wantExit)
			}
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Run() output = %q, want containing %q", output, tt.wantOutput)
			}
		})
	}
}

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
			want:        []string{"timeout", "60", "bash", "-lc", "--", "python3 -m unittest discover"},
		},
		{
			name:        "double timeout command",
			cmd:         "timeout 60 ls",
			timeoutSecs: 180,
			want:        []string{"timeout", "180", "bash", "-lc", "--", "timeout 60 ls"},
		},
		{
			name:        "command starting with hyphen",
			cmd:         "-s",
			timeoutSecs: 60,
			want:        []string{"timeout", "60", "bash", "-lc", "--", "-s"},
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
