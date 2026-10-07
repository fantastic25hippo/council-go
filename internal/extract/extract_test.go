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
	if IsAgreed("AGREED\nthen more") {
		t.Fatal("not last")
	}
	if IsAgreed("not agreed") {
		t.Fatal("false positive")
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

func TestParsePlan(t *testing.T) {
	text := `WORKER: local
OWNED_PATHS:
- internal/foo
CHANGE: add a thing
VERIFY: echo hi
AGREED`
	p, ok := ParsePlan(text)
	if !ok {
		t.Fatal("parse failed")
	}
	if p.Worker != "local" || p.Change != "add a thing" || p.Verify != "echo hi" {
		t.Fatalf("%+v", p)
	}
	if len(p.OwnedPaths) != 1 || p.OwnedPaths[0] != "internal/foo" {
		t.Fatalf("paths %v", p.OwnedPaths)
	}
}

func TestParsePlanRequiresAgreedAndFields(t *testing.T) {
	if _, ok := ParsePlan("WORKER: local\nCHANGE: x\nVERIFY: echo hi"); ok {
		t.Fatal("no AGREED")
	}
	if _, ok := ParsePlan("WORKER: local\nAGREED"); ok {
		t.Fatal("missing CHANGE")
	}
}

func TestParsePlanDoesNotTreatQuestionAsAgreed(t *testing.T) {
	if _, ok := ParsePlan("should we AGREED?"); ok {
		t.Fatal("question")
	}
}
