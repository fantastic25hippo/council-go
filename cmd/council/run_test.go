package main

import (
	"context"
	"errors"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestNewRelay(t *testing.T) {
	fake := &worker.Fake{}
	r := newRelay(fake)
	if r.Council == nil {
		t.Error("expected Council to not be nil")
	}
	if r.Worker != fake {
		t.Error("expected Worker to be the fake worker")
	}
}

func TestRun(t *testing.T) {
	fake := &worker.Fake{
		CancelErr: errors.New("fake cancel error"),
	}
	r := newRelay(fake)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, r)

	if fake.Cancels != 1 {
		t.Errorf("expected 1 cancel, got %d", fake.Cancels)
	}

	if err != fake.CancelErr {
		t.Errorf("expected %v, got %v", fake.CancelErr, err)
	}
}
