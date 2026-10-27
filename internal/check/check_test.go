package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaxGoFileLines(t *testing.T) {
	if MaxGoFileLines != 200 {
		t.Errorf("MaxGoFileLines = %d, want 200", MaxGoFileLines)
	}
}

func TestGoLineCounts(t *testing.T) {
	tmpDir := t.TempDir()

	// 3 lines
	err := os.WriteFile(filepath.Join(tmpDir, "3lines.go"), []byte("line 1\nline 2\nline 3"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 201 lines
	lines201 := strings.Repeat("line\n", 200) + "line"
	err = os.WriteFile(filepath.Join(tmpDir, "201lines.go"), []byte(lines201), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Not a .go file
	err = os.WriteFile(filepath.Join(tmpDir, "notgo.txt"), []byte("txt"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Vendor directory (should be skipped)
	vendorDir := filepath.Join(tmpDir, "vendor")
	err = os.Mkdir(vendorDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(vendorDir, "skipped.go"), []byte("skipped\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	counts, err := GoLineCounts(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if counts["3lines.go"] != 3 {
		t.Errorf("3lines.go count = %d, want 3", counts["3lines.go"])
	}

	if counts["201lines.go"] != 201 {
		t.Errorf("201lines.go count = %d, want 201", counts["201lines.go"])
	}

	if _, ok := counts["notgo.txt"]; ok {
		t.Errorf("notgo.txt should not be counted")
	}

	if _, ok := counts[filepath.Join("vendor", "skipped.go")]; ok {
		t.Errorf("vendor/skipped.go should not be counted")
	}
}

func TestGoFilesAtMost200(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root, err := ModuleRoot(wd)
	if err != nil {
		t.Fatal(err)
	}

	counts, err := GoLineCounts(root)
	if err != nil {
		t.Fatal(err)
	}

	for path, lines := range counts {
		if lines > MaxGoFileLines {
			t.Errorf("file %s has %d lines, exceeding limit of %d", path, lines, MaxGoFileLines)
		}
	}
}

func TestReadmeLinksHowToUse(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := ModuleRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "docs/HOW_TO_USE_COUNCIL-GO.md") {
		t.Fatal("README.md must link docs/HOW_TO_USE_COUNCIL-GO.md")
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "HOW_TO_USE_COUNCIL-GO.md")); err != nil {
		t.Fatal(err)
	}
}

func TestNoDangerouslySkipPermissionsOutsideMember(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := ModuleRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	const flag = "--dangerously-skip-permissions"
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "member" {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), flag) {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s has %s; use grok-agy-run job as the fix", rel, flag)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
