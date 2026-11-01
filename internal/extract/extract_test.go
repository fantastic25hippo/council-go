package extract

import "testing"

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
