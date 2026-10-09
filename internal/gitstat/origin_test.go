package gitstat

import (
	"os"
	"os/exec"
	"testing"
)

func TestOriginRepo(t *testing.T) {
	tests := []struct{
		u string
		want string
	}{
		{"git@github.com:fantastic25hippo/council-go.git", "fantastic25hippo/council-go"},
		{"https://github.com/fantastic25hippo/council-go.git", "fantastic25hippo/council-go"},
		{"https://github.com/fantastic25hippo/council-go", "fantastic25hippo/council-go"},
		{"git@gitlab.com:fantastic25hippo/council-go.git", ""},
	}
	for _, tc := range tests {
		m := originRe.FindStringSubmatch(tc.u)
		got := ""
		if m != nil {
			got = m[1]
		}
		if got != tc.want {
			t.Errorf("got %q want %q for %q", got, tc.want, tc.u)
		}
	}
}

func TestOriginRepoCmd(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitstat")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	exec.Command("git", "init", dir).Run()
	exec.Command("git", "-C", dir, "remote", "add", "origin", "https://github.com/a/b.git").Run()

	got := OriginRepo(dir)
	if got != "a/b" {
		t.Errorf("got %q want %q", got, "a/b")
	}
}
