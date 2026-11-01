package evidence

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkLocalDiff_UntrackedFiles(b *testing.B) {
	dir := setupGitRepo(&testing.T{})
	ownedDir := filepath.Join(dir, "owned")
	if err := os.Mkdir(ownedDir, 0755); err != nil {
		b.Fatalf("mkdir failed: %v", err)
	}

	const numFiles = 50
	for i := 0; i < numFiles; i++ {
		fn := filepath.Join(ownedDir, fmt.Sprintf("untracked_%d.txt", i))
		content := fmt.Sprintf("sample untracked file content %d\nwith some extra text\n", i)
		if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
			b.Fatalf("write file failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff, err := LocalDiff(dir, []string{"owned"})
		if err != nil {
			b.Fatalf("LocalDiff failed: %v", err)
		}
		if len(diff) == 0 {
			b.Fatalf("expected non-empty diff")
		}
	}
}

func BenchmarkBuildUntrackedDiff(b *testing.B) {
	dir := b.TempDir()
	const numFiles = 50
	var files []string
	for i := 0; i < numFiles; i++ {
		name := fmt.Sprintf("untracked_%d.txt", i)
		files = append(files, name)
		content := fmt.Sprintf("sample untracked file content %d\nwith some extra text\n", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			b.Fatalf("write file failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := buildUntrackedDiff(dir, files)
		if err != nil {
			b.Fatalf("buildUntrackedDiff failed: %v", err)
		}
		if len(out) == 0 {
			b.Fatalf("expected non-empty output")
		}
	}
}
