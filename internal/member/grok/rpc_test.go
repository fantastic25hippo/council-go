package grok

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestWriteReadRPC(t *testing.T) {
	var buf bytes.Buffer
	if err := writeRPC(&buf, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize"}); err != nil {
		t.Fatal(err)
	}
	msg, err := readRPC(bufio.NewReader(&buf))
	if err != nil {
		t.Fatal(err)
	}
	if msg["method"] != "initialize" {
		t.Fatalf("%v", msg)
	}
}

func TestReadRPCBareJSONLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("{\"id\":2,\"result\":{}}\n"))
	msg, err := readRPC(r)
	if err != nil {
		t.Fatal(err)
	}
	if msg["id"] != float64(2) {
		t.Fatalf("id %v", msg["id"])
	}
}
