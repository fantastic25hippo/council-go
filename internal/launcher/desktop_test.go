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

func TestDesktopFiles(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatalf("failed to find module root: %v", err)
	}

	desktopFile := filepath.Join(root, "council.desktop")
	desktopContent, err := os.ReadFile(desktopFile)
	if err != nil {
		t.Fatalf("failed to read council.desktop: %v", err)
	}
	desktopStr := string(desktopContent)

	if desktopStr == "" {
		t.Errorf("council.desktop is empty")
	}
	if strings.Contains(desktopStr, "/home/sahil") {
		t.Errorf("council.desktop contains /home/sahil")
	}
	if !strings.Contains(desktopStr, "Type=Application") {
		t.Errorf("council.desktop missing Type=Application")
	}
	if !strings.Contains(desktopStr, "Name=Council") {
		t.Errorf("council.desktop missing Name=Council")
	}
	if !strings.Contains(desktopStr, "Terminal=false") {
		t.Errorf("council.desktop missing Terminal=false")
	}
	if !strings.Contains(desktopStr, "%k") {
		t.Errorf("council.desktop missing %%k")
	}
	if !strings.Contains(desktopStr, "bin/from-desktop") && !strings.Contains(desktopStr, "bin','from-desktop") {
		t.Errorf("council.desktop missing bin/from-desktop")
	}

	fromDesktopFile := filepath.Join(root, "bin", "from-desktop")
	fromDesktopContent, err := os.ReadFile(fromDesktopFile)
	if err != nil {
		t.Fatalf("failed to read bin/from-desktop: %v", err)
	}
	fromDesktopStr := string(fromDesktopContent)

	if fromDesktopStr == "" {
		t.Errorf("bin/from-desktop is empty")
	}
	if strings.Contains(fromDesktopStr, "/home/sahil") {
		t.Errorf("bin/from-desktop contains /home/sahil")
	}
	if !strings.Contains(fromDesktopStr, "bin/council") && (!strings.Contains(fromDesktopStr, "bin") || !strings.Contains(fromDesktopStr, "council")) {
		t.Errorf("bin/from-desktop missing bin/council")
	}

	readmeFile := filepath.Join(root, "README.md")
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	readmeStr := string(readmeContent)

	if !strings.Contains(readmeStr, "council.desktop") {
		t.Errorf("README.md does not contain council.desktop")
	}
	if !strings.Contains(readmeStr, "Do not copy the launcher to the Desktop") {
		t.Errorf("README.md does not tell user to avoid copying to Desktop")
	}
}

func TestFromDesktopExec(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatalf("failed to find module root: %v", err)
	}

	tempDir := t.TempDir()

	err = os.Mkdir(filepath.Join(tempDir, "bin"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake bin dir: %v", err)
	}

	realDesktop, err := os.ReadFile(filepath.Join(root, "council.desktop"))
	if err != nil {
		t.Fatalf("failed to read real council.desktop: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "council.desktop"), realDesktop, 0755)
	if err != nil {
		t.Fatalf("failed to write fake council.desktop: %v", err)
	}

	realFromDesktop, err := os.ReadFile(filepath.Join(root, "bin", "from-desktop"))
	if err != nil {
		t.Fatalf("failed to read real from-desktop: %v", err)
	}
	fakeFromDesktop := filepath.Join(tempDir, "bin", "from-desktop")
	err = os.WriteFile(fakeFromDesktop, realFromDesktop, 0755)
	if err != nil {
		t.Fatalf("failed to write fake from-desktop: %v", err)
	}

	fakeCouncil := filepath.Join(tempDir, "bin", "council")
	fakeCouncilContent := []byte(`#!/usr/bin/env bash
echo PWD=$(pwd)
echo ARGV=$@
exit 0
`)
	err = os.WriteFile(fakeCouncil, fakeCouncilContent, 0755)
	if err != nil {
		t.Fatalf("failed to write fake council: %v", err)
	}

	runTest := func(name string, args []string, expectError bool, expectOutput string) {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("python3", args...)
			cmd.Env = os.Environ()
			// Unset environment variables as requested
			for i, env := range cmd.Env {
				if strings.HasPrefix(env, "DISPLAY=") || strings.HasPrefix(env, "WAYLAND_DISPLAY=") || strings.HasPrefix(env, "COUNCIL_START_DIR=") {
					cmd.Env[i] = "CLEARED_ENV_VAR=1"
				}
			}

			// Remove the ones that need unsetting
			var cleanEnv []string
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "DISPLAY=") && !strings.HasPrefix(e, "WAYLAND_DISPLAY=") && !strings.HasPrefix(e, "COUNCIL_START_DIR=") {
					cleanEnv = append(cleanEnv, e)
				}
			}
			cmd.Env = cleanEnv

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !expectError && err != nil {
				t.Errorf("unexpected error: %v, stderr: %s", err, stderr.String())
			}

			if expectOutput != "" && !strings.Contains(stdout.String(), expectOutput) {
				t.Errorf("expected output to contain %q, got: %q", expectOutput, stdout.String())
			}
		})
	}

	// 1. standard launch
	runTest("standard_path", []string{fakeFromDesktop, filepath.Join(tempDir, "council.desktop"), "extra"}, false, "ARGV=extra")

	// 2. file:// launch
	absDesktop, _ := filepath.Abs(filepath.Join(tempDir, "council.desktop"))
	fileURL := "file://" + absDesktop
	runTest("file_url", []string{fakeFromDesktop, fileURL, "extra"}, false, "ARGV=extra")

	// 3. missing arg -> exits 1
	runTest("missing_arg", []string{fakeFromDesktop}, true, "")
}
