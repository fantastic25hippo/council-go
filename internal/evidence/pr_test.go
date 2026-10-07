package evidence

import (
	"errors"
	"strings"
	"testing"
)

type fakeGHRunner struct {
	lastArgv []string
	stdout   string
	stderr   string
	exit     int
	err      error
}

func (f *fakeGHRunner) Run(argv []string) (stdout, stderr string, exit int, err error) {
	f.lastArgv = argv
	return f.stdout, f.stderr, f.exit, f.err
}

func TestPRDiff_Success(t *testing.T) {
	fake := &fakeGHRunner{
		stdout: "diff contents",
		exit:   0,
		err:    nil,
	}

	stdout, err := PRDiff(fake, "8")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if stdout != "diff contents" {
		t.Errorf("expected stdout 'diff contents', got '%s'", stdout)
	}
	if len(fake.lastArgv) != 4 || fake.lastArgv[0] != "gh" || fake.lastArgv[1] != "pr" || fake.lastArgv[2] != "diff" || fake.lastArgv[3] != "8" {
		t.Errorf("unexpected argv: %v", fake.lastArgv)
	}
}

func TestPRChecks_Success(t *testing.T) {
	fake := &fakeGHRunner{
		stdout: "checks passed",
		exit:   0,
		err:    nil,
	}

	stdout, err := PRChecks(fake, "8")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if stdout != "checks passed" {
		t.Errorf("expected stdout 'checks passed', got '%s'", stdout)
	}
	if len(fake.lastArgv) != 4 || fake.lastArgv[0] != "gh" || fake.lastArgv[1] != "pr" || fake.lastArgv[2] != "checks" || fake.lastArgv[3] != "8" {
		t.Errorf("unexpected argv: %v", fake.lastArgv)
	}
}

func TestPRDiff_EmptyPR(t *testing.T) {
	fake := &fakeGHRunner{}
	_, err := PRDiff(fake, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "PR") {
		t.Errorf("expected error containing 'PR', got '%v'", err)
	}
}

func TestPRChecks_EmptyPR(t *testing.T) {
	fake := &fakeGHRunner{}
	_, err := PRChecks(fake, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "PR") {
		t.Errorf("expected error containing 'PR', got '%v'", err)
	}
}

func TestPRDiff_GHExit(t *testing.T) {
	fake := &fakeGHRunner{
		exit: 1,
	}
	_, err := PRDiff(fake, "8")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "gh") {
		t.Errorf("expected error containing 'gh', got '%v'", err)
	}
}

func TestPRChecks_GHExit(t *testing.T) {
	fake := &fakeGHRunner{
		exit: 1,
	}
	_, err := PRChecks(fake, "8")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "gh") {
		t.Errorf("expected error containing 'gh', got '%v'", err)
	}
}

func TestPRDiff_RunnerErr(t *testing.T) {
	fake := &fakeGHRunner{
		err: errors.New("some runner error"),
	}
	_, err := PRDiff(fake, "8")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "some runner error" {
		t.Errorf("expected error 'some runner error', got '%v'", err)
	}
}

func TestPRChecks_RunnerErr(t *testing.T) {
	fake := &fakeGHRunner{
		err: errors.New("some runner error"),
	}
	_, err := PRChecks(fake, "8")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "some runner error" {
		t.Errorf("expected error 'some runner error', got '%v'", err)
	}
}
