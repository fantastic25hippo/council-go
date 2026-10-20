package evidence

import "testing"

func TestPRChecks_Exit1_NoChecksStderr(t *testing.T) {
	fake := &fakeGHRunner{
		stdout: "",
		stderr: "no checks reported on commit\n",
		exit:   1,
	}
	chk, err := PRChecks(fake, "42", "")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if chk.Failed {
		t.Errorf("expected Failed false when no checks reported on stderr, got %+v", chk)
	}
	if chk.Out != "no checks reported on commit\n" {
		t.Errorf("expected Out to contain stderr, got %q", chk.Out)
	}
}
