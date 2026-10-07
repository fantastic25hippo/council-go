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
}
