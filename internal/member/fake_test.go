package member

import "testing"

func TestFakeSendRecordsAndReturns(t *testing.T) {
	f := &Fake{Name: "grok", Shown: "Grok", Replies: []string{"hi", "ok"}}
	if f.ID() != "grok" || f.DisplayName() != "Grok" {
		t.Fatal("id/name")
	}
	if err := f.Start("."); err != nil || f.Started != 1 {
		t.Fatal("start")
	}
	got, err := f.Send("hello")
	if err != nil || got != "hi" || f.Sent[0] != "hello" {
		t.Fatalf("send %q %v", got, err)
	}
	got, _ = f.Send("next")
	if got != "ok" {
		t.Fatalf("second %q", got)
	}
	_ = f.Close()
	if f.Closed != 1 {
		t.Fatal("close")
	}
}
