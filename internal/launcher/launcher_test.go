package launcher_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/check"
)

func TestLauncherScripts(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}

	scripts := []string{"bin/council", "bin/council-here"}
	for _, script := range scripts {
		path := filepath.Join(root, script)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", script, err)
		}
		if len(b) == 0 {
			t.Errorf("%s is empty", script)
		}

		if strings.Contains(string(b), "/home/sahil") {
			t.Errorf("%s contains /home/sahil", script)
		}
	}

	councilPath := filepath.Join(root, "bin/council")
	councilBytes, err := os.ReadFile(councilPath)
	if err != nil {
		t.Fatalf("failed to read bin/council: %v", err)
	}
	councilContent := string(councilBytes)

	requiredStrings := []string{
		"COUNCIL_START_DIR",
		"zenity",
		"gnome-terminal",
		"council-bin",
		"~/.grok/bin",
		"~/.local/bin",
		"jules.env",
	}
	for _, s := range requiredStrings {
		if !strings.Contains(councilContent, s) {
			t.Errorf("bin/council does not contain %q", s)
		}
	}
}

func TestReadmeContents(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	readme := string(b)

	if !strings.Contains(readme, "bin/council-bin") {
		t.Errorf("README does not contain bin/council-bin")
	}
	if !strings.Contains(readme, "./bin/council") {
		t.Errorf("README does not contain ./bin/council")
	}
	if strings.Contains(readme, "go build -o bin/council ./cmd/council") {
		t.Errorf("README tells user to go build -o bin/council")
	}
}

func TestLauncherExec(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()

	// Setup fake repo
	fakeRepo := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(filepath.Join(fakeRepo, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(fakeRepo, "cmd", "council"), 0755); err != nil {
		t.Fatal(err)
	}

	// Copy launcher
	srcLauncher := filepath.Join(root, "bin", "council")
	dstLauncher := filepath.Join(fakeRepo, "bin", "council")
	b, err := os.ReadFile(srcLauncher)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstLauncher, b, 0755); err != nil {
		t.Fatal(err)
	}

	// Fake bin/council-bin
	fakeBin := filepath.Join(fakeRepo, "bin", "council-bin")
	fakeBinContent := "#!/usr/bin/env bash\necho PWD=$(pwd)\necho ARGV=$@\nexit 0\n"
	if err := os.WriteFile(fakeBin, []byte(fakeBinContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Unset environment that triggers GUI or interactive logic
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("COUNCIL_START_DIR", "")

	// Create a separate work directory to run the launcher from
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(dstLauncher, "dummy_arg")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "DISPLAY=", "WAYLAND_DISPLAY=", "COUNCIL_START_DIR=")
	cmd.Stdin = nil

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("launcher failed: %v", err)
	}

	out := stdout.String()

	// The TTY-less fallback should use the current working directory.
	// Since we ran it with cmd.Dir = workDir, it should have printed workDir.
	expectedPWD := "PWD=" + workDir
	if !strings.Contains(out, expectedPWD) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedPWD, out)
	}

	if !strings.Contains(out, "ARGV=dummy_arg") {
		t.Errorf("expected output to contain ARGV=dummy_arg, got:\n%s", out)
	}
}
