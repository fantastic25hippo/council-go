package evidence
import ( "os"; "os/exec"; "path/filepath"; "strings"; "testing" )
func TestLocalDiff_StagedChange(t *testing.T) {
	dir := setupGitRepo(t)
	_ = os.Mkdir(filepath.Join(dir, "owned"), 0755)
	p := filepath.Join(dir, "owned", "file.txt")
	_ = os.WriteFile(p, []byte("initial\n"), 0644)
	_ = exec.Command("git", "-C", dir, "add", "owned/file.txt").Run()
	_ = exec.Command("git", "-C", dir, "commit", "-m", "i").Run()
	_ = os.WriteFile(p, []byte("staged-new\n"), 0644)
	_ = exec.Command("git", "-C", dir, "add", "owned/file.txt").Run()
	diff, err := LocalDiff(dir, []string{"owned"})
	if err != nil { t.Fatal(err) }
	if !strings.Contains(diff, "staged-new") { t.Fatalf("missing staged-new: %q", diff) }
}
