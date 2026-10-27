package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/proc"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAgyRunnerCancelRace(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "sleep-agy")
	pgidFile := filepath.Join(dir, "pgid.txt")
	sh := fmt.Sprintf(`#!/bin/sh
echo $$ > %q
sleep 30 &
wait
`, pgidFile)
	if err := os.WriteFile(script, []byte(sh), 0755); err != nil {
		t.Fatal(err)
	}

	r := &AgyRunner{
		CLI:     script,
		Cwd:     dir,
		Timeout: 5 * time.Second,
	}

	r.afterStart = func() {
		for i := 0; i < 50; i++ {
			if _, err := os.Stat(pgidFile); err == nil {
				break
			}
			time.Sleep(2 * time.Millisecond)
		}
		if err := r.Cancel(); err != nil {
			t.Errorf("Cancel error: %v", err)
		}
	}

	_, _, exit, err := r.Run(worker.Job{Text: "test job"})
	if err == nil {
		t.Fatal("expected error on cancelled run")
	}
	if !r.IsCanceled() {
		t.Fatal("expected runner to be canceled")
	}
	if exit == 0 {
		t.Fatal("expected non-zero exit on cancel")
	}

	data, err := os.ReadFile(pgidFile)
	if err != nil {
		t.Fatalf("reading pgid file: %v", err)
	}
	var pgid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pgid); err != nil {
		t.Fatalf("parsing pgid %q: %v", string(data), err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := syscall.Kill(-pgid, 0); err == nil {
		t.Fatalf("process group %d still exists after cancel", pgid)
	}
}

func TestWorkerStartCancelRace(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "sleep-cmd")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0755); err != nil {
		t.Fatal(err)
	}

	runner := &AgyRunner{
		CLI:     script,
		Cwd:     dir,
		Timeout: 5 * time.Second,
	}
	w := New(runner)
	runner.afterStart = func() {
		_ = w.Cancel()
	}
	run := w.Start(worker.Job{Text: "test"})
	if run.Outcome != worker.OutcomeCancelled {
		t.Fatalf("expected OutcomeCancelled, got %v", run.Outcome)
	}
}

func TestAgyRunnerAttachCmdWhenCanceled(t *testing.T) {
	cmd, err := proc.Start([]string{"sleep", "30"}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{canceled: true}
	if r.attachCmd(cmd) {
		t.Fatal("expected attachCmd to return false when canceled")
	}
	_ = cmd.Wait()
}
