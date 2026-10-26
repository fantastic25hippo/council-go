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
	// Check variable assignments only. A host may inject other variables whose
	// value merely mentions a key name, which is not a leak.
	lines := strings.Split(out, "\n")
	for _, name := range []string{"XAI_API_KEY", "GOOGLE_API_KEY", "GEMINI_API_KEY"} {
		for _, line := range lines {
			if strings.HasPrefix(line, name+"=") {
				t.Errorf("output still sets %s", name)
			}
		}
	}
	for _, v := range []string{"secret_xai", "secret_google", "secret_gemini"} {
		if strings.Contains(out, v) {
			t.Errorf("output contains the value of a forbidden key (%s)", v)
		}
	}
	if !strings.Contains(out, "ALLOWED_VAR=allowed_value") {
		t.Errorf("output missing ALLOWED_VAR: %s", out)
	}
}
