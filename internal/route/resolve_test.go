package route

import (
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name         string
		override     string
		relayWorker  string
		originGitHub bool
		want         string
		wantErrMsg   string
	}{
		{
			name:         "override local beats relay jules and github origin",
			override:     "local",
			relayWorker:  "jules",
			originGitHub: true,
			want:         "local",
		},
		{
			name:         "override auto + originGitHub true",
			override:     "auto",
			relayWorker:  "local",
			originGitHub: true,
			want:         "jules",
		},
		{
			name:         "empty override, relay auto, originGitHub false",
			override:     "",
			relayWorker:  "auto",
			originGitHub: false,
			want:         "local",
		},
		{
			name:         "override jules",
			override:     "jules",
			relayWorker:  "local",
			originGitHub: false,
			want:         "jules",
		},
		{
			name:         "empty override, relay jules",
			override:     "",
			relayWorker:  "jules",
			originGitHub: false,
			want:         "jules",
		},
		{
			name:         "empty override, empty relay, originGitHub true",
			override:     "",
			relayWorker:  "",
			originGitHub: true,
			want:         "jules",
		},
		{
			name:         "invalid override",
			override:     "wrong",
			relayWorker:  "jules",
			originGitHub: false,
			wantErrMsg:   "invalid override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.override, tt.relayWorker, tt.originGitHub)
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
