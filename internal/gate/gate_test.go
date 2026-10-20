package gate

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestProtectedOverlap(t *testing.T) {
	tests := []struct {
		name    string
		owned   []string
		wantErr bool
	}{
		{
			name:    "nil for internal/evidence",
			owned:   []string{"internal/evidence"},
			wantErr: false,
		},
		{
			name:    "errors on hook.log",
			owned:   []string{"hook.log"},
			wantErr: true,
		},
		{
			name:    "errors on ./hook.log",
			owned:   []string{"./hook.log"},
			wantErr: true,
		},
		{
			name:    "allows logs/hook.log in unrelated dir",
			owned:   []string{"logs/hook.log"},
			wantErr: false,
		},
		{
			name:    "errors on council-jules-async.md",
			owned:   []string{"council-jules-async.md"},
			wantErr: true,
		},
		{
			name:    "errors on dot whole tree",
			owned:   []string{"."},
			wantErr: true,
		},
		{
			name:    "errors on dot slash whole tree",
			owned:   []string{"./"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ProtectedOverlap(tt.owned)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProtectedOverlap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "protected") {
				t.Errorf("ProtectedOverlap() error should contain 'protected', got %v", err)
			}
		})
	}
}

func TestGitHubReady(t *testing.T) {
	tests := []struct {
		name          string
		caps          worker.Capabilities
		dirty         bool
		unpushed      bool
		wantErr       bool
		errContains string
	}{
		{
			name:          "ReachesGitHub true + dirty true",
			caps:          worker.Capabilities{ReachesGitHub: true},
			dirty:         true,
			unpushed:      false,
			wantErr:       true,
			errContains: "dirty",
		},
		{
			name:          "ReachesGitHub true + unpushed true (dirty false)",
			caps:          worker.Capabilities{ReachesGitHub: true},
			dirty:         false,
			unpushed:      true,
			wantErr:       true,
			errContains: "unpushed",
		},
		{
			name:          "ReachesGitHub false + dirty true + unpushed true",
			caps:          worker.Capabilities{ReachesGitHub: false},
			dirty:         true,
			unpushed:      true,
			wantErr:       false,
		},
		{
			name:          "ReachesGitHub true + both false",
			caps:          worker.Capabilities{ReachesGitHub: true},
			dirty:         false,
			unpushed:      false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GitHubReady(tt.caps, tt.dirty, tt.unpushed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubReady() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("GitHubReady() error should contain %q, got %v", tt.errContains, err)
			}
		})
	}
}
