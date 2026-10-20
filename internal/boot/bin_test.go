package boot_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/check"
)

func TestCouncilTTYSkipsGUIPicker(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}

	councilPath := filepath.Join(root, "bin/council")
	councilBytes, err := os.ReadFile(councilPath)
	if err != nil {
		t.Fatalf("failed to read bin/council: %v", err)
	}
	councilStr := string(councilBytes)

	// Minimal source guard check
	if !strings.Contains(councilStr, "[ ! -t 0 ] && [ ! -p /dev/stdin ] && have_gui_picker") {
		t.Fatal("bin/council missing [ ! -t 0 ] guard before have_gui_picker")
	}

	// Dynamic test using script -qc to provide a TTY
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("script utility not available; minimal source check passed")
	}

	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "fake_bin")
	if err := os.MkdirAll(fakeBin, 0755); err != nil {
		t.Fatal(err)
	}

	zenityPath := filepath.Join(fakeBin, "zenity")
	if err := os.WriteFile(zenityPath, []byte("#!/bin/sh\necho ZENITY_CALLED\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	gnomeTermPath := filepath.Join(fakeBin, "gnome-terminal")
	if err := os.WriteFile(gnomeTermPath, []byte("#!/bin/sh\necho GNOME_CALLED\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("script", "-qc", councilPath, "/dev/null")
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+":"+os.Getenv("PATH"), "DISPLAY=:0", "COUNCIL_START_DIR=")
	cmd.Stdin = strings.NewReader("q\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	_ = cmd.Run()

	output := out.String()
	if strings.Contains(output, "ZENITY_CALLED") {
		t.Fatalf("expected CLI prompt on TTY, but zenity was called: %s", output)
	}
	if !strings.Contains(output, "Enter folder path for session") {
		t.Fatalf("expected CLI prompt in output, got: %s", output)
	}
}

func TestCouncilHereTTYRunsInPlace(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}

	councilHerePath := filepath.Join(root, "bin/council-here")
	councilHereBytes, err := os.ReadFile(councilHerePath)
	if err != nil {
		t.Fatalf("failed to read bin/council-here: %v", err)
	}
	councilHereStr := string(councilHereBytes)

	// Minimal source guard check
	if !strings.Contains(councilHereStr, "[ -t 0 ]") {
		t.Fatal("bin/council-here missing [ -t 0 ] guard")
	}

	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("script utility not available; minimal source check passed")
	}

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "workdir")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	stubBin := filepath.Join(tmpDir, "stub_bin")
	if err := os.MkdirAll(stubBin, 0755); err != nil {
		t.Fatal(err)
	}
	for _, tool := range []string{"bash", "sh", "dirname", "readlink", "env", "script"} {
		p, err := exec.LookPath(tool)
		if err == nil {
			_ = os.Symlink(p, filepath.Join(stubBin, tool))
		}
	}

	cmd := exec.Command("script", "-qc", councilHerePath+" "+workDir, "/dev/null")
	cmd.Env = []string{
		"PATH=" + stubBin,
		"COUNCIL_TMUX_SESSION=",
		"DISPLAY=",
		"WAYLAND_DISPLAY=",
		"HOME=" + tmpDir,
	}
	cmd.Stdin = strings.NewReader("quit\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	_ = cmd.Run()

	output := out.String()
	if strings.Contains(output, "error: no terminal was found") {
		t.Fatalf("expected council-here to run in place on TTY, but got: %s", output)
	}
}
