package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractJobDir(t *testing.T) {
	cases := []struct {
		name   string
		stdout string
		want   string
	}{
		{
			name:   "empty stdout",
			stdout: "",
			want:   "",
		},
		{
			name:   "no job line",
			stdout: "starting worker\nrunning commands\ndone\n",
			want:   "",
		},
		{
			name:   "single job line",
			stdout: "job: /path/to/workdir/job-123\n",
			want:   "/path/to/workdir/job-123",
		},
		{
			name:   "multiline with whitespace padding",
			stdout: "  logs before\n   job:   /custom/job/dir   \nmore logs after\n",
			want:   "/custom/job/dir",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractJobDir(tc.stdout)
			if got != tc.want {
				t.Fatalf("extractJobDir(%q) = %q, want %q", tc.stdout, got, tc.want)
			}
		})
	}
}

func TestReadJobReport(t *testing.T) {
	dir := t.TempDir()

	// 1. Missing report.json
	_, err := readJobReport(dir)
	if err == nil {
		t.Fatalf("expected error for missing report.json, got nil")
	}

	// 2. Malformed JSON
	reportPath := filepath.Join(dir, "report.json")
	if err := os.WriteFile(reportPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = readJobReport(dir)
	if err == nil {
		t.Fatalf("expected error for invalid json, got nil")
	}

	// 3. Status not ok (status: failed)
	failedJSON := `{"status": "failed", "final_summary": "failed details"}`
	if err := os.WriteFile(reportPath, []byte(failedJSON), 0644); err != nil {
		t.Fatal(err)
	}
	summary, err := readJobReport(dir)
	if err == nil {
		t.Fatalf("expected error for non-ok status, got nil")
	}
	if summary != "failed details" {
		t.Fatalf("readJobReport returned summary %q, want %q", summary, "failed details")
	}

	// 4. Status ok
	okJSON := `{"status": "ok", "final_summary": "completed successfully"}`
	if err := os.WriteFile(reportPath, []byte(okJSON), 0644); err != nil {
		t.Fatal(err)
	}
	summary, err = readJobReport(dir)
	if err != nil {
		t.Fatalf("readJobReport returned unexpected error: %v", err)
	}
	if summary != "completed successfully" {
		t.Fatalf("readJobReport returned summary %q, want %q", summary, "completed successfully")
	}
}
