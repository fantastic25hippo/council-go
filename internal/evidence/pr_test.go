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

	stdout, err := PRDiff(fake, "8", "")
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

	chk, err := PRChecks(fake, "8", "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if chk.Out != "checks passed" {
		t.Errorf("expected stdout 'checks passed', got '%s'", chk.Out)
	}
	if chk.Exit != 0 || chk.Pending || chk.Failed {
		t.Errorf("unexpected checks state: %+v", chk)
	}
	if len(fake.lastArgv) != 4 || fake.lastArgv[0] != "gh" || fake.lastArgv[1] != "pr" || fake.lastArgv[2] != "checks" || fake.lastArgv[3] != "8" {
		t.Errorf("unexpected argv: %v", fake.lastArgv)
	}
}

func TestPRDiff_EmptyPR(t *testing.T) {
	fake := &fakeGHRunner{}
	if _, err := PRDiff(fake, "", ""); err == nil || !strings.Contains(err.Error(), "PR") {
		t.Errorf("expected error containing 'PR', got '%v'", err)
	}
}

func TestPRChecks_EmptyPR(t *testing.T) {
	fake := &fakeGHRunner{}
	if _, err := PRChecks(fake, "", ""); err == nil || !strings.Contains(err.Error(), "PR") {
		t.Errorf("expected error containing 'PR', got '%v'", err)
	}
}

func TestPRDiff_GHExit(t *testing.T) {
	fake := &fakeGHRunner{exit: 1}
	if _, err := PRDiff(fake, "8", ""); err == nil || !strings.Contains(err.Error(), "gh") {
		t.Errorf("expected error containing 'gh', got '%v'", err)
	}
}

func TestPRChecks_Exit8_Pending(t *testing.T) {
	fake := &fakeGHRunner{stdout: "pending run", exit: 8}
	chk, err := PRChecks(fake, "8", "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if chk.Out != "pending run" || chk.Exit != 8 || !chk.Pending || chk.Failed {
		t.Errorf("unexpected checks result: %+v", chk)
	}
}

func TestPRChecks_Exit1_Failed(t *testing.T) {
	fake := &fakeGHRunner{stdout: "failed tests", exit: 1}
	chk, err := PRChecks(fake, "8", "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if chk.Out != "failed tests" || chk.Exit != 1 || chk.Pending || !chk.Failed {
		t.Errorf("unexpected checks result: %+v", chk)
	}
}

func TestPRChecks_Exit1_NoChecks(t *testing.T) {
	fake := &fakeGHRunner{stdout: "no checks reported on commit", exit: 1}
	chk, err := PRChecks(fake, "8", "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if chk.Failed {
		t.Errorf("expected Failed false when no checks reported, got %+v", chk)
	}
}

func TestPR_RepoAndURL(t *testing.T) {
	fake := &fakeGHRunner{}
	_, _ = PRDiff(fake, "8", "owner/repo")
	if len(fake.lastArgv) != 6 || fake.lastArgv[4] != "--repo" || fake.lastArgv[5] != "owner/repo" {
		t.Errorf("unexpected diff argv: %v", fake.lastArgv)
	}
	_, _ = PRChecks(fake, "https://github.com/myorg/myrepo/pull/42", "")
	if len(fake.lastArgv) != 6 || fake.lastArgv[4] != "--repo" || fake.lastArgv[5] != "myorg/myrepo" {
		t.Errorf("unexpected checks argv: %v", fake.lastArgv)
	}
}

func TestPRDiff_RunnerErr(t *testing.T) {
	fake := &fakeGHRunner{err: errors.New("some runner error")}
	if _, err := PRDiff(fake, "8", ""); err == nil || err.Error() != "some runner error" {
		t.Errorf("expected 'some runner error', got '%v'", err)
	}
}

func TestPRChecks_RunnerErr(t *testing.T) {
	fake := &fakeGHRunner{err: errors.New("some runner error")}
	if _, err := PRChecks(fake, "8", ""); err == nil || err.Error() != "some runner error" {
		t.Errorf("expected 'some runner error', got '%v'", err)
	}
}
