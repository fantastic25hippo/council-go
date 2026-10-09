package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fantastic25hippo/council-go/internal/boot"
	"github.com/fantastic25hippo/council-go/internal/pause"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: council\nStarts the council meeting in the current directory.\n")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			usage()
			return
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGTERM)
	defer stop()

	pc := pause.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			res := pc.OnSIGINT()
			switch res {
			case "quit":
				stop()
			case "pause":
				fmt.Fprint(os.Stdout, "\n^C paused\n")
			case "Already paused...":
				fmt.Fprint(os.Stdout, "\nAlready paused...\n")
			}
		}
	}()

	if err := boot.Run(ctx, os.Stdin, os.Stdout, cwd, pc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
