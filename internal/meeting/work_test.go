package meeting

import (
	"errors"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

type fakeGHRunnerForWork struct {
	argvs    [][]string
	diffOut  string
	diffExit int
	diffErr  error
	chkOut   string
	chkExit  int
	chkErr   error
}

func (f *fakeGHRunnerForWork) Run(argv []string) (string, string, int, error) {
	f.argvs = append(f.argvs, argv)
	if len(argv) >= 3 && argv[2] == "diff" {
		return f.diffOut, "", f.diffExit, f.diffErr
	}
	return f.chkOut, "", f.chkExit, f.chkErr
}

func TestCollect_PRDiffFailedNeverUsesReport(t *testing.T) {
	fake := &fakeGHRunnerForWork{
		diffExit: 1,
		diffErr:  errors.New("not found"),
	}
	s := &Session{
		GH:   fake,
		Plan: plan.Block{Repo: "org/repo"},
	}
	run := worker.Run{
		Report: "do not use this diff",
		Raw:    worker.Raw{PRURL: "123"},
	}
	ev := s.collect(run)
	if ev.Diff != "(gh pr diff failed: not found)" {
		t.Errorf("expected failed diff msg, got %q", ev.Diff)
	}
}

func TestCollect_PRChecksPending(t *testing.T) {
	fake := &fakeGHRunnerForWork{
		diffOut: "diff ok",
		chkOut:  "workflow in progress",
		chkExit: 8,
	}
	s := &Session{GH: fake}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if ev.VerifyExit != 8 {
		t.Errorf("expected exit 8, got %d", ev.VerifyExit)
	}
	if !strings.HasPrefix(ev.VerifyOutput, "checks pending\n") {
		t.Errorf("expected 'checks pending\\n' prefix, got %q", ev.VerifyOutput)
	}
}

func TestCollect_PRChecksFailedPrefix(t *testing.T) {
	fake := &fakeGHRunnerForWork{
		diffOut: "diff ok",
		chkOut:  "tests crashed",
		chkExit: 1,
	}
	s := &Session{GH: fake}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if ev.VerifyExit != 1 {
		t.Errorf("expected exit 1, got %d", ev.VerifyExit)
	}
	if ev.VerifyOutput != "checks failed\ntests crashed" {
		t.Errorf("expected prefixed failure, got %q", ev.VerifyOutput)
	}

	// When already containing "checks failed", do not double prefix
	fake.chkOut = "checks failed: build error"
	ev2 := s.collect(run)
	if ev2.VerifyOutput != "checks failed: build error" {
		t.Errorf("expected no double prefix, got %q", ev2.VerifyOutput)
	}
}

func TestCollect_PRChecksNoChecksReported(t *testing.T) {
	fake := &fakeGHRunnerForWork{
		chkOut:  "no checks reported on commit abc",
		chkExit: 1,
	}
	s := &Session{GH: fake}
	run := worker.Run{Raw: worker.Raw{PRURL: "123"}}
	ev := s.collect(run)
	if ev.VerifyExit != 1 {
		t.Errorf("expected exit 1, got %d", ev.VerifyExit)
	}
	if strings.HasPrefix(ev.VerifyOutput, "checks failed\n") {
		t.Errorf("should not prefix checks failed when no checks reported, got %q", ev.VerifyOutput)
	}
}
