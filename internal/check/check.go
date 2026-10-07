package check

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const MaxGoFileLines = 200

// ModuleRoot walks up from start until go.mod is found.
func ModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// GoLineCounts maps path-relative-to-root → line count for every *.go file
// under root, skipping the vendor/ directory if present.
func GoLineCounts(root string) (map[string]int, error) {
	counts := make(map[string]int)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == "vendor" {
				return fs.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		lines := bytes.Count(content, []byte("\n"))
		if len(content) > 0 && content[len(content)-1] != '\n' {
			lines++
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		counts[relPath] = lines
		return nil
	})

	if err != nil {
		return nil, err
	}

	return counts, nil
}
