package extract

import "testing"

func TestIsAgreed(t *testing.T) {
	if !IsAgreed("plan\nAGREED") {
		t.Fatal("plain AGREED")
	}
	if !IsAgreed("**AGREED**\n") {
		t.Fatal("bold")
	}
	if !IsAgreed("AGREED.") {
		t.Fatal("period")
	}
	if !IsAgreed("**AGREED.**") {
		t.Fatal("bold with period inside")
	}
	if IsAgreed("AGREED\nthen more") {
		t.Fatal("not last")
	}
	if IsAgreed("not agreed") {
		t.Fatal("false positive")
	}
	if IsAgreed("NOT AGREED") {
		t.Fatal("NOT AGREED should be false")
	}
}

func TestParseLaunch(t *testing.T) {
	if g := ParseLaunch("LAUNCH: local\nAGREED"); g != "local" {
		t.Fatalf("got %q", g)
	}
	if g := ParseLaunch("LAUNCH: worker=jules"); g != "jules" {
		t.Fatalf("got %q", g)
	}
	if ParseLaunch("no launch") != "" {
		t.Fatal("empty")
	}
}
