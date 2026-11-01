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
	wire := buf.String()
	if strings.Contains(wire, "Content-Length") {
		t.Fatalf("grok agent stdio is NDJSON, got framed write: %q", wire)
	}
	if !strings.HasPrefix(wire, "{") || !strings.HasSuffix(wire, "}\n") {
		t.Fatalf("want one JSON line, got %q", wire)
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

func TestReadRPCLeadingEmptyLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\r\n\n{\"id\":4,\"result\":{}}\n"))
	msg, err := readRPC(r)
	if err != nil {
		t.Fatalf("unexpected error on leading empty lines: %v", err)
	}
	if msg["id"] != float64(4) {
		t.Fatalf("got id %v, want 4", msg["id"])
	}
}

func TestReadRPCContentLength(t *testing.T) {
	body := `{"id":3,"result":{"ok":true}}`
	framed := "Content-Length: 29\r\n\r\n" + body
	if len(body) != 29 {
		t.Fatalf("fixture length %d", len(body))
	}
	msg, err := readRPC(bufio.NewReader(strings.NewReader(framed)))
	if err != nil {
		t.Fatal(err)
	}
	if msg["id"] != float64(3) {
		t.Fatalf("id %v", msg["id"])
	}
}
