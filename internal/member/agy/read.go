package agy

import (
	"bufio"
	"fmt"
	"time"
)

func readLine(r *bufio.Reader, d time.Duration) (string, error) {
	type res struct {
		s   string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		s, err := r.ReadString('\n')
		ch <- res{s, err}
	}()
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return "", fmt.Errorf("agy read timeout")
	case out := <-ch:
		return out.s, out.err
	}
}
