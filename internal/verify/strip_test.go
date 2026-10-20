package verify

import (
	"strings"
	"testing"
)

func TestRun_StripsForbiddenKeys(t *testing.T) {
	t.Setenv("XAI_API_KEY", "secret_xai")
	t.Setenv("GOOGLE_API_KEY", "secret_google")
	t.Setenv("GEMINI_API_KEY", "secret_gemini")
	t.Setenv("ALLOWED_VAR", "allowed_value")

	out, exit, err := Run(t.TempDir(), "env", 10)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if exit != 0 {
		t.Fatalf("Run exit = %d, want 0; out: %s", exit, out)
	}
	if strings.Contains(out, "secret_xai") || strings.Contains(out, "XAI_API_KEY") {
		t.Errorf("output contains XAI_API_KEY: %s", out)
	}
	if strings.Contains(out, "secret_google") || strings.Contains(out, "GOOGLE_API_KEY") {
		t.Errorf("output contains GOOGLE_API_KEY: %s", out)
	}
	if strings.Contains(out, "secret_gemini") || strings.Contains(out, "GEMINI_API_KEY") {
		t.Errorf("output contains GEMINI_API_KEY: %s", out)
	}
	if !strings.Contains(out, "ALLOWED_VAR=allowed_value") {
		t.Errorf("output missing ALLOWED_VAR: %s", out)
	}
}
