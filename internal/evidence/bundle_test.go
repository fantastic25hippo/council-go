package evidence

import (
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestBundle(t *testing.T) {
	expected := worker.Evidence{
		Diff:         "d",
		VerifyOutput: "out",
		VerifyExit:   1,
	}
	actual := Bundle("d", "out", 1)
	if actual != expected {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}
