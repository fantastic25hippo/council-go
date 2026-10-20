package evidence

import "testing"

func TestPathOwned_DotAndUnrelated(t *testing.T) {
	// Exact call from A4-11
	if !pathOwned("internal/gate/gate.go", []string{"."}) {
		t.Errorf(`pathOwned("internal/gate/gate.go", []string{"."}) = false, want true`)
	}
	if !pathOwned("internal/gate/gate.go", []string{"./"}) {
		t.Errorf(`pathOwned("internal/gate/gate.go", []string{"./"}) = false, want true`)
	}
	if !pathOwned("internal/gate/gate.go", []string{""}) {
		t.Errorf(`pathOwned("internal/gate/gate.go", []string{""}) = false, want true`)
	}
	if pathOwned("internal/gate/gate.go", []string{"unrelated"}) {
		t.Errorf(`pathOwned("internal/gate/gate.go", []string{"unrelated"}) = true, want false`)
	}
	if pathOwned("internal/gate/gate.go", []string{"internal/gat"}) {
		t.Errorf(`pathOwned("internal/gate/gate.go", []string{"internal/gat"}) = true, want false`)
	}
}
