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

func TestParsePlanMultiLine(t *testing.T) {
	text := `WORKER: local
CHANGE:
line 1
line 2
VERIFY:
echo hi
echo there
AGREED`
	p, ok := ParsePlan(text)
	if !ok {
		t.Fatal("parse failed")
	}
	if p.Worker != "local" {
		t.Fatalf("worker = %q", p.Worker)
	}
	if p.Change != "line 1\nline 2" {
		t.Fatalf("change = %q, want %q", p.Change, "line 1\nline 2")
	}
	if p.Verify != "echo hi\necho there" {
		t.Fatalf("verify = %q, want %q", p.Verify, "echo hi\necho there")
	}

	text2 := `WORKER: local
CHANGE: first line
second line
VERIFY: verify line 1
verify line 2
AGREED`
	p2, ok2 := ParsePlan(text2)
	if !ok2 {
		t.Fatal("parse failed")
	}
	if p2.Change != "first line\nsecond line" {
		t.Fatalf("change = %q", p2.Change)
	}
	if p2.Verify != "verify line 1\nverify line 2" {
		t.Fatalf("verify = %q", p2.Verify)
	}
}

func TestSplitField(t *testing.T) {
	tests := []struct {
		line    string
		wantKey string
		wantRes string
		wantOk  bool
	}{
		{"", "", "", false},
		{"   ", "", "", false},
		{"NO_COLON_HERE", "", "", false},
		{":leading_colon", "", "", false},
		{"  :leading_with_space", "", "", false},
		{"*:*", "", "", false},
		{"UNKNOWN: val", "", "", false},
		{"WORKER: local", "WORKER", "local", true},
		{"worker: local", "WORKER", "local", true},
		{"**WORKER:** local", "WORKER", "local", true},
		{"**REPO**: github.com/org/repo", "REPO", "github.com/org/repo", true},
		{"CHANGE:", "CHANGE", "", true},
		{"  CHANGE:   something done  ", "CHANGE", "something done", true},
		{"VERIFY: echo 1:2:3", "VERIFY", "echo 1:2:3", true},
		{"OWNED_PATHS: **path/to/file**", "OWNED_PATHS", "path/to/file", true},
	}

	for _, tt := range tests {
		gotKey, gotRest, gotOk := splitField(tt.line)
		if gotKey != tt.wantKey || gotRest != tt.wantRes || gotOk != tt.wantOk {
			t.Errorf("splitField(%q) = (%q, %q, %v); want (%q, %q, %v)",
				tt.line, gotKey, gotRest, gotOk, tt.wantKey, tt.wantRes, tt.wantOk)
		}
	}
}
