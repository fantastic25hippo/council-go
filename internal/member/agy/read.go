package agy

import (
	"bufio"
	"io"
)

func startReader(r io.Reader) chan string {
	msgs := make(chan string, 32)
	go func() {
		defer close(msgs)
		br := bufio.NewReader(r)
		for {
			line, err := br.ReadString('\n')
			if line != "" {
				msgs <- line
			}
			if err != nil {
				return
			}
		}
	}()
	return msgs
}
