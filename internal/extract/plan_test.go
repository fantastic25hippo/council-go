package extract

import (
	"reflect"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/plan"
)

func TestParsePlanBasic(t *testing.T) {
	text := `WORKER: local
REPO: council-go
OWNED_PATHS:
- internal/foo
- internal/bar
CHANGE: add a thing
VERIFY: echo hi
AGREED`
	p, ok := ParsePlan(text)
	if !ok {
		t.Fatal("expected ok")
	}
	want := plan.Block{
		Worker:     "local",
		Repo:       "council-go",
		OwnedPaths: []string{"internal/foo", "internal/bar"},
		Change:     "add a thing",
		Verify:     "echo hi",
	}
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("got %+v, want %+v", p, want)
	}
}

func TestParsePlanMarkdownAndCase(t *testing.T) {
	text := `**WORKER:** Jules
**Repo:** my-repo
*CHANGE:* implement feature
*VERIFY:* run tests
**AGREED**`
	p, ok := ParsePlan(text)
	if !ok {
		t.Fatal("expected ok")
	}
	if p.Worker != "jules" {
		t.Fatalf("worker = %q, want %q", p.Worker, "jules")
	}
	if p.Repo != "my-repo" {
		t.Fatalf("repo = %q, want %q", p.Repo, "my-repo")
	}
	if p.Change != "implement feature" {
		t.Fatalf("change = %q, want %q", p.Change, "implement feature")
	}
	if p.Verify != "run tests" {
		t.Fatalf("verify = %q, want %q", p.Verify, "run tests")
	}
}

func TestParsePlanOwnedPathsVariants(t *testing.T) {
	// Case 1: single inline path
	textInline := `WORKER: local
OWNED_PATHS: internal/baz
CHANGE: test
VERIFY: test
AGREED`
	p, ok := ParsePlan(textInline)
	if !ok {
		t.Fatal("expected ok")
	}
	if !reflect.DeepEqual(p.OwnedPaths, []string{"internal/baz"}) {
		t.Fatalf("owned paths = %v", p.OwnedPaths)
	}

	// Case 2: inline path plus list items, followed by another field
	textBoth := `WORKER: local
OWNED_PATHS: cmd/main.go
- internal/foo.go
- internal/bar.go
CHANGE: test
VERIFY: test
AGREED`
	p2, ok2 := ParsePlan(textBoth)
	if !ok2 {
		t.Fatal("expected ok")
	}
	wantPaths := []string{"cmd/main.go", "internal/foo.go", "internal/bar.go"}
	if !reflect.DeepEqual(p2.OwnedPaths, wantPaths) {
		t.Fatalf("owned paths = %v, want %v", p2.OwnedPaths, wantPaths)
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

	text2 := "WORKER: local\r\nCHANGE: first line\r\nsecond line\r\nVERIFY: verify line 1\r\nverify line 2\r\nAGREED\r\n"
	p2, ok2 := ParsePlan(text2)
	if !ok2 {
		t.Fatal("parse failed with CRLF")
	}
	if p2.Change != "first line\nsecond line" {
		t.Fatalf("change = %q", p2.Change)
	}
	if p2.Verify != "verify line 1\nverify line 2" {
		t.Fatalf("verify = %q", p2.Verify)
	}
}

func TestParsePlanInvalid(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"no AGREED", "WORKER: local\nCHANGE: x\nVERIFY: echo hi"},
		{"missing CHANGE", "WORKER: local\nVERIFY: echo hi\nAGREED"},
		{"missing VERIFY", "WORKER: local\nCHANGE: x\nAGREED"},
		{"missing WORKER", "CHANGE: x\nVERIFY: echo hi\nAGREED"},
		{"question mark agreed", "WORKER: local\nCHANGE: x\nVERIFY: echo hi\nAGREED?"},
		{"empty string", ""},
		{"unrelated text with AGREED", "Just some conversation\nAGREED"},
		{"unknown field ignored", "UNKNOWN: foo\nWORKER: local\nCHANGE: x\nVERIFY: v\nAGREED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := ParsePlan(tt.text)
			if tt.name == "unknown field ignored" {
				if !ok || p.Worker != "local" {
					t.Fatalf("expected valid parse ignoring unknown field, got ok=%v, p=%+v", ok, p)
				}
			} else {
				if ok {
					t.Fatalf("expected false, got %+v", p)
				}
			}
		})
	}
}
