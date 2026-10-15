package jules

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type blockingPollDoer struct {
	inPoll chan struct{}
	once   sync.Once
}

func (b *blockingPollDoer) Do(method, path string, body []byte) (int, []byte, error) {
	if method == "POST" && path == "/sessions" {
		return 200, []byte(`{"id":"s1"}`), nil
	}
	if method == "GET" && path == "/sessions/s1" {
		return 200, []byte(`{"id":"s1","state":"IN_PROGRESS"}`), nil
	}
	if method == "GET" && strings.HasPrefix(path, "/sessions/s1/activities") {
		b.once.Do(func() {
			close(b.inPoll)
		})
		return 200, []byte(`{}`), nil
	}
	return 200, []byte(`{}`), nil
}

func TestLiveStartCancelOverlap(t *testing.T) {
	for i := 0; i < 10; i++ {
		doer := &blockingPollDoer{inPoll: make(chan struct{})}
		w := NewLive(&Client{API: doer}, time.Hour)
		w.live.Sleep = func(time.Duration) {}
		w.live.Interval = time.Hour

		done := make(chan worker.Run, 1)
		go func() {
			done <- w.Start(worker.Job{Repo: "o/r", Text: "do it"})
		}()

		<-doer.inPoll

		if err := w.Cancel(); err != nil {
			t.Fatalf("cancel failed: %v", err)
		}

		select {
		case run := <-done:
			if run.Outcome != worker.OutcomeCancelled {
				t.Fatalf("expected OutcomeCancelled, got %v", run.Outcome)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for Start to cancel")
		}
	}
}
