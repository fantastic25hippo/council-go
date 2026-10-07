package prompt

import "strings"

import "testing"

func TestSystemAndGreet(t *testing.T) {
	s := System("Grok", "Gemini", 3, 3, []string{"local", "jules"})
	if !strings.Contains(s, "Grok") || !strings.Contains(s, "AGREED") || strings.Contains(s, "python3") {
		t.Fatalf("system %q", s)
	}
	g := Greet("Grok", "Gemini")
	if !strings.Contains(g, "Grok") || strings.Contains(strings.ToLower(g), "propose") && strings.Contains(g, "Do not start proposing") {
		// must contain do not start proposing
	}
	if !strings.Contains(g, "Do not start proposing") {
		t.Fatalf("greet %q", g)
	}
}

func TestReviewAskHasVerdictInstruction(t *testing.T) {
	s := ReviewAsk("d", "out", 0, "claim")
	if !strings.Contains(s, "VERDICT: PASS") || strings.Contains(s, "python3") {
		t.Fatalf("%q", s)
	}
}
