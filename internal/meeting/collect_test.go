package meeting

import (
	"errors"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/attach"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestCollect_JulesEmptyPRURL(t *testing.T) {
	tests := []struct {
		name string
		sess *Session
	}{
		{
			name: "ResolvedWorker jules",
			sess: &Session{ResolvedWorker: "jules"},
		},
		{
			name: "Council WorkerID jules",
			sess: &Session{Council: &state.Council{WorkerID: "jules"}},
		},
		{
			name: "Attach SessionID non-empty",
			sess: &Session{Attach: attach.Request{SessionID: "sess-123"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := worker.Run{Raw: worker.Raw{PRURL: ""}}
			ev := tt.sess.collect(run)
			if ev.Diff != "" {
				t.Errorf("expected empty Diff, got %q", ev.Diff)
			}
			if ev.VerifyOutput != "no PR found" {
				t.Errorf("expected 'no PR found', got %q", ev.VerifyOutput)
			}
			if ev.VerifyExit != 1 {
				t.Errorf("expected VerifyExit 1, got %d", ev.VerifyExit)
			}
		})
	}
}

func TestCollect_PRURL_PlanVerifyEmpty_StubUnused(t *testing.T) {
	orig := verifyPRHead
	defer func() { verifyPRHead = orig }()
	verifyPRHead = func(cwd, prURL, cmd string, secs int) (string, int, error) {
		t.Fatal("verifyPRHead should not be called when Plan.Verify is empty")
		return "", 0, nil
	}

	fake := &fakeGHRunnerForWork{diffOut: "diff ok", chkOut: "checks ok", chkExit: 0}
	s := &Session{GH: fake, Plan: plan.Block{Verify: ""}}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if ev.VerifyExit != 0 {
		t.Errorf("expected exit 0, got %d", ev.VerifyExit)
	}
	if ev.VerifyOutput != "checks ok" {
		t.Errorf("expected 'checks ok', got %q", ev.VerifyOutput)
	}
}

func TestCollect_PRURL_PlanVerifyRuns(t *testing.T) {
	orig := verifyPRHead
	defer func() { verifyPRHead = orig }()
	called := false
	verifyPRHead = func(cwd, prURL, cmd string, secs int) (string, int, error) {
		called = true
		if cmd != "echo hi" {
			t.Errorf("expected cmd 'echo hi', got %q", cmd)
		}
		return "ok\n", 0, nil
	}

	fake := &fakeGHRunnerForWork{diffOut: "diff ok", chkOut: "checks ok", chkExit: 0}
	s := &Session{GH: fake, Plan: plan.Block{Verify: "echo hi"}}
	run := worker.Run{Raw: worker.Raw{PRURL: "https://github.com/org/repo/pull/12"}}
	ev := s.collect(run)
	if !called {
		t.Fatal("expected verifyPRHead to be called")
	}
	if !strings.Contains(ev.VerifyOutput, "VERIFY (local on PR head)") {
		t.Errorf("expected VERIFY heading, got %q", ev.VerifyOutput)
	}
	if !strings.Contains(ev.VerifyOutput, "ok") {
		t.Errorf("expected ok in output, got %q", ev.VerifyOutput)
	}
	if ev.VerifyExit != 0 {
		t.Errorf("expected exit 0, got %d", ev.VerifyExit)
	}
}

func TestCollect_PRURL_PlanVerifyError(t *testing.T) {
	orig := verifyPRHead
	defer func() { verifyPRHead = orig }()
	verifyPRHead = func(cwd, prURL, cmd string, secs int) (string, int, error) {
		return "", 0, errors.New("worktree fetch failed")
	}

	fake := &fakeGHRunnerForWork{diffOut: "diff ok", chkOut: "checks ok", chkExit: 0}
	s := &Session{GH: fake, Plan: plan.Block{Verify: "echo hi"}}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if !strings.Contains(ev.VerifyOutput, "(skipped: worktree fetch failed)") {
		t.Errorf("expected skipped message, got %q", ev.VerifyOutput)
	}
	if ev.VerifyExit != 0 {
		t.Errorf("expected exit to remain unchanged (0), got %d", ev.VerifyExit)
	}
}

func TestCollect_PRURL_PlanVerifyNonZeroExit(t *testing.T) {
	orig := verifyPRHead
	defer func() { verifyPRHead = orig }()
	verifyPRHead = func(cwd, prURL, cmd string, secs int) (string, int, error) {
		return "test failed\n", 2, nil
	}

	fake := &fakeGHRunnerForWork{diffOut: "diff ok", chkOut: "checks ok", chkExit: 0}
	s := &Session{GH: fake, Plan: plan.Block{Verify: "echo hi"}}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if ev.VerifyExit != 2 {
		t.Errorf("expected VerifyExit 2, got %d", ev.VerifyExit)
	}
	if !strings.Contains(ev.VerifyOutput, "test failed") {
		t.Errorf("expected test failure output, got %q", ev.VerifyOutput)
	}
}
