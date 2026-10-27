package local

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAgyRunnerArgvDangerouslySkipPermissions(t *testing.T) {
	r := &AgyRunner{CLI: "agy", Model: "m", Cwd: t.TempDir()}
	argv := r.argv(worker.Job{Text: "task"})

	for _, a := range argv {
		if strings.Contains(a, "dangerously-skip-permissions") || a == "-p" || a == "--output-format" {
			t.Fatalf("forbidden flag found in argv: %s", a)
		}
	}
}

func TestAgyRunnerDeniedActionsInReport(t *testing.T) {
	js := map[string]interface{}{
		"response": "did-work",
		"result": map[string]interface{}{
			"denied_actions": []interface{}{
				map[string]interface{}{"display_name": "RunCommand"},
				map[string]interface{}{"display_name": "WriteFile"},
			},
		},
	}
	report := parseReport(js, "", "")
	if !strings.Contains(report, "did-work") {
		t.Fatalf("expected response in report, got %q", report)
	}
	if !strings.Contains(report, "RunCommand") || !strings.Contains(report, "WriteFile") {
		t.Fatalf("expected denied actions in report, got %q", report)
	}
}

func TestAgyRunnerErrorAndStderrInReport(t *testing.T) {
	js := map[string]interface{}{
		"status": "ERROR",
		"error":  "agent failure message",
	}
	report := parseReport(js, "", "error details on stderr")
	if !strings.Contains(report, "agent failure message") {
		t.Fatalf("expected JSON error in report, got %q", report)
	}
	if !strings.Contains(report, "error details on stderr") {
		t.Fatalf("expected stderr in report, got %q", report)
	}
	if !strings.Contains(report, "status: ERROR") {
		t.Fatalf("expected status: ERROR in report, got %q", report)
	}
}

func TestParseReportTruncatesStderrValidUTF8(t *testing.T) {
	prefix := strings.Repeat("a", 2046)
	emoji := "😃" // 4 bytes: 0xF0 0x9F 0x98 0x86
	suffix := strings.Repeat("x", 2045)
	rawStderr := prefix + emoji + suffix

	out := parseReport(nil, "", rawStderr)
	if !utf8.ValidString(out) {
		t.Fatalf("parseReport returned invalid UTF-8 string: %q", out)
	}
	if out != suffix {
		t.Fatalf("expected truncated stderr %q, got %q", suffix, out)
	}
}
