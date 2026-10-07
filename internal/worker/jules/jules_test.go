package jules

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type fakeSession struct {
	state   string
	msg     string
	id      string
	prURL   string
	answers []string
}

func (f *fakeSession) State() string            { return f.state }
func (f *fakeSession) LastAgentMessage() string { return f.msg }
func (f *fakeSession) ID() string               { return f.id }
func (f *fakeSession) PRURL() string            { return f.prURL }

func (f *fakeSession) Record(answer string) {
	f.answers = append(f.answers, answer)
}

func TestWorker_Capabilities(t *testing.T) {
	w := New(&fakeSession{})
	caps := w.Capabilities()

	if !caps.CanAskAndWait {
		t.Errorf("expected CanAskAndWait true")
	}
	if !caps.KeepsMemoryAcrossFixes {
		t.Errorf("expected KeepsMemoryAcrossFixes true")
	}
	if !caps.ReachesGitHub {
		t.Errorf("expected ReachesGitHub true")
	}
	if caps.TouchesLocalCheckout {
		t.Errorf("expected TouchesLocalCheckout false")
	}
	if !caps.OpensPR {
		t.Errorf("expected OpensPR true")
	}
}

func TestWorker_Start(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		msg      string
		expected worker.Outcome
		reqMsg   string
		reqQ     string
	}{
		{
			name:     "State COMPLETED -> DONE",
			state:    "COMPLETED",
			msg:      "all done",
			expected: worker.OutcomeDone,
			reqMsg:   "all done",
		},
		{
			name:     "A question in COMPLETED does not make it ASKED",
			state:    "COMPLETED",
			msg:      "all done?",
			expected: worker.OutcomeDone,
			reqMsg:   "all done?",
		},
		{
			name:     "State AWAITING_USER_FEEDBACK -> ASKED",
			state:    "AWAITING_USER_FEEDBACK",
			msg:      "which branch?",
			expected: worker.OutcomeAsked,
			reqMsg:   "which branch?",
			reqQ:     "which branch?",
		},
		{
			name:     "State FAILED -> FAILED",
			state:    "FAILED",
			msg:      "crash",
			expected: worker.OutcomeFailed,
			reqMsg:   "crash",
		},
		{
			name:     "State IN_PROGRESS -> FAILED with unusable state",
			state:    "IN_PROGRESS",
			msg:      "working",
			expected: worker.OutcomeFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := &fakeSession{
				state: tt.state,
				msg:   tt.msg,
				id:    "id-123",
				prURL: "pr-url",
			}
			w := New(sess)

			run := w.Start(worker.Job{})

			if run.Outcome != tt.expected {
				t.Errorf("expected outcome %v, got %v", tt.expected, run.Outcome)
			}
			if run.Raw.SessionID != "id-123" {
				t.Errorf("expected SessionID id-123, got %q", run.Raw.SessionID)
			}
			if run.Raw.PRURL != "pr-url" {
				t.Errorf("expected PRURL pr-url, got %q", run.Raw.PRURL)
			}

			if tt.state == "IN_PROGRESS" {
				if !strings.Contains(run.Report, "unusable state") {
					t.Errorf("expected report to contain 'unusable state', got %q", run.Report)
				}
			} else {
				if run.Report != tt.reqMsg {
					t.Errorf("expected report %q, got %q", tt.reqMsg, run.Report)
				}
			}

			if run.Question != tt.reqQ {
				t.Errorf("expected question %q, got %q", tt.reqQ, run.Question)
			}
		})
	}
}

func TestWorker_Continue(t *testing.T) {
	sess := &fakeSession{
		state: "COMPLETED",
		msg:   "done now",
	}
	w := New(sess)

	run := w.Continue(worker.Job{}, "my answer")

	if run.Outcome != worker.OutcomeDone {
		t.Errorf("expected outcome DONE, got %v", run.Outcome)
	}

	if len(sess.answers) != 1 || sess.answers[0] != "my answer" {
		t.Errorf("expected recorded answer 'my answer', got %v", sess.answers)
	}
}
