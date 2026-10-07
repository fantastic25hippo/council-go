package grok

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func writeRPC(w io.Writer, msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(body), body)
	return err
}

func readRPC(r *bufio.Reader) (map[string]any, error) {
	var n int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trim), "content-length:") {
			_, rest, _ := strings.Cut(trim, ":")
			n, err = strconv.Atoi(strings.TrimSpace(rest))
			if err != nil {
				return nil, err
			}
		}
		if trim == "" {
			break
		}
		if strings.HasPrefix(trim, "{") {
			var m map[string]any
			if err := json.Unmarshal([]byte(trim), &m); err != nil {
				return nil, err
			}
			return m, nil
		}
	}
	if n <= 0 {
		return nil, fmt.Errorf("no Content-Length")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, err
	}
	return m, nil
}
