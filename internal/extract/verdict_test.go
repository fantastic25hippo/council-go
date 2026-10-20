package extract

import "testing"

func TestParseVerdict(t *testing.T) {
	if ParseVerdict("notes\nVERDICT: PASS\n") != "PASS" {
		t.Fatal("pass")
	}
	if ParseVerdict("VERDICT: FAIL\nVERDICT: PASS") != "PASS" {
		t.Fatal("last wins")
	}
	if ParseVerdict("nope") != "" {
		t.Fatal("empty")
	}
	if ParseVerdict("**VERDICT: PASS**") != "PASS" {
		t.Fatal("bold verdict")
	}
	if ParseVerdict("**VERDICT:** PASS") != "PASS" {
		t.Fatal("bold label")
	}
	if ParseVerdict("`VERDICT: PASS`") != "PASS" {
		t.Fatal("backtick wrapped")
	}
	if ParseVerdict("VERDICT: FAIL\n> VERDICT: PASS") != "FAIL" {
		t.Fatal("quoted line should not override")
	}
}
