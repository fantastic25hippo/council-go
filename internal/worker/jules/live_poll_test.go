package jules

import (
	"testing"
	"time"
)

func TestLivePollNilContext(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"IN_PROGRESS"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
	w := NewLive(&Client{API: api}, time.Hour)
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	w.live.sess = &APISession{C: w.live.Client, SID: "s1"}

	err := w.live.poll(nil)
	if err != nil {
		t.Fatalf("expected nil error from poll, got %v", err)
	}
}
