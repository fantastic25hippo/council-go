package meeting_test

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestPrepare(t *testing.T) {
	t.Run("conflicting WORKER-OVERRIDE errors, no worker id", func(t *testing.T) {
		task := "WORKER-OVERRIDE: local\nWORKER-OVERRIDE: jules\n"
		var p plan.Block
		id, err := meeting.Prepare(task, p, "local", false, worker.Capabilities{}, false, false)
		if err == nil {
			t.Fatal("expected error for conflicting overrides")
		}
		if id != "" {
			t.Fatalf("expected empty worker id, got %q", id)
		}
	})

	t.Run("Prepare with override local, matching plan, local caps, empty repo, Verify `echo hi`, dirty false", func(t *testing.T) {
		task := "WORKER-OVERRIDE: local"
		p := plan.Block{
			Worker: "local",
			Verify: "echo hi",
		}
		caps := worker.Capabilities{
			TouchesLocalCheckout: true,
		}
		id, err := meeting.Prepare(task, p, "local", false, caps, false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "local" {
			t.Fatalf("expected worker id 'local', got %q", id)
		}
	})

	t.Run("GitHub caps, unpushed true", func(t *testing.T) {
		task := ""
		p := plan.Block{
			Worker: "jules",
			Verify: "echo hi",
			Repo:   "fantastic25hippo/council-go",
		}
		caps := worker.Capabilities{
			ReachesGitHub: true,
		}
		id, err := meeting.Prepare(task, p, "jules", false, caps, false, true)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "unpushed") {
			t.Fatalf("expected error containing 'unpushed', got: %v", err)
		}
		if id != "" {
			t.Fatalf("expected empty id, got %q", id)
		}
	})

	t.Run("Protected owned path hook.log", func(t *testing.T) {
		task := ""
		p := plan.Block{
			Worker:     "local",
			Verify:     "echo hi",
			OwnedPaths: []string{"hook.log"},
		}
		caps := worker.Capabilities{
			TouchesLocalCheckout: true,
		}
		id, err := meeting.Prepare(task, p, "local", false, caps, false, false)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "protected") {
			t.Fatalf("expected error containing 'protected', got: %v", err)
		}
		if id != "" {
			t.Fatalf("expected empty id, got %q", id)
		}
	})

	t.Run("Validate mismatch WORKER", func(t *testing.T) {
		task := ""
		p := plan.Block{
			Worker: "jules",
			Verify: "echo hi",
		}
		caps := worker.Capabilities{
			TouchesLocalCheckout: true,
		}
		id, err := meeting.Prepare(task, p, "local", false, caps, false, false)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "NACK") {
			t.Fatalf("expected error containing 'NACK', got: %v", err)
		}
		if id != "" {
			t.Fatalf("expected empty id, got %q", id)
		}
	})

	t.Run("p is not mutated", func(t *testing.T) {
		task := ""
		p := plan.Block{
			Worker:     "local",
			Verify:     "echo hi",
			OwnedPaths: []string{"a.go"},
		}
		pCopy := plan.Block{
			Worker:     "local",
			Verify:     "echo hi",
			OwnedPaths: []string{"a.go"},
		}
		caps := worker.Capabilities{
			TouchesLocalCheckout: true,
		}
		meeting.Prepare(task, p, "local", false, caps, false, false)

		if p.Worker != pCopy.Worker || p.Verify != pCopy.Verify || len(p.OwnedPaths) != len(pCopy.OwnedPaths) {
			t.Fatalf("p was mutated")
		}
		for i := range p.OwnedPaths {
			if p.OwnedPaths[i] != pCopy.OwnedPaths[i] {
				t.Fatalf("p was mutated")
			}
		}
	})
}
