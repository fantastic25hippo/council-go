package jules

import (
	"errors"
	"reflect"
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

func TestFallbackPR(t *testing.T) {
	tests := []struct {
		name          string
		repo          string
		stdout        string
		exit          int
		err           error
		expectedPRURL string
	}{
		{
			name:          "One open PR",
			repo:          "test/repo",
			stdout:        `[{"url":"https://github.com/test/repo/pull/1","title":"test pr"}]`,
			exit:          0,
			expectedPRURL: "https://github.com/test/repo/pull/1",
		},
		{
			name:          "No open PRs",
			repo:          "test/repo",
			stdout:        `[]`,
			exit:          0,
			expectedPRURL: "",
		},
		{
			name: "Multiple open PRs",
			repo: "test/repo",
			stdout: `[
				{"url":"https://github.com/test/repo/pull/1","title":"test pr 1"},
				{"url":"https://github.com/test/repo/pull/2","title":"test pr 2"}
			]`,
			exit:          0,
			expectedPRURL: "",
		},
		{
			name:          "Command fails",
			repo:          "test/repo",
			stdout:        ``,
			exit:          1,
			expectedPRURL: "",
		},
		{
			name:          "Runner error",
			repo:          "test/repo",
			stdout:        ``,
			exit:          0,
			err:           errors.New("gh not found"),
			expectedPRURL: "",
		},
		{
			name:          "Empty repo",
			repo:          "",
			stdout:        `[{"url":"https://github.com/test/repo/pull/1","title":"test pr"}]`,
			exit:          0,
			expectedPRURL: "",
		},
		{
			name:          "Invalid json",
			repo:          "test/repo",
			stdout:        `not json`,
			exit:          0,
			expectedPRURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := &APISession{info: SessionInfo{}}
			runner := &fakeGHRunner{
				stdout: tt.stdout,
				exit:   tt.exit,
				err:    tt.err,
			}

			live := &Live{
				Repo:   tt.repo,
				Runner: runner,
				sess:   sess,
			}

			live.fallbackPR()

			if sess.PRURL() != tt.expectedPRURL {
				t.Errorf("expected PR URL %q, got %q", tt.expectedPRURL, sess.PRURL())
			}

			if tt.repo != "" && runner.lastArgv != nil {
				expectedArgv := []string{"gh", "pr", "list", "--repo", tt.repo, "--state", "open", "--json", "url,title", "--limit", "5"}
				if !reflect.DeepEqual(runner.lastArgv, expectedArgv) {
					t.Errorf("expected argv %v, got %v", expectedArgv, runner.lastArgv)
				}
			}
		})
	}
}
