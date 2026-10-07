package route

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		p          plan.Block
		resolved   string
		caps       worker.Capabilities
		wantErrMsg string
	}{
		{
			name: "wrong Worker",
			p: plan.Block{
				Worker: "jules",
			},
			resolved:   "local",
			wantErrMsg: "WORKER",
		},
		{
			name: "wrong Worker returns NACK",
			p: plan.Block{
				Worker: "jules",
			},
			resolved:   "local",
			wantErrMsg: "NACK",
		},
		{
			name: "jules caps with empty Repo",
			p: plan.Block{
				Worker: "jules",
				Repo:   "",
			},
			resolved: "jules",
			caps: worker.Capabilities{
				ReachesGitHub: true,
			},
			wantErrMsg: "REPO",
		},
		{
			name: "jules caps with empty Repo NACK",
			p: plan.Block{
				Worker: "jules",
				Repo:   "  ",
			},
			resolved: "jules",
			caps: worker.Capabilities{
				ReachesGitHub: true,
			},
			wantErrMsg: "NACK",
		},
		{
			name: "local caps with empty Repo",
			p: plan.Block{
				Worker: "local",
				Repo:   "",
			},
			resolved: "local",
			caps: worker.Capabilities{
				ReachesGitHub: false,
				OpensPR:       false,
			},
		},
		{
			name: "valid with question mark in Change",
			p: plan.Block{
				Worker: "local",
				Change: "what?",
			},
			resolved: "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare a copy to ensure it is not modified
			origWorker := tt.p.Worker
			origRepo := tt.p.Repo
			origChange := tt.p.Change

			err := Validate(tt.p, tt.resolved, tt.caps)
			if tt.p.Worker != origWorker || tt.p.Repo != origRepo || tt.p.Change != origChange {
				t.Errorf("Validate mutated plan block")
			}

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
			}
		})
	}
}
