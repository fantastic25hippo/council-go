package prompt

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestLocalJob(t *testing.T) {
	p := plan.Block{Worker: "w", Change: "mychange", Verify: "myverify"}
	caps := worker.Capabilities{}
	out := LocalJob(p, caps)

	mustContain := []string{"COUNCIL-QUESTION:", "mychange"}
	for _, req := range mustContain {
		if !strings.Contains(out, req) {
			t.Errorf("LocalJob missing %q", req)
		}
	}

	mustNotContain := []string{"do not ask", "AWAITING_USER_FEEDBACK", "python3", "timeout 60"}
	for _, req := range mustNotContain {
		if strings.Contains(out, req) {
			t.Errorf("LocalJob should not contain %q", req)
		}
	}
}

func TestJulesJob(t *testing.T) {
	p := plan.Block{Worker: "w", Change: "mychange", Verify: "myverify"}
	caps := worker.Capabilities{}
	out := JulesJob(p, caps)

	mustContain := []string{"wait", "mychange"}
	for _, req := range mustContain {
		if !strings.Contains(out, req) {
			t.Errorf("JulesJob missing %q", req)
		}
	}

	mustNotContain := []string{"do not ask", "COUNCIL-QUESTION:"}
	for _, req := range mustNotContain {
		if strings.Contains(out, req) {
			t.Errorf("JulesJob should not contain %q", req)
		}
	}
}

func TestVerifyReview(t *testing.T) {
	ev := worker.Evidence{Diff: "d", VerifyOutput: "vo", VerifyExit: 42}
	claim := "c"
	out := VerifyReview(ev, claim)

	mustContain := []string{"worker's claim, unverified", "42"}
	for _, req := range mustContain {
		if !strings.Contains(out, req) {
			t.Errorf("VerifyReview missing %q", req)
		}
	}

	mustNotContain := []string{"python3", "timeout 60"}
	for _, req := range mustNotContain {
		if strings.Contains(out, req) {
			t.Errorf("VerifyReview should not contain %q", req)
		}
	}
}
