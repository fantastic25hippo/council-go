package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fantastic25hippo/council-go/internal/worker/local"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: council <command>\n")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGTERM, os.Interrupt)
	defer stop()

	r := newRelay(local.New(nil))
	if err := run(ctx, r); err != nil {
		os.Exit(1)
	}
}
