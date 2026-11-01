package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func extractJobDir(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "job: ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "job: "))
		}
	}
	return ""
}

func readJobReport(jobDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(jobDir, "report.json"))
	if err != nil {
		return "", fmt.Errorf("reading report.json: %w", err)
	}
	var rep struct {
		Status       string `json:"status"`
		FinalSummary string `json:"final_summary"`
	}
	if err := json.Unmarshal(data, &rep); err != nil {
		return "", fmt.Errorf("parsing report.json: %w", err)
	}
	if rep.Status != "ok" {
		return rep.FinalSummary, fmt.Errorf("job failed with status %q", rep.Status)
	}
	return rep.FinalSummary, nil
}
