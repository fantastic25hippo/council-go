package jules

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type createReq struct {
	SourceContext struct {
		GithubRepoContext struct {
			StartingBranch string `json:"startingBranch"`
		} `json:"githubRepoContext"`
	} `json:"sourceContext"`
}

func completedSessionDoer() *seqDoer {
	return &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://pr"}}]}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
}

func parseStartingBranch(t *testing.T, api *seqDoer) string {
	t.Helper()
	if len(api.reqBodies) == 0 {
		t.Fatal("no requests made")
	}
	var req createReq
	if err := json.Unmarshal(api.reqBodies[0], &req); err != nil {
		t.Fatalf("unmarshal create request: %v", err)
	}
	return req.SourceContext.GithubRepoContext.StartingBranch
}

func TestLiveStart_DefaultBranchFromGH(t *testing.T) {
	api := completedSessionDoer()
	w := NewLive(&Client{API: api}, time.Minute)
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	runner := &fakeGHRunner{stdout: "main\n", exit: 0}
	w.live.Runner = runner

	run := w.Start(worker.Job{Repo: "owner/repo", Text: "do it", Branch: ""})
	if run.Outcome != worker.OutcomeDone {
		t.Fatalf("outcome %v report %s", run.Outcome, run.Report)
	}

	wantArgv := []string{"gh", "repo", "view", "owner/repo", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name"}
	if !reflect.DeepEqual(runner.lastArgv, wantArgv) {
		t.Errorf("argv = %v, want %v", runner.lastArgv, wantArgv)
	}

	gotBranch := parseStartingBranch(t, api)
	if gotBranch != "main" {
		t.Errorf("startingBranch = %q, want %q", gotBranch, "main")
	}
}

func TestLiveStart_GHRunnerErrorOrNonzero(t *testing.T) {
	cases := []struct {
		name   string
		runner *fakeGHRunner
	}{
		{name: "nonzero exit", runner: &fakeGHRunner{exit: 1, stderr: "repo not found"}},
		{name: "error", runner: &fakeGHRunner{err: errors.New("network failure")}},
		{name: "empty stdout", runner: &fakeGHRunner{stdout: "   \n", exit: 0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := completedSessionDoer()
			w := NewLive(&Client{API: api}, time.Minute)
			w.live.Sleep = func(time.Duration) {}
			w.live.Interval = time.Millisecond
			w.live.Runner = tc.runner

			run := w.Start(worker.Job{Repo: "owner/repo", Text: "do it"})
			if run.Outcome != worker.OutcomeDone {
				t.Fatalf("outcome %v report %s", run.Outcome, run.Report)
			}

			gotBranch := parseStartingBranch(t, api)
			if gotBranch != "master" {
				t.Errorf("startingBranch = %q, want %q", gotBranch, "master")
			}
		})
	}
}

func TestLiveStart_JobBranchWins(t *testing.T) {
	api := completedSessionDoer()
	w := NewLive(&Client{API: api}, time.Minute)
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	runner := &fakeGHRunner{stdout: "main\n", exit: 0}
	w.live.Runner = runner
	w.live.Branch = "other"

	run := w.Start(worker.Job{Repo: "owner/repo", Text: "do it", Branch: "feature"})
	if run.Outcome != worker.OutcomeDone {
		t.Fatalf("outcome %v report %s", run.Outcome, run.Report)
	}

	gotBranch := parseStartingBranch(t, api)
	if gotBranch != "feature" {
		t.Errorf("startingBranch = %q, want %q", gotBranch, "feature")
	}
}
