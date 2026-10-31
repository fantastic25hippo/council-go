package jules

import (
	"errors"
	"strings"
	"testing"
)

type pagingDoer struct {
	calls []string
}

func (p *pagingDoer) Do(method, path string, body []byte) (int, []byte, error) {
	p.calls = append(p.calls, path)
	if len(p.calls) == 1 {
		return 200, []byte(`{"activities":[{"agentMessaged":{"agentMessage":"old"}}],"nextPageToken":"t2"}`), nil
	}
	if len(p.calls) == 2 {
		if !strings.Contains(path, "pageToken=t2") {
			return 500, []byte("expected pageToken=t2 in path"), nil
		}
		return 200, []byte(`{"activities":[{"agentMessaged":{"agentMessage":"newest"}}]}`), nil
	}
	return 500, []byte("unexpected call"), nil
}

func TestActivities_Paging(t *testing.T) {
	doer := &pagingDoer{}
	c := &Client{API: doer}

	sum, err := c.Activities("s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doer.calls) != 2 {
		t.Fatalf("expected 2 pages fetched, got %d", len(doer.calls))
	}
	if !strings.Contains(doer.calls[1], "pageToken=t2") {
		t.Fatalf("expected second path to contain pageToken=t2, got %s", doer.calls[1])
	}
	if sum.LastAgent != "newest" {
		t.Fatalf("expected LastAgent 'newest', got %q", sum.LastAgent)
	}
}

func TestActivities_Error(t *testing.T) {
	fake := &fakeDoer{err: errors.New("boom")}
	c := &Client{API: fake}

	_, err := c.Activities("s1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing 'boom', got %v", err)
	}
}

type benchDoer struct {
	calls int
}

func (b *benchDoer) Do(method, path string, body []byte) (int, []byte, error) {
	b.calls++
	if b.calls < 20 {
		return 200, []byte(`{"nextPageToken":"tok-page-` + string(rune('0'+b.calls)) + `"}`), nil
	}
	return 200, []byte(`{}`), nil
}

func BenchmarkActivities_Pagination(b *testing.B) {
	doer := &benchDoer{}
	c := &Client{API: doer}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		doer.calls = 0
		_, _ = c.Activities("session-123456789")
	}
}
