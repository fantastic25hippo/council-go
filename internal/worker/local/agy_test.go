package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAgyRunnerArgvShape(t *testing.T) {
	r := &AgyRunner{Model: "gpt", Effort: "high", Cwd: "/workspace", Timeout: 30 * time.Second}
	job := worker.Job{Text: "hello"}
	argv := r.argv(job, "/tmp/prompt.txt")
	expected := []string{
		"grok-agy-run", "job",
		"--prompt-file", "/tmp/prompt.txt",
		"--workdir", "/workspace",
		"--tools", "code",
		"--effort", "high",
		"--model", "gpt",
		"--wait-secs", "30",
	}
	for _, a := range argv {
		if strings.Contains(a, "dangerously") || a == "-p" || a == "--output-format" || a == "--conversation" || a == "--add-dir" {
			t.Fatalf("forbidden arg found in argv: %v", argv)
		}
	}
	foundName := false
	for i, a := range argv {
		if a == "--name" && i+1 < len(argv) && strings.HasPrefix(argv[i+1], "council-") {
			foundName = true
		}
	}
	if !foundName {
		t.Fatalf("argv missing --name council-<id>: %v", argv)
	}
	for i := 2; i < len(expected); i += 2 {
		val := expected[i+1]
		found := false
		for j, a := range argv {
			if a == expected[i] && j+1 < len(argv) && argv[j+1] == val {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("argv missing %s %s: %v", expected[i], val, argv)
		}
	}
}

func TestAgyRunnerFakeScript(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-grok-agy-run")
	jobDir := filepath.Join(dir, "job-1")
	sh := fmt.Sprintf(`#!/bin/sh
mkdir -p %q
echo "job: %s"
cat > %q << 'EOF'
{"status": "ok", "final_summary": "done first"}
EOF
echo "$@" > %q
while [ $# -gt 0 ]; do
  if [ "$1" = "--prompt-file" ]; then
    cat "$2" > %q
    echo "$2" > %q
    shift 2
  else
    shift
  fi
done
`, jobDir, jobDir, filepath.Join(jobDir, "report.json"), filepath.Join(dir, "argv.txt"), filepath.Join(dir, "prompt.txt"), filepath.Join(dir, "prompt_path.txt"))
	if err := os.WriteFile(script, []byte(sh), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	summary, _, exit, err := r.Run(worker.Job{Text: "first run"})
	if err != nil || exit != 0 || summary != "done first" {
		t.Fatalf("Run err=%v exit=%d summary=%q", err, exit, summary)
	}
	recordedArgv, _ := os.ReadFile(filepath.Join(dir, "argv.txt"))
	for _, tok := range strings.Fields(string(recordedArgv)) {
		if strings.Contains(tok, "dangerously") || tok == "-p" || tok == "--output-format" || tok == "--conversation" || tok == "--add-dir" {
			t.Fatalf("forbidden arg recorded in argv: %s (token %q)", recordedArgv, tok)
		}
	}
	recordedPrompt, _ := os.ReadFile(filepath.Join(dir, "prompt.txt"))
	if strings.TrimSpace(string(recordedPrompt)) != "first run" {
		t.Fatalf("prompt file content %q", string(recordedPrompt))
	}
	recordedPath, _ := os.ReadFile(filepath.Join(dir, "prompt_path.txt"))
	if _, err := os.Stat(strings.TrimSpace(string(recordedPath))); !os.IsNotExist(err) {
		t.Fatalf("prompt file was not deleted after Run: %v", err)
	}

	// Continue test: later Run's prompt file starts with previous final_summary
	_, _, _, err = r.Run(worker.Job{Text: "second run"})
	if err != nil {
		t.Fatal(err)
	}
	recordedPrompt2, _ := os.ReadFile(filepath.Join(dir, "prompt.txt"))
	if !strings.HasPrefix(string(recordedPrompt2), "done first\n\nsecond run") {
		t.Fatalf("continue prompt should start with previous final_summary, got: %q", string(recordedPrompt2))
	}
}

func TestAgyRunnerFakeScriptFailedStatus(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-grok-agy-run")
	jobDir := filepath.Join(dir, "job-fail")
	sh := fmt.Sprintf(`#!/bin/sh
mkdir -p %q
echo "job: %s"
cat > %q << 'EOF'
{"status": "failed", "final_summary": "agent failure"}
EOF
`, jobDir, jobDir, filepath.Join(jobDir, "report.json"))
	if err := os.WriteFile(script, []byte(sh), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	summary, _, _, err := r.Run(worker.Job{Text: "job"})
	if err == nil {
		t.Fatalf("expected error on status != ok")
	}
	if summary != "agent failure" {
		t.Fatalf("expected summary 'agent failure', got %q", summary)
	}
}

func TestAgyRunnerMissingJobLine(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-grok-agy-run")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho no-job-line\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	_, _, _, err := r.Run(worker.Job{Text: "job"})
	if err == nil {
		t.Fatalf("expected error on missing job: line")
	}
}

func TestAgyRunnerTimeout(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "sleep-agy")
	pgidFile := filepath.Join(dir, "pgid.txt")
	sh := fmt.Sprintf(`#!/bin/sh
ps -o pgid= -p $$ > %q
sleep 30 &
wait
`, pgidFile)
	if err := os.WriteFile(script, []byte(sh), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 200 * time.Millisecond}
	_, _, _, err := r.Run(worker.Job{Text: "job"})
	if err != ErrTimeout {
		t.Fatalf("want ErrTimeout, got %v", err)
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
		t.Fatalf("process group %d still exists after timeout", pgid)
	}
}

func TestWorkerStartTimeoutOutcome(t *testing.T) {
	w := New(&timeoutRunner{})
	run := w.Start(worker.Job{})
	if run.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("outcome %v", run.Outcome)
	}
}

type timeoutRunner struct{}

func (timeoutRunner) Run(worker.Job) (string, string, int, error) {
	return "", "", 1, ErrTimeout
}
