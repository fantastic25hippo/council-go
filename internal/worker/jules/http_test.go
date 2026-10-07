package jules

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPDo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "k" {
			t.Errorf("missing key header")
		}
		if r.URL.Path != "/sessions/x" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"id":"x"}`)
	}))
	defer srv.Close()
	h := &HTTP{Base: srv.URL, Key: "k", Client: srv.Client()}
	st, body, err := h.Do("GET", "/sessions/x", nil)
	if err != nil || st != 200 || string(body) != `{"id":"x"}` {
		t.Fatalf("%d %s %v", st, body, err)
	}
}

func TestHTTPDoNoKey(t *testing.T) {
	h := NewHTTP("")
	_, _, err := h.Do("GET", "/sessions/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
